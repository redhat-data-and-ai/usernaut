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
	"time"

	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

const tokenRefreshBuffer = 5 * time.Minute

// getJWTToken returns a valid Bearer token for API requests.
// If a SCIM token is configured, returns it directly (no exchange needed).
// Otherwise, fetches/refreshes a JWT token via api.app.preset.io.
// Thread-safe via mutex.
func (pc *PresetClient) getJWTToken(ctx context.Context) (string, error) {
	if pc.scimToken != "" {
		return pc.scimToken, nil
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.jwtToken != "" && time.Now().Before(pc.jwtExpiry.Add(-tokenRefreshBuffer)) {
		return pc.jwtToken, nil
	}

	token, expiry, err := pc.fetchJWTToken(ctx)
	if err != nil {
		return "", err
	}

	pc.jwtToken = token
	pc.jwtExpiry = expiry
	return token, nil
}

// fetchJWTToken exchanges API token + secret for a short-lived JWT via api.app.preset.io
func (pc *PresetClient) fetchJWTToken(ctx context.Context) (string, time.Time, error) {
	log := logger.Logger(ctx).WithField("service", "preset")
	log.Debug("fetching new JWT token from Preset")

	url := fmt.Sprintf("%s/v1/auth/", pc.apiURL)
	reqBody, _ := json.Marshal(map[string]string{
		"name":   pc.apiToken,
		"secret": pc.apiSecret,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := pc.client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read auth response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("auth request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp authResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("empty access token in auth response")
	}

	expiry := time.Unix(authResp.Payload.Exp, 0)
	if expiry.IsZero() || expiry.Before(time.Now()) {
		// Default to 1 hour if expiry is missing or invalid
		expiry = time.Now().Add(1 * time.Hour)
	}

	log.Debug("successfully fetched JWT token from Preset")
	return authResp.AccessToken, expiry, nil
}
