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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	"github.com/sirupsen/logrus"
)

// apiError represents a non-success HTTP response from the Preset SCIM API.
type apiError struct {
	StatusCode int
	Body       []byte
}

func (e *apiError) Error() string {
	return fmt.Sprintf("unexpected status code: %d, response: %s", e.StatusCode, string(e.Body))
}

func responseStatus(err error) int {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return 0
}

func isResponseStatus(err error, statusCode int) bool {
	return responseStatus(err) == statusCode
}

const presetHTTPRetryAttempts = 3

// presetHTTPRetryBackoff is longer than other backends to reduce Preset SCIM rate-limit pressure.
var presetHTTPRetryBackoff = 5 * time.Second

const (
	presetRateLimitRetryAttempts  = 3
	presetRateLimitDefaultBackoff = 5 * time.Second
	presetRateLimitMaxBackoff     = 60 * time.Second
)

func rateLimitBackoff(headers http.Header) time.Duration {
	retryAfter := strings.TrimSpace(headers.Get("Retry-After"))
	if retryAfter == "" {
		return presetRateLimitDefaultBackoff
	}

	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		backoff := time.Duration(seconds) * time.Second
		return capRateLimitBackoff(backoff)
	}

	if retryAt, err := http.ParseTime(retryAfter); err == nil {
		return capRateLimitBackoff(time.Until(retryAt))
	}

	return presetRateLimitDefaultBackoff
}

func capRateLimitBackoff(backoff time.Duration) time.Duration {
	if backoff < 0 {
		return 0
	}
	if backoff > presetRateLimitMaxBackoff {
		return presetRateLimitMaxBackoff
	}
	return backoff
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// NewClient creates a new Preset client with SCIM token authentication
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
	if presetConfig.SCIMToken == "" {
		return nil, fmt.Errorf("preset configuration is missing required field: scim_token")
	}

	client, err := httpclient.InitializeClient(
		"preset",
		connectionPoolConfig,
		hystrixResiliencyConfig,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(presetHTTPRetryBackoff, 0)),
		presetHTTPRetryAttempts,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	return &PresetClient{
		client:    client,
		baseURL:   strings.TrimRight(presetConfig.BaseURL, "/"),
		scimToken: presetConfig.SCIMToken,
		teamSlug:  presetConfig.TeamSlug,
	}, nil
}

// scimURL returns the base SCIM v2 endpoint URL including the team slug.
// Format: {baseURL}/api/v1/teams/{teamSlug}/scim/v2
func (pc *PresetClient) scimURL() string {
	return fmt.Sprintf("%s/api/v1/teams/%s/scim/v2", pc.baseURL, url.PathEscape(pc.teamSlug))
}

func (pc *PresetClient) userURL(userID string) string {
	return fmt.Sprintf("%s/Users/%s", pc.scimURL(), url.PathEscape(userID))
}

func (pc *PresetClient) groupURL(teamID string) string {
	return fmt.Sprintf("%s/Groups/%s", pc.scimURL(), url.PathEscape(teamID))
}

// sendRequest makes an authenticated HTTP request to the Preset API using pkg/request.
func (pc *PresetClient) sendRequest(
	ctx context.Context, reqURL string, method string, body interface{},
) ([]byte, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
	})

	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	reqHeaders := map[string]string{
		"Authorization": "Bearer " + pc.scimToken,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	for attempt := 0; ; attempt++ {
		req, err := request.NewRequest(ctx, method, reqURL, requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.SetHeaders(reqHeaders)

		respBody, respHeaders, statusCode, err := req.MakeRequestWithHeader(pc.client, method, "preset")
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusNoContent {
			return respBody, nil
		}

		if statusCode == http.StatusTooManyRequests && attempt < presetRateLimitRetryAttempts {
			backoff := rateLimitBackoff(respHeaders)
			log.WithFields(logrus.Fields{
				"attempt":      attempt + 1,
				"max_attempts": presetRateLimitRetryAttempts,
				"backoff_ms":   backoff.Milliseconds(),
			}).Warn("Preset API rate limit exceeded, retrying after backoff")
			if err := sleepWithContext(ctx, backoff); err != nil {
				return nil, err
			}
			continue
		}

		const maxLogBodyLen = 512
		responseBodyPreview := string(respBody)
		if len(responseBodyPreview) > maxLogBodyLen {
			responseBodyPreview = responseBodyPreview[:maxLogBodyLen] + "...(truncated)"
		}

		log.WithFields(logrus.Fields{
			"status_code":           statusCode,
			"response_body_preview": responseBodyPreview,
			"response_body_size":    len(respBody),
		}).Debug("unexpected response from Preset API")
		return nil, &apiError{StatusCode: statusCode, Body: respBody}
	}
}

func escapeSCIMLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func (pc *PresetClient) querySCIMByFilter(
	ctx context.Context, resource string, filter string, fieldKey string, fieldValue string,
) ([]byte, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		fieldKey:  fieldValue,
	})

	reqURL := fmt.Sprintf("%s/%s?filter=%s", pc.scimURL(), resource, url.QueryEscape(filter))
	response, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		log.WithError(err).Errorf("failed to query SCIM %s by filter", resource)
		return nil, err
	}
	return response, nil
}
