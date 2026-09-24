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
	"math"
	"math/rand/v2"
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

// rateLimitJitter applies full jitter in [0, d]. Overridable in tests for determinism.
var rateLimitJitter = fullJitter

func fullJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(d) + 1))
}

// exponentialRateLimitBackoff returns base * 2^attempt, capped at max.
func exponentialRateLimitBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	backoff := float64(presetRateLimitBaseBackoff) * math.Pow(presetRateLimitExpFactor, float64(attempt))
	if backoff > float64(presetRateLimitMaxBackoff) {
		return presetRateLimitMaxBackoff
	}
	return time.Duration(backoff)
}

// parseRetryAfter returns the wait duration from a Retry-After header when present.
// ok is false when the header is missing or unparseable.
func parseRetryAfter(headers http.Header) (time.Duration, bool) {
	retryAfter := strings.TrimSpace(headers.Get("Retry-After"))
	if retryAfter == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return capRateLimitBackoff(time.Duration(seconds) * time.Second), true
	}

	if retryAt, err := http.ParseTime(retryAfter); err == nil {
		return capRateLimitBackoff(time.Until(retryAt)), true
	}

	return 0, false
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

// rateLimitBackoff chooses how long to wait before retrying a 429.
//
// Step A: When Retry-After is present it is honored in full (capped at
// presetRateLimitMaxBackoff), including 0 = retry immediately.
//
// Step B: When Retry-After is missing, fall back to exponential backoff
// Delay = BaseDelay * 2^attempt, with full jitter so concurrent reconciles
// do not retry in lockstep.
//
// Step C (retry budget) is enforced by the caller: at most
// presetRateLimitRetryAttempts retries and delays are capped at
// presetRateLimitMaxBackoff.
//
// log may be nil (used by unit tests).
func rateLimitBackoff(log *logrus.Entry, headers http.Header, attempt int) time.Duration {
	retryAfterHeader := strings.TrimSpace(headers.Get("Retry-After"))
	if retryAfter, ok := parseRetryAfter(headers); ok {
		if log != nil {
			log.WithFields(logrus.Fields{
				logKeyAttempt:    attempt + 1,
				"backoff_source": "retry_after",
				"retry_after":    retryAfterHeader,
				"retry_after_ms": retryAfter.Milliseconds(),
				"backoff_ms":     retryAfter.Milliseconds(),
				"max_backoff_ms": presetRateLimitMaxBackoff.Milliseconds(),
			}).Info("computed Preset rate limit backoff from Retry-After header")
		}
		return retryAfter
	}

	exponential := exponentialRateLimitBackoff(attempt)
	wait := rateLimitJitter(exponential)
	if log != nil {
		log.WithFields(logrus.Fields{
			logKeyAttempt:     attempt + 1,
			"backoff_source":  "exponential",
			"retry_after":     retryAfterHeader,
			"exponential_ms":  exponential.Milliseconds(),
			"backoff_ms":      wait.Milliseconds(),
			"base_backoff_ms": presetRateLimitBaseBackoff.Milliseconds(),
			"exp_factor":      presetRateLimitExpFactor,
			"max_backoff_ms":  presetRateLimitMaxBackoff.Milliseconds(),
		}).Info("computed Preset rate limit backoff using exponential strategy")
	}
	return wait
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
		serviceName,
		connectionPoolConfig,
		hystrixResiliencyConfig,
		heimdall.NewRetrier(heimdall.NewExponentialBackoff(
			presetHTTPRetryInitialBackoff,
			presetHTTPRetryMaxBackoff,
			presetHTTPRetryExpFactor,
			presetHTTPRetryMaxJitter,
		)),
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
		logKeyService: serviceName,
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

	logURL := requestLogURL(reqURL)

	for attempt := 0; ; attempt++ {
		log.WithFields(logrus.Fields{
			logKeyAttempt: attempt + 1,
			"max_retries": presetRateLimitRetryAttempts,
			"method":      method,
			"url":         logURL,
		}).Debug("sending Preset API request")

		req, err := request.NewRequest(ctx, method, reqURL, requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.SetHeaders(reqHeaders)

		respBody, respHeaders, statusCode, err := req.MakeRequestWithHeader(pc.client, method, serviceName)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				logKeyAttempt: attempt + 1,
				"method":      method,
				"url":         logURL,
			}).Error("Preset API request failed")
			return nil, fmt.Errorf("request failed: %w", err)
		}

		log.WithFields(logrus.Fields{
			logKeyAttempt: attempt + 1,
			"status_code": statusCode,
			"method":      method,
			"url":         logURL,
		}).Debug("received Preset API response")

		if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusNoContent {
			if attempt > 0 {
				log.WithFields(logrus.Fields{
					logKeyAttempt: attempt + 1,
					"status_code": statusCode,
				}).Info("Preset API request succeeded after rate limit retry")
			}
			return respBody, nil
		}

		if statusCode == http.StatusTooManyRequests {
			if attempt >= presetRateLimitRetryAttempts {
				log.WithFields(logrus.Fields{
					logKeyAttempt: attempt + 1,
					"max_retries": presetRateLimitRetryAttempts,
					"status_code": statusCode,
					"retry_after": respHeaders.Get("Retry-After"),
					"method":      method,
					"url":         logURL,
				}).Error("Preset API rate limit retries exhausted")
			} else {
				log.WithFields(logrus.Fields{
					logKeyAttempt: attempt + 1,
					"max_retries": presetRateLimitRetryAttempts,
					"status_code": statusCode,
					"retry_after": respHeaders.Get("Retry-After"),
					"method":      method,
					"url":         logURL,
				}).Warn("Preset API rate limit exceeded (429)")

				backoff := rateLimitBackoff(log, respHeaders, attempt)
				log.WithFields(logrus.Fields{
					logKeyAttempt: attempt + 1,
					"backoff_ms":  backoff.Milliseconds(),
				}).Info("waiting before Preset API rate limit retry")

				if err := sleepWithContext(ctx, backoff); err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						logKeyAttempt: attempt + 1,
						"backoff_ms":  backoff.Milliseconds(),
					}).Warn("Preset API rate limit backoff interrupted")
					return nil, err
				}

				log.WithFields(logrus.Fields{
					logKeyAttempt:  attempt + 1,
					"next_attempt": attempt + 2,
				}).Info("resuming Preset API request after rate limit backoff")
				continue
			}
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
			logKeyAttempt:           attempt + 1,
		}).Debug("unexpected response from Preset API")
		return nil, &apiError{StatusCode: statusCode, Body: respBody}
	}
}

// requestLogURL returns a URL safe to write to logs. SCIM filter values can
// contain usernames, so the filter query parameter is redacted. The original
// request URL is left unchanged.
func requestLogURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	q := parsed.Query()
	if _, ok := q["filter"]; !ok {
		return parsed.String()
	}
	q.Set("filter", "[redacted]")
	parsed.RawQuery = q.Encode()
	return parsed.String()
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
		logKeyService: serviceName,
		fieldKey:      fieldValue,
	})

	reqURL := fmt.Sprintf("%s/%s?filter=%s", pc.scimURL(), resource, url.QueryEscape(filter))
	response, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		log.WithError(err).Errorf("failed to query SCIM %s by filter", resource)
		return nil, err
	}
	return response, nil
}
