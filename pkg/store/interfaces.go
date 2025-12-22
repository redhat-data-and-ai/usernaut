package store

import "context"

// UserStoreInterface defines operations for user-related cache operations
// This interface enables mocking in tests and follows the dependency inversion principle
type UserStoreInterface interface {
	// GetBackends returns a map of backend IDs for a user
	// Returns an empty map if the user is not found in cache
	// Map format: {"backend_name_type": "backend_user_id"}
	GetBackends(ctx context.Context, email string) (map[string]string, error)

	// SetBackend sets a backend ID for a user
	// If the user doesn't exist, it will be created
	// If the user exists, the backend ID will be added/updated in the map
	SetBackend(ctx context.Context, email, backendKey, backendID string) error

	// DeleteBackend removes a specific backend ID from a user's record
	// If this was the last backend, the entire user entry is deleted
	DeleteBackend(ctx context.Context, email, backendKey string) error

	// Delete removes a user entirely from cache
	Delete(ctx context.Context, email string) error

	// Exists checks if a user exists in cache
	Exists(ctx context.Context, email string) (bool, error)

	// GetByPattern searches for users matching a pattern and returns their data
	// Pattern should NOT include the "user:" prefix - it will be added automatically
	// Example: pattern "*@example.com" searches for "user:*@example.com"
	// Returns: map[email]backends where backends is map[backendKey]backendID
	GetByPattern(ctx context.Context, pattern string) (map[string]map[string]string, error)
}

// TeamStoreInterface defines operations for team-related cache operations
// Key format: "team:<transformedTeamName>"
// This store is used for preloading team data from backends where teams
// are identified by their transformed names.
type TeamStoreInterface interface {
	// GetBackends returns a map of backend IDs for a team
	// Returns an empty map if the team is not found in cache
	// Map format: {"backend_name_type": "backend_team_id"}
	GetBackends(ctx context.Context, teamName string) (map[string]string, error)

	// SetBackend sets a backend ID for a team
	// If the team doesn't exist, it will be created
	// If the team exists, the backend ID will be added/updated in the map
	SetBackend(ctx context.Context, teamName, backendKey, teamID string) error

	// DeleteBackend removes a specific backend ID from a team's record
	// If this was the last backend, the entire team entry is deleted
	DeleteBackend(ctx context.Context, teamName, backendKey string) error

	// Delete removes a team entirely from cache
	Delete(ctx context.Context, teamName string) error

	// Exists checks if a team exists in cache
	Exists(ctx context.Context, teamName string) (bool, error)
}

// GroupStoreInterface defines operations for consolidated group cache operations
// Key format: "group:<groupName>"
// Value: JSON object containing members and backends
type GroupStoreInterface interface {
	// Get retrieves the full group data from cache
	// Returns empty GroupData if the group is not found in cache
	Get(ctx context.Context, groupName string) (*GroupData, error)

	// Set stores the full group data in cache
	Set(ctx context.Context, groupName string, data *GroupData) error

	// Delete removes a group entirely from cache
	Delete(ctx context.Context, groupName string) error

	// Exists checks if a group exists in cache
	Exists(ctx context.Context, groupName string) (bool, error)

	// --- Member Operations ---

	// GetMembers returns the list of user emails for a group
	// Returns an empty slice if the group is not found in cache
	GetMembers(ctx context.Context, groupName string) ([]string, error)

	// SetMembers sets the complete list of user emails for a group
	// This replaces any existing members while preserving backends
	SetMembers(ctx context.Context, groupName string, members []string) error

	// --- Backend Operations ---

	// GetBackends returns a map of backend info for a group
	// Returns an empty map if the group is not found in cache
	// Map format: {"backend_name_type": BackendInfo{ID, Name, Type}}
	GetBackends(ctx context.Context, groupName string) (map[string]BackendInfo, error)

	// GetBackendID returns the backend ID for a specific backend
	// Returns empty string if the backend is not found
	GetBackendID(ctx context.Context, groupName, backendName, backendType string) (string, error)

	// SetBackend sets a backend for a group
	// If the group doesn't exist, it will be created
	// If the backend exists, it will be updated
	SetBackend(ctx context.Context, groupName, backendName, backendType, backendID string) error

	// DeleteBackend removes a specific backend from a group's record
	DeleteBackend(ctx context.Context, groupName, backendName, backendType string) error

	// BackendExists checks if a specific backend exists for a group
	BackendExists(ctx context.Context, groupName, backendName, backendType string) (bool, error)
}

// UserGroupsStoreInterface defines operations for user-to-groups reverse index
// Key format: "user:groups:<email>"
type UserGroupsStoreInterface interface {
	// GetGroups returns the list of groups for a user
	// Returns an empty slice if the user is not found in cache
	GetGroups(ctx context.Context, email string) ([]string, error)

	// AddGroup adds a group to a user's group list if not already present
	AddGroup(ctx context.Context, email, groupName string) error

	// SetGroups sets the complete list of groups for a user
	// This replaces any existing groups
	SetGroups(ctx context.Context, email string, groups []string) error

	// RemoveGroup removes a specific group from a user's group list
	// If this was the last group, the entry is deleted
	RemoveGroup(ctx context.Context, email, groupName string) error

	// Delete removes the user's groups entry entirely
	Delete(ctx context.Context, email string) error

	// Exists checks if a user has any groups in cache
	Exists(ctx context.Context, email string) (bool, error)
}

// StoreInterface is the main interface that combines all store operations
// This is the primary interface that should be used by consumers
type StoreInterface interface {
	// GetUserStore returns the user store operations
	GetUserStore() UserStoreInterface

	// GetTeamStore returns the team store operations (for preload with transformed names)
	GetTeamStore() TeamStoreInterface

	// GetGroupStore returns the group store operations (for reconciliation with original names)
	GetGroupStore() GroupStoreInterface

	// GetUserGroupsStore returns the user groups store operations
	GetUserGroupsStore() UserGroupsStoreInterface
}
