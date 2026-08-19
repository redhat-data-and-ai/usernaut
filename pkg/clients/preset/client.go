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

package preset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
)

// NewClient creates a new Preset client with JWT-based authentication
func NewClient(presetAppConfig map[string]interface{},
	connectionPoolConfig httpclient.ConnectionPoolConfig,
	hystrixResiliencyConfig httpclient.HystrixResiliencyConfig) (*PresetClient, error) {

	presetConfig := PresetConfig{}
	if err := utils.MapToStruct(presetAppConfig, &presetConfig); err != nil {
		return nil, fmt.Errorf("failed to parse preset configuration: %w", err)
	}

	if presetConfig.BaseURL == "" {
		return nil, fmt.Errorf("preset configuration is missing required field: base_url")
	}
	if presetConfig.TeamSlug == "" {
		return nil, fmt.Errorf("preset configuration is missing required field: team_slug")
	}

	// Validate auth: either scim_token OR (api_token + api_secret) must be provided
	if presetConfig.SCIMToken == "" && (presetConfig.APIToken == "" || presetConfig.APISecret == "") {
		return nil, fmt.Errorf("preset configuration requires either scim_token or both api_token and api_secret")
	}

	apiURL := presetConfig.APIURL
	if apiURL == "" {
		apiURL = "https://api.app.preset.io"
	}

	client, err := httpclient.InitializeClient(
		"preset",
		connectionPoolConfig,
		hystrixResiliencyConfig,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)),
		3,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	return &PresetClient{
		client:    client,
		baseURL:   presetConfig.BaseURL,
		apiURL:    apiURL,
		apiToken:  presetConfig.APIToken,
		apiSecret: presetConfig.APISecret,
		scimToken: presetConfig.SCIMToken,
		teamSlug:  presetConfig.TeamSlug,
	}, nil
}

// scimURL returns the base SCIM v2 endpoint URL including the team slug.
// Format: {baseURL}/api/v1/teams/{teamSlug}/scim/v2
func (pc *PresetClient) scimURL() string {
	return fmt.Sprintf("%s/api/v1/teams/%s/scim/v2", pc.baseURL, pc.teamSlug)
}

// sendRequest makes an authenticated HTTP request to the Preset API
func (pc *PresetClient) sendRequest(
	ctx context.Context, url string, method string, body interface{},
) ([]byte, int, error) {
	log := logger.Logger(ctx).WithField("service", "preset")

	token, err := pc.getJWTToken(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get JWT token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if !slices.Contains([]int{http.StatusOK, http.StatusCreated, http.StatusNoContent}, resp.StatusCode) {
		log.WithField("status_code", resp.StatusCode).
			WithField("response", string(respBody)).
			Debug("unexpected response from Preset API")
		return respBody, resp.StatusCode, fmt.Errorf(
			"unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}
