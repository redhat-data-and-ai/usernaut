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

package atlan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	atlansdk "github.com/atlanhq/atlan-go/atlan/assets"
	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
)

// NewClient creates a new Atlan client with simple API token authentication
func NewClient(atlanAppConfig map[string]interface{},
	connectionPoolConfig httpclient.ConnectionPoolConfig,
	hystrixResiliencyConfig httpclient.HystrixResiliencyConfig) (*AtlanClient, error) {

	atlanConfig := AtlanConfig{}
	if err := utils.MapToStruct(atlanAppConfig, &atlanConfig); err != nil {
		return nil, fmt.Errorf("failed to parse atlan configuration: %w", err)
	}

	// Validate required fields
	if atlanConfig.URL == "" {
		return nil, fmt.Errorf("atlan configuration is missing required field: URL")
	}
	if atlanConfig.APIToken == "" {
		return nil, fmt.Errorf("atlan configuration is missing required field: APIToken")
	}

	// Initialize HTTP client without certificates (Atlan uses API token, not certs)
	client, err := httpclient.InitializeClient(
		"atlan",
		connectionPoolConfig,
		hystrixResiliencyConfig,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)), // retry logic
		3,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	// Initialize the Atlan SDK client for operations that require it (e.g., DeleteUser)
	sdkClient, err := atlansdk.Context(atlanConfig.URL, atlanConfig.APIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Atlan SDK client: %w", err)
	}

	return &AtlanClient{
		client:                client,
		sdkClient:             sdkClient,
		url:                   atlanConfig.URL,
		apiToken:              atlanConfig.APIToken,
		identityProviderAlias: atlanConfig.IdentityProviderAlias,
		defaultOwnerUserName:  atlanConfig.DefaultOwnerUserName,
	}, nil
}

// sendRequest makes an HTTP request to the Atlan API with proper authentication
func (aC *AtlanClient) sendRequest(ctx context.Context, url string, method string, body interface{},
	headers map[string]string, methodName string) ([]byte, int, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := request.NewRequest(ctx, method, url, requestBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + aC.apiToken
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	req.SetHeaders(headers)

	response, statusCode, err := req.MakeRequest(aC.client, methodName, "atlan")
	if err != nil {
		return nil, statusCode, fmt.Errorf("request failed: %w", err)
	}

	if !slices.Contains([]int{http.StatusOK, http.StatusCreated, http.StatusNoContent}, statusCode) {
		return response, statusCode, fmt.Errorf("unexpected status code: %d, response: %s", statusCode, string(response))
	}

	return response, statusCode, nil
}

// SetSSOSync sets the SSO sync flag for the Atlan client
// When SSO sync is enabled, user creation is handled automatically via SSO provider
func (ac *AtlanClient) SetSSOSync(ssoSync bool) {
	ac.ssoSync = ssoSync
}

// SetLdapSync sets the LDAP sync flag and the SSO group name for the Atlan client
// When LDAP sync is enabled (depends_on rover is defined), SSO group mapping is created
func (ac *AtlanClient) SetLdapSync(ldapSync bool, ssoGroupName string) {
	ac.ldapSync = ldapSync
	ac.ssoGroupName = ssoGroupName
}
