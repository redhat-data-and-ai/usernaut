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

package snowflake

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
)

// Compiled regex patterns for Link header parsing (performance optimization)
var (
	linkPattern    = regexp.MustCompile(`<([^>]+)>\s*;\s*(?:[^,]*;\s*)*rel="([^"]+)"(?:\s*;[^,]*)*`)
	reversePattern = regexp.MustCompile(`<([^>]+)>\s*;\s*rel="([^"]+)"(?:\s*;[^,]*)*`)
)

// SnowflakeConfig holds the configuration for Snowflake client
type SnowflakeConfig struct {
	PAT     string
	BaseURL string
}

// SnowflakeClient is the client for interacting with Snowflake REST API
type SnowflakeClient struct {
	config *SnowflakeConfig
	client heimdall.Doer
}

// NewClient creates a new Snowflake client with the given configuration
func NewClient(config SnowflakeConfig, poolCfg httpclient.ConnectionPoolConfig,
	hystrixCfg httpclient.HystrixResiliencyConfig) (*SnowflakeClient, error) {
	client, err := httpclient.InitializeClient(
		"snowflake",
		poolCfg,
		hystrixCfg,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)), 3,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	return &SnowflakeClient{
		config: &config,
		client: client,
	}, nil
}

// sendRequest sends a HTTP request to the Snowflake REST API and returns response body, headers, and status
func (c *SnowflakeClient) sendRequest(ctx context.Context,
	endpoint, method string, body interface{}) ([]byte, http.Header,
	int, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, 0, err
	}

	url := c.config.BaseURL + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, 0, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.PAT)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	response, err := c.client.Do(httpReq)
	if err != nil {
		return nil, nil, http.StatusBadGateway, err
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, http.StatusBadGateway, err
	}

	return responseBody, response.Header, response.StatusCode, nil
}

// parseLinkHeader parses the Link header and returns the URL for the specified rel
// Uses regex to handle diverse Link header formats robustly, including:
// - Multiple parameters (e.g., rel="next"; title="Next Page")
// - Varying whitespace
// - Different parameter orders
func parseLinkHeader(linkHeader, rel string) string {
	// RFC 5988 compliant regex pattern for Link header parsing
	// Matches: <url>; rel="value" (with optional whitespace and other parameters)
	matches := linkPattern.FindAllStringSubmatch(linkHeader, -1)

	for _, match := range matches {
		if len(match) == 3 && match[2] == rel {
			return match[1]
		}
	}

	// Also try the reverse pattern: rel="value" before other parameters
	reverseMatches := reversePattern.FindAllStringSubmatch(linkHeader, -1)

	for _, match := range reverseMatches {
		if len(match) == 3 && match[2] == rel {
			return match[1]
		}
	}

	return ""
}

// GetConfig returns the client configuration
func (c *SnowflakeClient) GetConfig() *SnowflakeConfig {
	return c.config
}
