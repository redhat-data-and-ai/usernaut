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

// MetaStoreInterface defines operations for metadata cache operations
type MetaStoreInterface interface {
	// GetUserList retrieves the list of user IDs from cache
	// Returns empty slice if not found
	GetUserList(ctx context.Context) ([]string, error)

	// SetUserList stores the list of user IDs in cache
	SetUserList(ctx context.Context, users []string) error
}

// StoreInterface is the main interface that combines all store operations
// This is the primary interface that should be used by consumers
type StoreInterface interface {
	// GetUserStore returns the user store operations
	GetUserStore() UserStoreInterface

	// GetTeamStore returns the team store operations
	GetTeamStore() TeamStoreInterface

	// GetMetaStore returns the metadata store operations
	GetMetaStore() MetaStoreInterface
}
