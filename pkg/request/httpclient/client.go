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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

type ConnectionPoolConfig struct {
	Timeout            int    `yaml:"timeout"`          // in milliseconds
	KeepAliveTimeout   int    `yaml:"keepAliveTimeout"` // in milliseconds
	MaxIdleConnections int    `yaml:"maxIdleConnections"`
	PrivateKeyPath     string `yaml:"-"`
	CertPath           string `yaml:"-"`
}

type HystrixResiliencyConfig struct {
	// MaxConcurrentRequests is the maximum number of concurrent requests allowed
	// Default is 100
	MaxConcurrentRequests int `yaml:"maxConcurrentRequests"`
	// RequestVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	// Default is 20
	RequestVolumeThreshold int `yaml:"requestVolumeThreshold"`
	// CircuitBreakerSleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	// Default is 5000
	CircuitBreakerSleepWindow int `yaml:"circuitBreakerSleepWindow"`
	// ErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	// Default is 50
	ErrorPercentThreshold int `yaml:"errorPercentThreshold"`
	// CircuitBreakerTimeout is how long to wait for command to complete, in milliseconds
	// Default is 1000
	CircuitBreakerTimeout int `yaml:"circuitBreakerTimeout"`
}

// InitializeClient initialises the client
func InitializeClient(hystrixCommand string, connectionPoolConfig ConnectionPoolConfig,
	hystrixConfig HystrixResiliencyConfig, retriable heimdall.Retriable,
	retryCount int, fallbackFunc func(error) error) (*hystrix.Client, error) {
	// for http conn pool
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: time.Duration(connectionPoolConfig.KeepAliveTimeout) * time.Millisecond,
		}).DialContext,
		MaxIdleConnsPerHost:   connectionPoolConfig.MaxIdleConnections,
		MaxIdleConns:          connectionPoolConfig.MaxIdleConnections,
		IdleConnTimeout:       time.Duration(connectionPoolConfig.KeepAliveTimeout) * time.Millisecond,
		TLSHandshakeTimeout:   time.Duration(connectionPoolConfig.Timeout) * time.Millisecond,
		ExpectContinueTimeout: time.Duration(connectionPoolConfig.Timeout) * time.Millisecond,
	}

	if len(connectionPoolConfig.PrivateKeyPath) > 0 && len(connectionPoolConfig.CertPath) > 0 {
		cert, err := tls.LoadX509KeyPair(connectionPoolConfig.CertPath, connectionPoolConfig.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate and key: %w", err)
		}

		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system cert pool: %w", err)
		}

		transport.TLSClientConfig = &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{cert},
		}
	}

	if retriable == nil {
		retriable = heimdall.NewNoRetrier()
	}

	options := []hystrix.Option{
		hystrix.WithHTTPClient(&http.Client{
			Transport: &nethttp.Transport{RoundTripper: transport},
		}),
		hystrix.WithHTTPTimeout(time.Duration(connectionPoolConfig.Timeout) * time.Millisecond),
		hystrix.WithCommandName(hystrixCommand),
		hystrix.WithHystrixTimeout(time.Duration(hystrixConfig.CircuitBreakerTimeout) * time.Millisecond),
		hystrix.WithMaxConcurrentRequests(hystrixConfig.MaxConcurrentRequests),
		hystrix.WithRequestVolumeThreshold(hystrixConfig.RequestVolumeThreshold),
		hystrix.WithSleepWindow(hystrixConfig.CircuitBreakerSleepWindow),
		hystrix.WithErrorPercentThreshold(hystrixConfig.ErrorPercentThreshold),
		hystrix.WithRetrier(retriable),
		hystrix.WithRetryCount(retryCount),
	}

	if fallbackFunc != nil {
		options = append(options, hystrix.WithFallbackFunc(fallbackFunc))
	}

	// Return a new hystrix-wrapped HTTP client with the command name, along with other required options
	return hystrix.NewClient(options...), nil
}
