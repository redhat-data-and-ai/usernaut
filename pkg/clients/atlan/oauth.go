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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthTokenManager handles OAuth token generation for user deletion
type OAuthTokenManager struct {
	baseURL      string
	clientID     string
	clientSecret string

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
}

// NewOAuthTokenManager creates a new OAuthTokenManager
func NewOAuthTokenManager(baseURL, clientID, clientSecret string) *OAuthTokenManager {
	return &OAuthTokenManager{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// GetToken returns a valid OAuth token, refreshing if needed
func (tm *OAuthTokenManager) GetToken() (string, error) {
	tm.mu.RLock()
	if tm.accessToken != "" && time.Now().Add(30*time.Second).Before(tm.expiresAt) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after lock
	if tm.accessToken != "" && time.Now().Add(30*time.Second).Before(tm.expiresAt) {
		return tm.accessToken, nil
	}

	// Fetch new token
	tokenURL := fmt.Sprintf("%s/auth/realms/default/protocol/openid-connect/token", strings.TrimSuffix(tm.baseURL, "/"))

	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", tm.clientID)
	formData.Set("client_secret", tm.clientSecret)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	tm.accessToken = tokenResp.AccessToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tm.accessToken, nil
}
