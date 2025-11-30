package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// deleteBackendHelper is a shared helper function for deleting a backend from an entity's record
// It handles the common logic of:
// 1. Getting the entity from cache
// 2. Removing the specified backend
// 3. Either updating the remaining backends or deleting the entire entry if no backends remain
//
// Parameters:
//   - ctx: context for cache operations
//   - c: the cache instance
//   - key: the full cache key (with prefix already applied)
//   - backendKey: the backend identifier to remove
//   - entityType: the entity type name (e.g., "user", "team") for error messages
func deleteBackendHelper(ctx context.Context, c cache.Cache, key, backendKey, entityType string) error {
	// Get existing backends
	backends := make(map[string]string)
	val, err := c.Get(ctx, key)
	if err != nil {
		// Entity doesn't exist, nothing to delete
		return nil
	}

	if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
		return fmt.Errorf("failed to unmarshal %s backends: %w", entityType, err)
	}

	// Delete the backend
	delete(backends, backendKey)

	// If there are remaining backends, update the cache
	if len(backends) > 0 {
		data, err := json.Marshal(backends)
		if err != nil {
			return fmt.Errorf("failed to marshal %s backends: %w", entityType, err)
		}

		if err := c.Set(ctx, key, string(data), cache.NoExpiration); err != nil {
			return fmt.Errorf("failed to update %s in cache: %w", entityType, err)
		}
	} else {
		// No backends left, delete the entire entry
		if err := c.Delete(ctx, key); err != nil {
			return fmt.Errorf("failed to delete %s from cache: %w", entityType, err)
		}
	}

	return nil
}
