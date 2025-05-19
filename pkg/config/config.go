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

type AppConfig struct {
	Application AppInfo
	Cache       cache.Config
	Backends    clients.Backends
}

type AppInfo struct {
	Name        string
	Environment string
}

var config *AppConfig

func loadConfig(env string) (*AppConfig, error) {
	// Init config
	config = &AppConfig{}
	err := NewDefaultConfig().Load(env, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetConfig() (*AppConfig, error) {
	var err error
	if config == nil {
		config, err = loadConfig(getOrDefaultEnv())
	}
	return config, err
}

func getOrDefaultEnv() string {
	env := os.Getenv("APP_ENV")
	if len(env) == 0 {
		return "default"
	}
	return env
}
