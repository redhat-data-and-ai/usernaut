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
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
)

// Config represents the top-level configuration structure
type AppConfig struct {
	App        App                                   `yaml:"app"`
	Cache      cache.Config                          `yaml:"cache"`
	Backends   []clients.Backend                     `yaml:"backends"`
	BackendMap map[string]map[string]clients.Backend `yaml:"-"`
}

// App represents the application configuration
type App struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
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
	config.BackendMap = make(map[string]map[string]clients.Backend)
	for _, backend := range config.Backends {
		if config.BackendMap[backend.Type] == nil {
			config.BackendMap[backend.Type] = make(map[string]clients.Backend)
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
