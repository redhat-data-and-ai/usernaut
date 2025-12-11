package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// UserStore handles all user-related cache operations with "user:" prefix
// NOTE: This store does NOT handle locking - callers must ensure proper synchronization
type UserStore struct {
	cache cache.Cache
}

// newUserStore creates a new UserStore instance
func newUserStore(c cache.Cache) *UserStore {
	return &UserStore{
		cache: c,
	}
}

// userKey returns the prefixed cache key for a user
func (s *UserStore) userKey(email string) string {
	return "user:" + email
}

// GetBackends returns a map of backend IDs for a user
// Returns an empty map if the user is not found in cache
// Map format: {"backend_name_type": "backend_user_id"}
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) GetBackends(ctx context.Context, email string) (map[string]string, error) {
	key := s.userKey(email)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		// User not found, return empty map (not an error condition)
		return make(map[string]string), nil
	}

	var backends map[string]string
	if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user backends: %w", err)
	}

	return backends, nil
}

// SetBackend sets a backend ID for a user
// If the user doesn't exist, it will be created
// If the user exists, the backend ID will be added/updated in the map
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) SetBackend(ctx context.Context, email, backendKey, backendID string) error {
	key := s.userKey(email)

	// Get existing backends or create new map
	backends := make(map[string]string)
	val, err := s.cache.Get(ctx, key)
	if err == nil {
		// User exists, unmarshal existing data
		if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
			return fmt.Errorf("failed to unmarshal existing user backends: %w", err)
		}
	}

	// Update the backend ID
	backends[backendKey] = backendID

	// Marshal and store back
	data, err := json.Marshal(backends)
	if err != nil {
		return fmt.Errorf("failed to marshal user backends: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set user in cache: %w", err)
	}

	return nil
}

// DeleteBackend removes a specific backend ID from a user's record
// If this was the last backend, the entire user entry is deleted
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) DeleteBackend(ctx context.Context, email, backendKey string) error {
	key := s.userKey(email)
	return deleteBackendHelper(ctx, s.cache, key, backendKey, "user")
}

// Delete removes a user entirely from cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) Delete(ctx context.Context, email string) error {
	key := s.userKey(email)
	return s.cache.Delete(ctx, key)
}

// Exists checks if a user exists in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) Exists(ctx context.Context, email string) (bool, error) {
	key := s.userKey(email)
	_, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetByPattern searches for users matching a pattern and returns their data
// Pattern should NOT include the "user:" prefix - it will be added automatically
// Example: pattern "*@example.com" searches for "user:*@example.com"
// Returns: map[email]backends where backends is map[backendKey]backendID
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) GetByPattern(ctx context.Context, pattern string) (map[string]map[string]string, error) {
	// Add user: prefix to the pattern
	fullPattern := "user:" + pattern

	// Use cache's GetByPattern to find matching keys
	results, err := s.cache.GetByPattern(ctx, fullPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users by pattern: %w", err)
	}

	// Parse results into structured format
	userMap := make(map[string]map[string]string)
	for key, value := range results {
		// Extract email from key (remove "user:" prefix)
		email := strings.TrimPrefix(key, "user:")

		var backends map[string]string
		if err := json.Unmarshal([]byte(value.(string)), &backends); err != nil {
			// Skip invalid entries
			continue
		}

		userMap[email] = backends
	}

	return userMap, nil
}
