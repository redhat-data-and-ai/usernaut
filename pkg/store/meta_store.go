package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// MetaStore handles all metadata-related cache operations with "meta:" prefix
// Metadata includes things like user lists, configuration, etc.
// NOTE: This store does NOT handle locking - callers must ensure proper synchronization
type MetaStore struct {
	cache cache.Cache
}

// newMetaStore creates a new MetaStore instance
func newMetaStore(c cache.Cache) *MetaStore {
	return &MetaStore{
		cache: c,
	}
}

// metaKey returns the prefixed cache key for metadata
func (s *MetaStore) metaKey(key string) string {
	return "meta:" + key
}

// GetUserList returns the list of active users
// Returns an empty slice if the list doesn't exist
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *MetaStore) GetUserList(ctx context.Context) ([]string, error) {
	key := s.metaKey("user_list")
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		// List not found, return empty slice
		return []string{}, nil
	}

	var userList []string
	if err := json.Unmarshal([]byte(val.(string)), &userList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user list: %w", err)
	}

	// Ensure we always return an empty slice instead of nil
	if userList == nil {
		return []string{}, nil
	}

	return userList, nil
}

// SetUserList sets the list of active users
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *MetaStore) SetUserList(ctx context.Context, users []string) error {
	key := s.metaKey("user_list")

	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("failed to marshal user list: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set user list in cache: %w", err)
	}

	return nil
}

// Get retrieves a generic metadata value by key
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *MetaStore) Get(ctx context.Context, key string) (string, error) {
	metaKey := s.metaKey(key)
	val, err := s.cache.Get(ctx, metaKey)
	if err != nil {
		return "", fmt.Errorf("failed to get meta key %s: %w", key, err)
	}

	return val.(string), nil
}

// Set stores a generic metadata value by key
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *MetaStore) Set(ctx context.Context, key, value string) error {
	metaKey := s.metaKey(key)
	if err := s.cache.Set(ctx, metaKey, value, cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set meta key %s: %w", key, err)
	}

	return nil
}

// Delete removes a metadata entry
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *MetaStore) Delete(ctx context.Context, key string) error {
	metaKey := s.metaKey(key)
	return s.cache.Delete(ctx, metaKey)
}
