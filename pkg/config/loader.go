/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
)

// Default options for configuration loading.
const (
	DefaultConfigType     = "yaml"
	DefaultConfigDir      = "./appconfig"
	DefaultConfigFileName = "default"
	WorkDirEnv            = "WORKDIR"
	EnvPrefix             = "env|"
	FilePrefix            = "file|"
	maxConfigSize         = 2 * 1024 * 1024 // 2MB
	configHTTPTimeout     = 30 * time.Second
	configUserAgent       = "usernaut-config-loader/1.0"
)

var (
	// configHTTPClient is a shared HTTP client for loading configs from URLs
	// It's initialized once and reused for all HTTP config requests
	configHTTPClient     *http.Client
	configHTTPClientOnce sync.Once
)

// Options is config options.
type Options struct {
	configType            string
	configPath            string
	defaultConfigFileName string
}

// Config is a wrapper over a underlying config loader implementation.
type Config struct {
	opts  Options
	viper *viper.Viper
}

func NewDefaultOptions() Options {
	var configPath string
	workDir := os.Getenv(WorkDirEnv)
	if workDir != "" {
		configPath = path.Join(workDir, DefaultConfigDir)
	} else {
		_, thisFile, _, _ := runtime.Caller(1)
		configPath = path.Join(path.Dir(thisFile), "../../"+DefaultConfigDir)
	}
	return NewOptions(DefaultConfigType, configPath, DefaultConfigFileName)
}

// NewOptions returns new Options struct.
func NewOptions(configType string, configPath string, defaultConfigFileName string) Options {
	return Options{configType, configPath, defaultConfigFileName}
}

// NewDefaultConfig returns new config struct with default options.
func NewDefaultConfig() *Config {
	return NewConfig(NewDefaultOptions())
}

// NewConfig returns new config struct.
func NewConfig(opts Options) *Config {
	return &Config{opts, viper.New()}
}

// Load reads environment specific configurations and along with the defaults
// unmarshalls into config.
func (c *Config) Load(env string, config interface{}) error {
	if err := c.loadByConfigName(c.opts.defaultConfigFileName, config); err != nil {
		return err
	}
	if err := c.loadByConfigName(env, config); err != nil {
		return err
	}
	SubstituteConfigValues(reflect.ValueOf(config))
	return nil
}

// SubstituteConfigValues recursively walks through the config struct and replaces
// string values of the form 'env|VAR' or 'file|/path' with the corresponding value.
func SubstituteConfigValues(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	// If it's a pointer, resolve it
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		SubstituteConfigValues(v.Elem())
		return
	}
	// If it's a struct, process its fields
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() || field.Kind() == reflect.Ptr ||
				field.Kind() == reflect.Struct || field.Kind() == reflect.Map ||
				field.Kind() == reflect.Slice {
				SubstituteConfigValues(field)
			}
		}
		return
	}
	// If it's a map, process its values
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			if val.Kind() == reflect.Interface && !val.IsNil() {
				val = val.Elem()
			}
			// Only settable if map value is addressable, so we replace by setting
			if val.Kind() == reflect.String {
				newVal := reflect.ValueOf(substituteString(val.String()))
				v.SetMapIndex(key, newVal)
			} else {
				// Recursively process nested maps/structs
				copyVal := reflect.New(val.Type()).Elem()
				copyVal.Set(val)
				SubstituteConfigValues(copyVal)
				v.SetMapIndex(key, copyVal)
			}
		}
		return
	}
	// If it's a slice or array, process its elements
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			SubstituteConfigValues(v.Index(i))
		}
		return
	}
	// If it's a string, substitute if needed
	if v.Kind() == reflect.String && v.CanSet() {
		v.SetString(substituteString(v.String()))
	}
}

// substituteString replaces 'env|VAR' and 'file|/path' patterns with their values
func substituteString(s string) string {
	if len(s) > len(EnvPrefix) && s[:len(EnvPrefix)] == EnvPrefix {
		return os.Getenv(s[len(EnvPrefix):])
	}
	if len(s) > len(FilePrefix) && s[:len(FilePrefix)] == FilePrefix {
		b, err := os.ReadFile(s[len(FilePrefix):])
		if err != nil {
			panic(fmt.Sprintf("ERROR: %v", err))
		}
		return strings.TrimSpace(string(b))
	}
	return s
}

