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
	"os"
	"path"
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
	return c.loadByConfigName(env, config)
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
	return c.viper.Unmarshal(config)
}
