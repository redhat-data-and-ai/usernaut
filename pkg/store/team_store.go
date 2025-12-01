package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// TeamStore handles team-related cache operations with "team:" prefix
// Key format: "team:<transformedTeamName>"
// Value: JSON map of {"backend_name_type": "backend_team_id"}
// This store is used for preloading team data from backends where teams
// are identified by their transformed names.
// NOTE: This store does NOT handle locking - callers must ensure proper synchronization
type TeamStore struct {
	cache cache.Cache
}

// newTeamStore creates a new TeamStore instance
func newTeamStore(c cache.Cache) *TeamStore {
	return &TeamStore{
		cache: c,
	}
}

// teamKey returns the prefixed cache key for a team
func (s *TeamStore) teamKey(teamName string) string {
	return "team:" + teamName
}

// GetBackends returns a map of backend IDs for a team
// Returns an empty map if the team is not found in cache
// Map format: {"backend_name_type": "backend_team_id"}
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *TeamStore) GetBackends(ctx context.Context, teamName string) (map[string]string, error) {
	key := s.teamKey(teamName)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		// Team not found, return empty map (not an error condition)
		return make(map[string]string), nil
	}

	var backends map[string]string
	if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
		return nil, fmt.Errorf("failed to unmarshal team backends: %w", err)
	}

	return backends, nil
}

// SetBackend sets a backend ID for a team
// If the team doesn't exist, it will be created
// If the team exists, the backend ID will be added/updated in the map
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *TeamStore) SetBackend(ctx context.Context, teamName, backendKey, teamID string) error {
	key := s.teamKey(teamName)

	// Get existing backends or create new map
	backends := make(map[string]string)
	val, err := s.cache.Get(ctx, key)
	if err == nil {
		// Team exists, unmarshal existing data
		if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
			return fmt.Errorf("failed to unmarshal existing team backends: %w", err)
		}
	}

	// Update the backend ID
	backends[backendKey] = teamID

	// Marshal and store back
	data, err := json.Marshal(backends)
	if err != nil {
		return fmt.Errorf("failed to marshal team backends: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set team in cache: %w", err)
	}

	return nil
}

// DeleteBackend removes a specific backend ID from a team's record
// If this was the last backend, the entire team entry is deleted
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *TeamStore) DeleteBackend(ctx context.Context, teamName, backendKey string) error {
	key := s.teamKey(teamName)
	return deleteBackendHelper(ctx, s.cache, key, backendKey, "team")
}

// Delete removes a team entirely from cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *TeamStore) Delete(ctx context.Context, teamName string) error {
	key := s.teamKey(teamName)
	return s.cache.Delete(ctx, key)
}

// Exists checks if a team exists in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *TeamStore) Exists(ctx context.Context, teamName string) (bool, error) {
	key := s.teamKey(teamName)
	_, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}
