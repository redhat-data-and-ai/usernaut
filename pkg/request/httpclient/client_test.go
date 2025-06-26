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

package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeClient(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		// Create a basic configuration
		connConfig := ConnectionPoolConfig{
			Timeout:            1000,
			KeepAliveTimeout:   5000,
			MaxIdleConnections: 10,
		}

		hystrixConfig := HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    20,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     50,
			CircuitBreakerTimeout:     1000,
		}

		// Initialize client with basic configuration
		client, err := InitializeClient(
			"TestCommand",
			connConfig,
			hystrixConfig,
			nil,
			0,
			nil,
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with custom retrier", func(t *testing.T) {
		// Create a basic configuration
		connConfig := ConnectionPoolConfig{
			Timeout:            1000,
			KeepAliveTimeout:   5000,
			MaxIdleConnections: 10,
		}

		hystrixConfig := HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    20,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     50,
			CircuitBreakerTimeout:     1000,
		}

		// Create a custom retrier
		customRetrier := heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 3*time.Second))

		// Initialize client with custom retrier
		client, err := InitializeClient(
			"TestCommand",
			connConfig,
			hystrixConfig,
			customRetrier,
			3,
			nil,
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with fallback function", func(t *testing.T) {
		// Create a basic configuration
		connConfig := ConnectionPoolConfig{
			Timeout:            1000,
			KeepAliveTimeout:   5000,
			MaxIdleConnections: 10,
		}

		hystrixConfig := HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    20,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     50,
			CircuitBreakerTimeout:     1000,
		}

		// Create a fallback function
		fallbackFunc := func(err error) error {
			return nil
		}

		// Initialize client with fallback function
		client, err := InitializeClient(
			"TestCommand",
			connConfig,
			hystrixConfig,
			nil,
			0,
			fallbackFunc,
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// We cannot test the fallback function directly as it's used internally
		// by the hystrix library, and would require actual circuit breaking
	})

	t.Run("with empty certificate paths", func(t *testing.T) {
		// Create configuration with empty certificate paths
		connConfig := ConnectionPoolConfig{
			Timeout:            1000,
			KeepAliveTimeout:   5000,
			MaxIdleConnections: 10,
			CertPath:           "",
			PrivateKeyPath:     "",
		}

		hystrixConfig := HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    20,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     50,
			CircuitBreakerTimeout:     1000,
		}

		// Initialize client with empty certificate paths
		client, err := InitializeClient(
			"TestCommand",
			connConfig,
			hystrixConfig,
			nil,
			0,
			nil,
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with invalid certificate paths", func(t *testing.T) {
		// Create configuration with invalid certificate paths
		connConfig := ConnectionPoolConfig{
			Timeout:            1000,
			KeepAliveTimeout:   5000,
			MaxIdleConnections: 10,
			CertPath:           "/non/existent/path/cert.pem",
			PrivateKeyPath:     "/non/existent/path/key.pem",
		}

		hystrixConfig := HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    20,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     50,
			CircuitBreakerTimeout:     1000,
		}

		// Initialize client with invalid certificate paths
		client, err := InitializeClient(
			"TestCommand",
			connConfig,
			hystrixConfig,
			nil,
			0,
			nil,
		)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to load certificate and key")
	})
}

// Integration test to verify client can make actual HTTP requests
func TestClientIntegration(t *testing.T) {
	// Skip this test when running in CI environments
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Start a test HTTP server using httptest
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer testServer.Close()

	// Create client configuration
	connConfig := ConnectionPoolConfig{
		Timeout:            1000,
		KeepAliveTimeout:   5000,
		MaxIdleConnections: 10,
	}

	hystrixConfig := HystrixResiliencyConfig{
		MaxConcurrentRequests:     100,
		RequestVolumeThreshold:    20,
		CircuitBreakerSleepWindow: 5000,
		ErrorPercentThreshold:     50,
		CircuitBreakerTimeout:     1000,
	}

	// Create client
	client, err := InitializeClient(
		"TestCommand",
		connConfig,
		hystrixConfig,
		nil,
		0,
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, client)

	// Create request
	req, err := http.NewRequest("GET", testServer.URL, nil)
	require.NoError(t, err)

	// Execute request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "success", string(body))
}
