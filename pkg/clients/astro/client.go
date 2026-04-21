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

package astro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
)

// NewClient creates a new Astro client with the given configuration
func NewClient(connection map[string]interface{}, poolCfg httpclient.ConnectionPoolConfig,
	hystrixCfg httpclient.HystrixResiliencyConfig) (*AstroClient, error) {

	// Extract connection parameters
	apiToken, _ := connection["api_token"].(string)
	organizationID, _ := connection["organization_id"].(string)
	baseURL, _ := connection["base_url"].(string)

	if apiToken == "" || organizationID == "" {
		return nil, errors.New("missing required astro connection params: api_token and organization_id")
	}

	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	config := AstroConfig{
		APIToken:       apiToken,
		OrganizationID: organizationID,
		BaseURL:        baseURL,
	}

	client, err := httpclient.InitializeClient(
		"astro",
		poolCfg,
		hystrixCfg,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)), 3,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	return &AstroClient{
		config: &config,
		client: client,
	}, nil
}

// prepareRequest creates and configures a request with common Astro headers
func (c *AstroClient) prepareRequest(ctx context.Context, endpoint, method string,
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

	// Set Astro-specific headers
	headers := map[string]string{
		"Authorization": "Bearer " + c.config.APIToken,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	req.SetHeaders(headers)

	return req, nil
}

// makeRequest uses the common request package for HTTP requests
func (c *AstroClient) makeRequest(ctx context.Context, endpoint,
	method string, body interface{}) ([]byte, int, error) {
	req, err := c.prepareRequest(ctx, endpoint, method, body)
	if err != nil {
		return nil, 0, err
	}

	return req.MakeRequest(c.client, method, "astro")
}

// makeRequestWithHeader uses the common request package for HTTP requests
// and returns headers (with logging, tracing, etc.)
func (c *AstroClient) makeRequestWithHeader(ctx context.Context, endpoint,
	method string, body interface{}) ([]byte, http.Header, int, error) {
	req, err := c.prepareRequest(ctx, endpoint, method, body)
	if err != nil {
		return nil, nil, 0, err
	}

	return req.MakeRequestWithHeader(c.client, method, "astro")
}

// fetchAllWithPagination handles paginated requests using offset-based pagination
func (c *AstroClient) fetchAllWithPagination(ctx context.Context,
	baseEndpoint string, processPage func([]byte) (int, error)) error {
	offset := 0
	limit := DefaultPageLimit

	for {
		// Build endpoint with pagination parameters
		endpoint := fmt.Sprintf("%s?offset=%d&limit=%d", baseEndpoint, offset, limit)

		resp, _, status, err := c.makeRequestWithHeader(ctx, endpoint, http.MethodGet, nil)
		if err != nil {
			return err
		}

		if status != http.StatusOK {
			return fmt.Errorf("failed to fetch data from %s, status: %s, body: %s",
				endpoint, http.StatusText(status), string(resp))
		}

		// Process page and get count of items returned
		count, err := processPage(resp)
		if err != nil {
			return err
		}

		// If fewer items than limit, we've reached the end
		if count < limit {
			break
		}

		offset += limit
	}

	return nil
}

// getOrganizationEndpoint returns the base endpoint for organization-scoped resources
func (c *AstroClient) getOrganizationEndpoint() string {
	return fmt.Sprintf("/v1/organizations/%s", c.config.OrganizationID)
}

// GetConfig returns the client configuration
func (c *AstroClient) GetConfig() *AstroConfig {
	return c.config
}