// loadByConfigName reads configuration from file and unmarshalls into config.
func (c *Config) loadByConfigName(configName string, config interface{}) error {
	c.viper.SetConfigName(configName)
	c.viper.SetConfigType(c.opts.configType)
	c.viper.AddConfigPath(c.opts.configPath)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}
	if err := c.viper.Unmarshal(config); err != nil {
		return err
	}
	return nil
}

// getConfigHTTPClient returns a shared HTTP client for config loading.
// The client is initialized once and reused for all requests.
func getConfigHTTPClient() *http.Client {
	configHTTPClientOnce.Do(func() {
		configHTTPClient = &http.Client{
			Timeout: configHTTPTimeout,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		}
	})
	return configHTTPClient
}

// LoadConfigFromPath reads a configuration from a file path or HTTP URL and unmarshalls into config.
// The path can be:
//   - A local file path (relative to WORKDIR/appconfig or absolute)
//   - An HTTP/HTTPS URL
//
// For local file paths, it uses the same config path resolution as the main config loader.
// For HTTP URLs, it fetches the content via HTTP GET request with context support.
func LoadConfigFromPath(ctx context.Context, configPath string, config interface{}) error {
	// Check if it's an HTTP URL and validate it
	if strings.HasPrefix(configPath, "http://") || strings.HasPrefix(configPath, "https://") {
		return loadConfigFromURL(ctx, configPath, config)
	}

	// Reject other URL schemes (like file://, ftp://, etc.)
	if strings.Contains(configPath, "://") {
		parsedURL, err := url.Parse(configPath)
		if err == nil && parsedURL.Scheme != "" {
			return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", parsedURL.Scheme)
		}
	}

	// Handle local file path
	resolvedPath, err := resolveLocalConfigPath(configPath)
	if err != nil {
		return err
	}

	// Read and unmarshal the file
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to unmarshal config from %s: %w", resolvedPath, err)
	}

	return nil
}

// resolveLocalConfigPath resolves a local file path, handling relative paths and extensions.
func resolveLocalConfigPath(configPath string) (string, error) {
	// If it's already absolute, return as-is
	if filepath.IsAbs(configPath) {
		return configPath, nil
	}

	// Add .yaml extension if not present
	if !strings.Contains(filepath.Base(configPath), ".") {
		configPath = configPath + "." + DefaultConfigType
	}

	// Try to find in config directory using same resolution as NewDefaultOptions
	opts := NewDefaultOptions()
	configDir := opts.configPath

	fullPath := filepath.Join(configDir, configPath)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// If not found in config dir, try as relative to current dir
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	return "", fmt.Errorf("config file not found: %s", configPath)
}

// loadConfigFromURL fetches configuration from an HTTP/HTTPS URL and unmarshalls it.
// It validates the URL, uses a shared HTTP client, and includes proper error handling.
func loadConfigFromURL(ctx context.Context, configURL string, config interface{}) error {
	// Validate URL format and scheme
	parsedURL, err := url.Parse(configURL)
	if err != nil {
		return fmt.Errorf("invalid URL format %q: %w", configURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", parsedURL.Scheme)
	}

	// Create request with context for cancellation/timeout support
	req, err := http.NewRequestWithContext(ctx, "GET", configURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s: %w", configURL, err)
	}

	// Set User-Agent header to identify the client
	req.Header.Set("User-Agent", configUserAgent)

	// Use shared HTTP client
	client := getConfigHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch config from %s: %w", configURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("warning: failed to close response body from %s: %v", configURL, err)
		}
	}()

	// Check status code and include response body in error for debugging
	if resp.StatusCode != http.StatusOK {
		// Read error response body (limited) for better error messages
		errorBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxConfigSize)) // Limit error body to 1KB
		errorMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if readErr == nil && len(errorBody) > 0 {
			errorMsg += fmt.Sprintf(" - %s", string(errorBody))
		}
		return fmt.Errorf("failed to fetch config from %s: %s", configURL, errorMsg)
	}

	// Read response body with size limit
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxConfigSize))
	if err != nil {
		return fmt.Errorf("failed to read response body from %s: %w", configURL, err)
	}

	// Check if we hit the size limit (io.LimitReader doesn't return error on limit)
	if len(data) >= maxConfigSize {
		return fmt.Errorf("config file from %s exceeds maximum size of %d bytes", configURL, maxConfigSize)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to unmarshal config from %s: %w", configURL, err)
	}

	return nil
}
