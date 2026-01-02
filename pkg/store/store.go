package store

import (
	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
)

// Store provides a high-level interface for managing users, teams, groups, and metadata in cache
// It encapsulates key prefixing and JSON serialization
// NOTE: This store does NOT handle locking - callers are responsible for proper synchronization
type Store struct {
	User       UserStoreInterface
	Team       TeamStoreInterface  // For preload with transformed team names
	Group      GroupStoreInterface // For reconciliation with original group names
	UserGroups UserGroupsStoreInterface
}

// New creates a new Store instance with all sub-stores initialized
func New(cache cache.Cache) *Store {
	return &Store{
		User:       newUserStore(cache),
		Team:       newTeamStore(cache),
		Group:      newGroupStore(cache),
		UserGroups: newUserGroupsStore(cache),
	}
}

// Compile-time interface compliance checks
var (
	_ UserStoreInterface       = (*UserStore)(nil)
	_ TeamStoreInterface       = (*TeamStore)(nil)
	_ GroupStoreInterface      = (*GroupStore)(nil)
	_ UserGroupsStoreInterface = (*UserGroupsStore)(nil)
)
