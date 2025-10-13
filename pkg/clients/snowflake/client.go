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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
)

// Compiled regex patterns for Link header parsing (performance optimization)
var (
	linkPattern    = regexp.MustCompile(`<([^>]+)>\s*;\s*(?:[^,]*;\s*)*rel="([^"]+)"(?:\s*;[^,]*)*`)
	reversePattern = regexp.MustCompile(`<([^>]+)>\s*;\s*rel="([^"]+)"(?:\s*;[^,]*)*`)
)

// NewClient creates a new Snowflake client with the given configuration
func NewClient(connection map[string]interface{}, poolCfg httpclient.ConnectionPoolConfig,
	hystrixCfg httpclient.HystrixResiliencyConfig) (*SnowflakeClient, error) {

	// Extract connection parameters
	// pat, _ := connection["pat"].(string)
	// baseURL, _ := connection["base_url"].(string)

	// if pat == "" || baseURL == "" {
	// 	return nil, errors.New("missing required connection
	// parameters for snowflake backend: pat and base_url are required")
	// }

	config := SnowflakeConfig{}
	if err := utils.MapToStruct(connection, &config); err != nil {
		return nil, err
	}

	// config := SnowflakeConfig{
	// 	PAT:     pat,
	// 	BaseURL: baseURL,
	// }
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

// prepareRequest creates and configures a request with common Snowflake headers
func (c *SnowflakeClient) prepareRequest(ctx context.Context, endpoint, method string,
	body interface{}) (request.IRequester, error) {
	var requestBody []byte
	if body != nil && (method != http.MethodGet && method != http.MethodDelete) {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	url := c.config.BaseURL + endpoint
	req, err := request.NewRequest(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}

	// Set Snowflake-specific headers
	headers := map[string]string{
		"Authorization": "Bearer " + c.config.PAT,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	req.SetHeaders(headers)

	return req, nil
}

// makeRequest uses the common request package for standard HTTP requests (with logging, tracing, etc.)
func (c *SnowflakeClient) makeRequest(ctx context.Context, endpoint,
	method string, body interface{}) ([]byte, int, error) {
	req, err := c.prepareRequest(ctx, endpoint, method, body)
	if err != nil {
		return nil, 0, err
	}

	return req.MakeRequest(c.client, method, "snowflake")
}

// makeRequestWithHeader uses the common request package for HTTP requests
// and returns headers (with logging, tracing, etc.)
func (c *SnowflakeClient) makeRequestWithHeader(ctx context.Context, endpoint,
	method string, body interface{}) ([]byte, http.Header, int, error) {
	req, err := c.prepareRequest(ctx, endpoint, method, body)
	if err != nil {
		return nil, nil, 0, err
	}

	return req.MakeRequestWithHeader(c.client, method, "snowflake")
}

func (c *SnowflakeClient) fetchAllWithPagination(ctx context.Context,
	endpoint string, processPage func([]byte) error) error {
	// First request to get initial page and Link header
	resp, headers, status, err := c.makeRequestWithHeader(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("failed to fetch data from %s, status: %s, body: %s",
			endpoint, http.StatusText(status), string(resp))
	}

	// Process first page
	if err := processPage(resp); err != nil {
		return err
	}

	// Check for additional pages in Link header
	linkHeader := headers.Get("Link")
	if linkHeader != "" {
		nextURL := parseLinkHeader(linkHeader, "next")

		// Follow pagination using Link header URLs
		for nextURL != "" {
			resp, headers, status, err := c.makeRequestWithHeader(ctx, nextURL, http.MethodGet, nil)
			if err != nil {
				return err
			}
			if status != http.StatusOK {
				return fmt.Errorf("unexpected status during pagination: %s, body: %s", http.StatusText(status), string(resp))
			}

			// Process this page
			if err := processPage(resp); err != nil {
				return err
			}

			// Get next page URL
			linkHeader = headers.Get("Link")
			if linkHeader != "" {
				nextURL = parseLinkHeader(linkHeader, "next")
			} else {
				nextURL = ""
			}
		}
	}

	return nil
}

func parseLinkHeader(linkHeader, rel string) string {
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
