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

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
)

// Config represents the top-level configuration structure
type AppConfig struct {
	App        App                       `yaml:"app"`
	LDAP       ldap.LDAP                 `yaml:"ldap"`
	Cache      cache.Config              `yaml:"cache"`
	Backends   []Backend                 `yaml:"backends"`
	Pattern    map[string][]PatternEntry `yaml:"pattern"`
	HttpClient struct {
		ConnectionPoolConfig    httpclient.ConnectionPoolConfig    `yaml:"connectionPoolConfig"`
		HystrixResiliencyConfig httpclient.HystrixResiliencyConfig `yaml:"hystrixResiliencyConfig"`
	} `yaml:"httpClient"`
	APIServer  APIServerConfig               `yaml:"apiServer"`
	BackendMap map[string]map[string]Backend `yaml:"-"`
}

type APIServerConfig struct {
	Host string     `yaml:"host"`
	Port int        `yaml:"port"`
	Auth AuthConfig `yaml:"auth"`
	CORS CORSConfig `yaml:"cors"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

type AuthConfig struct {
	Enabled bool `yaml:"enabled"`
}

// PatternEntry represents the input and output pattern of group names
type PatternEntry struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

// App represents the application configuration
type App struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// Backend represents a backend service configuration
type Backend struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Connection map[string]interface{} `yaml:"connection"`
}

func (b *Backend) GetStringConnection(name string, defaultValue string) string {
	if val, ok := b.Connection[name].(string); ok {
		return val
	}
	return defaultValue
}

var config *AppConfig

func LoadConfig(env string) (*AppConfig, error) {
	// Init config
	config = &AppConfig{}
	err := NewDefaultConfig().Load(env, config)
	if err != nil {
		return nil, err
	}

	// convert backends to a map for easier access
	config.BackendMap = make(map[string]map[string]Backend)
	for _, backend := range config.Backends {
		if config.BackendMap[backend.Type] == nil {
			config.BackendMap[backend.Type] = make(map[string]Backend)
		}
		config.BackendMap[backend.Type][backend.Name] = backend
	}

	return config, nil
}

func getOrDefaultEnv() string {
	env := os.Getenv("APP_ENV")
	if len(env) == 0 {
		return "default"
	}
	return env
}

func GetConfig() (*AppConfig, error) {
	var err error
	if config == nil {
		config, err = LoadConfig(getOrDefaultEnv())
	}

	return config, err
}
