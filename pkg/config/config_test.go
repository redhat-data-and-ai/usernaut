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
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	AppEnv   string
	Cache    TestCache
	Backends Backends
}

type Backends struct {
	Snowflake map[string]Snowflake
	Fivetran  map[string]Fivetran
}

type Snowflake struct {
	URL     string
	Account string
}

type Fivetran struct {
	ApiKey    string
	ApiSecret string
}

type TestCache struct {
	// Driver is the type of cache client
	Driver string
	Redis  TestRedisConfig
}

type TestRedisConfig struct {
	Host              string
	Port              string
	Database          int32
	Password          string
	DefaultExpiration int32
	CleanupInterval   int32
}

func TestLoadConfigFromYAML(t *testing.T) {
	var c TestConfig

	key := "CACHE_REDIS_PASSWORD"

	_ = os.Setenv(key, "redispassword")

	err := NewConfig(NewOptions("yaml", "./testdata", "default")).Load("dev", &c)
	assert.Nil(t, err)

	assertConfigs(t, &c)

	_ = os.Unsetenv(key)
}

func TestLoadConfigFromTOML(t *testing.T) {
	var c TestConfig

	key := "CACHE_REDIS_PASSWORD"

	_ = os.Setenv(key, "redispassword")

	err := NewConfig(NewOptions("toml", "./testdata", "defaulttoml")).Load("devtoml", &c)
	assert.Nil(t, err)

	assertConfigs(t, &c)

	_ = os.Unsetenv(key)
}

func assertConfigs(t *testing.T, c *TestConfig) {
	// assert that app environment got overridden with dev.toml
	assert.Equal(t, "dev", c.AppEnv)

	// asserts that cache driver is redis
	assert.Equal(t, "redis", c.Cache.Driver)
	// asserts that redis properties are being fetched from toml file
	assert.Equal(t, int32(5), c.Cache.Redis.Database)
	assert.Equal(t, "localhost", c.Cache.Redis.Host)

	// assert that redis password set via environment variable is fetched accurately
	assert.Equal(t, "redispassword", c.Cache.Redis.Password)
}
