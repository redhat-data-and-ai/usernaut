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
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"

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

	if appConfig, ok := config.(*AppConfig); ok {
		c.APIKeyloading(appConfig)
	}

	return nil
}

func (c *Config) APIKeyloading(appConfig *AppConfig) {
	if appConfig.APIServer.Enabled && len(appConfig.APIServer.Auth.APIKeys) == 0 {
		if keys := c.viper.GetStringSlice("apiServer.auth.api_keys"); len(keys) > 0 {
			appConfig.APIServer.Auth.APIKeys = keys
		}
	}
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
