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
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFileSubstitution(t *testing.T) {
	// Create a temp file with a secret value
	secretValue := "supersecretfilevalue"
	tmpfile, err := os.CreateTemp("", "secretfile-*.txt")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()
	_, err = tmpfile.Write([]byte(secretValue))
	assert.Nil(t, err)
	_ = tmpfile.Close()

	type FileConfig struct {
		Secret string
	}

	fileConfig := FileConfig{}
	// Simulate what would be loaded from YAML
	fileConfig.Secret = "file|" + tmpfile.Name()

	// Substitute
	SubstituteConfigValues(reflect.ValueOf(&fileConfig))

	assert.Equal(t, secretValue, fileConfig.Secret)
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

func TestLoadConfigFromPathHTTPURL(t *testing.T) {
	type TestConfig struct {
		Value string `yaml:"value"`
	}

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("value: test-config-value\n"))
	}))
	defer server.Close()

	var config TestConfig
	err := LoadConfigFromPath(context.Background(), server.URL+"/config.yaml", &config)
	require.NoError(t, err)
	assert.Equal(t, "test-config-value", config.Value)
}

func TestLoadConfigFromPathInvalidURL(t *testing.T) {
	type TestConfig struct {
		Value string
	}

	var config TestConfig

	// Test invalid URL format (malformed HTTP URL)
	err := LoadConfigFromPath(context.Background(), "http://[invalid-url", &config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL format")

	// Test unsupported scheme (file:// is not allowed, only http/https)
	err = LoadConfigFromPath(context.Background(), "file:///etc/config.yaml", &config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestLoadConfigFromPathHTTPError(t *testing.T) {
	type TestConfig struct {
		Value string
	}

	// Create a test HTTP server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	var config TestConfig
	err := LoadConfigFromPath(context.Background(), server.URL+"/config.yaml", &config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
	assert.Contains(t, err.Error(), "Not Found")
}
