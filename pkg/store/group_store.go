package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// BackendInfo represents backend metadata stored for a group
type BackendInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GroupData represents the consolidated data stored for a group
// Key format: "group:<groupName>"
type GroupData struct {
	Members  []string               `json:"members"`
	Backends map[string]BackendInfo `json:"backends"` // key: "backendName_backendType"
}

// GroupStore handles consolidated group cache operations
// Key format: "group:<groupName>"
// Value: JSON object with members and backends
// NOTE: This store does NOT handle locking - callers must ensure proper synchronization
type GroupStore struct {
	cache cache.Cache
}

// newGroupStore creates a new GroupStore instance
func newGroupStore(c cache.Cache) *GroupStore {
	return &GroupStore{
		cache: c,
	}
}

// groupKey returns the prefixed cache key for a group
func (s *GroupStore) groupKey(groupName string) string {
	return "group:" + groupName
}

// backendKey returns the composite key for a backend
func backendKey(backendName, backendType string) string {
	return backendName + "_" + backendType
}

// Get retrieves the full group data from cache
// Returns empty GroupData if the group is not found in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) Get(ctx context.Context, groupName string) (*GroupData, error) {
	key := s.groupKey(groupName)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		// Group not found, return empty data
		return &GroupData{
			Members:  []string{},
			Backends: make(map[string]BackendInfo),
		}, nil
	}

	var data GroupData
	if err := json.Unmarshal([]byte(val.(string)), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal group data: %w", err)
	}

	// Ensure maps and slices are initialized
	if data.Members == nil {
		data.Members = []string{}
	}
	if data.Backends == nil {
		data.Backends = make(map[string]BackendInfo)
	}

	return &data, nil
}

// Set stores the full group data in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) Set(ctx context.Context, groupName string, data *GroupData) error {
	key := s.groupKey(groupName)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal group data: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(jsonData), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set group data in cache: %w", err)
	}

	return nil
}

// Delete removes a group entirely from cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) Delete(ctx context.Context, groupName string) error {
	key := s.groupKey(groupName)
	return s.cache.Delete(ctx, key)
}

// Exists checks if a group exists in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) Exists(ctx context.Context, groupName string) (bool, error) {
	key := s.groupKey(groupName)
	_, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// --- Member Operations ---

// GetMembers returns the list of user emails for a group
// Returns an empty slice if the group is not found in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) GetMembers(ctx context.Context, groupName string) ([]string, error) {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return nil, err
	}
	return data.Members, nil
}

// SetMembers sets the complete list of user emails for a group
// This replaces any existing members while preserving backends
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) SetMembers(ctx context.Context, groupName string, members []string) error {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return err
	}

	data.Members = members
	return s.Set(ctx, groupName, data)
}

// --- Backend Operations ---

// GetBackends returns a map of backend info for a group
// Returns an empty map if the group is not found in cache
// Map format: {"backend_name_type": BackendInfo{ID, Name, Type}}
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) GetBackends(ctx context.Context, groupName string) (map[string]BackendInfo, error) {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return nil, err
	}
	return data.Backends, nil
}

// GetBackendID returns the backend ID for a specific backend
// Returns empty string if the backend is not found
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) GetBackendID(ctx context.Context, groupName, backendName, backendType string) (string, error) {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return "", err
	}

	key := backendKey(backendName, backendType)
	if backend, exists := data.Backends[key]; exists {
		return backend.ID, nil
	}
	return "", nil
}

// SetBackend sets a backend for a group
// If the group doesn't exist, it will be created
// If the backend exists, it will be updated
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) SetBackend(ctx context.Context, groupName, backendName, backendType, backendID string) error {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return err
	}

	key := backendKey(backendName, backendType)
	data.Backends[key] = BackendInfo{
		ID:   backendID,
		Name: backendName,
		Type: backendType,
	}

	return s.Set(ctx, groupName, data)
}

// DeleteBackend removes a specific backend from a group's record
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) DeleteBackend(ctx context.Context, groupName, backendName, backendType string) error {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return err
	}

	key := backendKey(backendName, backendType)
	delete(data.Backends, key)

	return s.Set(ctx, groupName, data)
}

// BackendExists checks if a specific backend exists for a group
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) BackendExists(ctx context.Context, groupName, backendName, backendType string) (bool, error) {
	data, err := s.Get(ctx, groupName)
	if err != nil {
		return false, err
	}

	key := backendKey(backendName, backendType)
	_, exists := data.Backends[key]
	return exists, nil
}
