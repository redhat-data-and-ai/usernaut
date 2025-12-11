package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// UserGroupsStore handles user-to-groups reverse index cache operations
// Key format: "user:groups:<email>"
// Value: JSON array of group names
// NOTE: This store does NOT handle locking - callers must ensure proper synchronization
type UserGroupsStore struct {
	cache cache.Cache
}

// newUserGroupsStore creates a new UserGroupsStore instance
func newUserGroupsStore(c cache.Cache) *UserGroupsStore {
	return &UserGroupsStore{
		cache: c,
	}
}

// userGroupsKey returns the prefixed cache key for user's groups
func (s *UserGroupsStore) userGroupsKey(email string) string {
	return "user:groups:" + email
}

// GetGroups returns the list of groups for a user
// Returns an empty slice if the user is not found in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) GetGroups(ctx context.Context, email string) ([]string, error) {
	key := s.userGroupsKey(email)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		// User not found, return empty slice (not an error condition)
		return []string{}, nil
	}

	var groups []string
	if err := json.Unmarshal([]byte(val.(string)), &groups); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user groups: %w", err)
	}

	// Ensure we always return an empty slice instead of nil
	if groups == nil {
		return []string{}, nil
	}

	return groups, nil
}

// AddGroup adds a group to a user's group list if not already present
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) AddGroup(ctx context.Context, email, groupName string) error {
	key := s.userGroupsKey(email)

	// Get existing groups
	groups, err := s.GetGroups(ctx, email)
	if err != nil {
		return err
	}

	// Check if group already exists
	for _, g := range groups {
		if g == groupName {
			// Group already exists, nothing to do
			return nil
		}
	}

	// Add the new group
	groups = append(groups, groupName)

	// Marshal and store
	data, err := json.Marshal(groups)
	if err != nil {
		return fmt.Errorf("failed to marshal user groups: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set user groups in cache: %w", err)
	}

	return nil
}

// SetGroups sets the complete list of groups for a user
// This replaces any existing groups
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) SetGroups(ctx context.Context, email string, groups []string) error {
	key := s.userGroupsKey(email)

	data, err := json.Marshal(groups)
	if err != nil {
		return fmt.Errorf("failed to marshal user groups: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to set user groups in cache: %w", err)
	}

	return nil
}

// RemoveGroup removes a specific group from a user's group list
// If this was the last group, the entry is deleted
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) RemoveGroup(ctx context.Context, email, groupName string) error {
	key := s.userGroupsKey(email)

	// Get existing groups
	groups, err := s.GetGroups(ctx, email)
	if err != nil {
		return err
	}

	// If no groups, nothing to remove
	if len(groups) == 0 {
		return nil
	}

	// Find and remove the group
	newGroups := make([]string, 0, len(groups))
	for _, g := range groups {
		if g != groupName {
			newGroups = append(newGroups, g)
		}
	}

	// If no groups left, delete the entry
	if len(newGroups) == 0 {
		return s.cache.Delete(ctx, key)
	}

	// Update with remaining groups
	data, err := json.Marshal(newGroups)
	if err != nil {
		return fmt.Errorf("failed to marshal user groups: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
		return fmt.Errorf("failed to update user groups in cache: %w", err)
	}

	return nil
}

// Delete removes the user's groups entry entirely
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) Delete(ctx context.Context, email string) error {
	key := s.userGroupsKey(email)
	return s.cache.Delete(ctx, key)
}

// Exists checks if a user has any groups in cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserGroupsStore) Exists(ctx context.Context, email string) (bool, error) {
	key := s.userGroupsKey(email)
	_, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}
