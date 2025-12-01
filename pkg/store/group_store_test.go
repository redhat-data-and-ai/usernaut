package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGroupStore(t *testing.T) (*GroupStore, cache.Cache) {
	t.Helper()
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)
	return newGroupStore(c), c
}

func TestGroupStore_Get(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		setup       func(t *testing.T, store *GroupStore, c cache.Cache)
		wantMembers []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "group not found returns empty data",
			groupName: "nonexistent-group",
			setup: func(t *testing.T, store *GroupStore, c cache.Cache) {
				// No setup - group doesn't exist
			},
			wantMembers: []string{},
			wantErr:     false,
		},
		{
			name:      "group found with data",
			groupName: "data-team",
			setup: func(t *testing.T, store *GroupStore, c cache.Cache) {
				err := store.SetMembers(context.Background(), "data-team", []string{"user@example.com"})
				require.NoError(t, err)
			},
			wantMembers: []string{"user@example.com"},
			wantErr:     false,
		},
		{
			name:      "invalid JSON returns error",
			groupName: "invalid-group",
			setup: func(t *testing.T, store *GroupStore, c cache.Cache) {
				err := c.Set(context.Background(), "group:invalid-group", "invalid json{{{", cache.NoExpiration)
				require.NoError(t, err)
			},
			wantMembers: nil,
			wantErr:     true,
			errContains: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, c := setupGroupStore(t)
			tt.setup(t, store, c)

			got, err := store.Get(context.Background(), tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMembers, got.Members)
			}
		})
	}
}

func TestGroupStore_SetAndGet(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set full group data
	data := &GroupData{
		Members: []string{"user1@example.com", "user2@example.com"},
		Backends: map[string]BackendInfo{
			"fivetran_fivetran": {ID: "team_123", Name: "fivetran", Type: "fivetran"},
		},
	}
	err := store.Set(ctx, "data-team", data)
	require.NoError(t, err)

	// Get and verify
	got, err := store.Get(ctx, "data-team")
	require.NoError(t, err)
	assert.Equal(t, data.Members, got.Members)
	assert.Equal(t, data.Backends["fivetran_fivetran"].ID, got.Backends["fivetran_fivetran"].ID)
}

func TestGroupStore_Delete(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		setup     func(t *testing.T, store *GroupStore)
		wantErr   bool
	}{
		{
			name:      "delete existing group",
			groupName: "data-team",
			setup: func(t *testing.T, store *GroupStore) {
				err := store.SetMembers(context.Background(), "data-team", []string{"user@example.com"})
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name:      "delete nonexistent group is no-op",
			groupName: "nonexistent-group",
			setup:     func(t *testing.T, store *GroupStore) {},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupGroupStore(t)
			tt.setup(t, store)

			err := store.Delete(context.Background(), tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				exists, err := store.Exists(context.Background(), tt.groupName)
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestGroupStore_Exists(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		setup     func(t *testing.T, store *GroupStore)
		wantExist bool
	}{
		{
			name:      "exists returns true for existing group",
			groupName: "data-team",
			setup: func(t *testing.T, store *GroupStore) {
				err := store.SetMembers(context.Background(), "data-team", []string{"user@example.com"})
				require.NoError(t, err)
			},
			wantExist: true,
		},
		{
			name:      "exists returns false for nonexistent group",
			groupName: "nonexistent-group",
			setup:     func(t *testing.T, store *GroupStore) {},
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupGroupStore(t)
			tt.setup(t, store)

			exists, err := store.Exists(context.Background(), tt.groupName)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantExist, exists)
		})
	}
}

// Member Operations Tests

func TestGroupStore_GetMembers(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		setup     func(t *testing.T, store *GroupStore)
		want      []string
		wantErr   bool
	}{
		{
			name:      "group not found returns empty slice",
			groupName: "nonexistent-group",
			setup:     func(t *testing.T, store *GroupStore) {},
			want:      []string{},
			wantErr:   false,
		},
		{
			name:      "group found with members",
			groupName: "data-team",
			setup: func(t *testing.T, store *GroupStore) {
				err := store.SetMembers(context.Background(), "data-team", []string{
					"user1@example.com",
					"user2@example.com",
				})
				require.NoError(t, err)
			},
			want:    []string{"user1@example.com", "user2@example.com"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupGroupStore(t)
			tt.setup(t, store)

			got, err := store.GetMembers(context.Background(), tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGroupStore_SetMembers(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// First add a backend
	err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)

	// Set members
	err = store.SetMembers(ctx, "data-team", []string{"user1@example.com", "user2@example.com"})
	require.NoError(t, err)

	// Verify members are set
	members, err := store.GetMembers(ctx, "data-team")
	require.NoError(t, err)
	assert.Equal(t, []string{"user1@example.com", "user2@example.com"}, members)

	// Verify backend is preserved
	backendID, err := store.GetBackendID(ctx, "data-team", "fivetran", "fivetran")
	require.NoError(t, err)
	assert.Equal(t, "team_123", backendID)
}

// Backend Operations Tests

func TestGroupStore_GetBackends(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		setup     func(t *testing.T, store *GroupStore)
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "group not found returns empty map",
			groupName: "nonexistent-group",
			setup:     func(t *testing.T, store *GroupStore) {},
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "group found with backends",
			groupName: "data-team",
			setup: func(t *testing.T, store *GroupStore) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "data-team", "rover", "rover", "team_456")
				require.NoError(t, err)
			},
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupGroupStore(t)
			tt.setup(t, store)

			got, err := store.GetBackends(context.Background(), tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestGroupStore_GetBackendID(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set backend
	err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)

	// Get existing backend ID
	id, err := store.GetBackendID(ctx, "data-team", "fivetran", "fivetran")
	require.NoError(t, err)
	assert.Equal(t, "team_123", id)

	// Get non-existing backend ID
	id, err = store.GetBackendID(ctx, "data-team", "rover", "rover")
	require.NoError(t, err)
	assert.Equal(t, "", id)
}

func TestGroupStore_SetBackend(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set first backend
	err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)

	// Verify
	backends, err := store.GetBackends(ctx, "data-team")
	require.NoError(t, err)
	assert.Equal(t, "team_123", backends["fivetran_fivetran"].ID)
	assert.Equal(t, "fivetran", backends["fivetran_fivetran"].Name)
	assert.Equal(t, "fivetran", backends["fivetran_fivetran"].Type)

	// Set second backend
	err = store.SetBackend(ctx, "data-team", "rover", "rover", "team_456")
	require.NoError(t, err)

	// Verify both exist
	backends, err = store.GetBackends(ctx, "data-team")
	require.NoError(t, err)
	assert.Len(t, backends, 2)
	assert.Equal(t, "team_123", backends["fivetran_fivetran"].ID)
	assert.Equal(t, "team_456", backends["rover_rover"].ID)

	// Update existing backend
	err = store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_789")
	require.NoError(t, err)

	backends, err = store.GetBackends(ctx, "data-team")
	require.NoError(t, err)
	assert.Equal(t, "team_789", backends["fivetran_fivetran"].ID)
}

func TestGroupStore_DeleteBackend(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set two backends
	err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)
	err = store.SetBackend(ctx, "data-team", "rover", "rover", "team_456")
	require.NoError(t, err)

	// Delete one backend
	err = store.DeleteBackend(ctx, "data-team", "fivetran", "fivetran")
	require.NoError(t, err)

	// Verify only one remains
	backends, err := store.GetBackends(ctx, "data-team")
	require.NoError(t, err)
	assert.Len(t, backends, 1)
	assert.Equal(t, "team_456", backends["rover_rover"].ID)
}

func TestGroupStore_BackendExists(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set backend
	err := store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)

	// Check existing backend
	exists, err := store.BackendExists(ctx, "data-team", "fivetran", "fivetran")
	require.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing backend
	exists, err = store.BackendExists(ctx, "data-team", "rover", "rover")
	require.NoError(t, err)
	assert.False(t, exists)

	// Check non-existing group
	exists, err = store.BackendExists(ctx, "nonexistent-group", "fivetran", "fivetran")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGroupStore_KeyPrefix(t *testing.T) {
	store, c := setupGroupStore(t)
	ctx := context.Background()

	// Set data
	err := store.SetMembers(ctx, "data-team", []string{"user@example.com"})
	require.NoError(t, err)

	// Verify key has correct prefix
	val, err := c.Get(ctx, "group:data-team")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Verify key without prefix doesn't exist
	_, err = c.Get(ctx, "data-team")
	assert.Error(t, err)
}

func TestGroupStore_ConsolidatedData(t *testing.T) {
	store, _ := setupGroupStore(t)
	ctx := context.Background()

	// Set members
	err := store.SetMembers(ctx, "data-team", []string{"user1@example.com", "user2@example.com"})
	require.NoError(t, err)

	// Set multiple backends
	err = store.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_123")
	require.NoError(t, err)
	err = store.SetBackend(ctx, "data-team", "rover", "rover", "team_456")
	require.NoError(t, err)
	err = store.SetBackend(ctx, "data-team", "rhplatformtest", "snowflake", "team_789")
	require.NoError(t, err)

	// Get full data and verify consolidation
	data, err := store.Get(ctx, "data-team")
	require.NoError(t, err)

	// Verify members
	assert.Equal(t, []string{"user1@example.com", "user2@example.com"}, data.Members)

	// Verify backends
	assert.Len(t, data.Backends, 3)
	assert.Equal(t, BackendInfo{ID: "team_123", Name: "fivetran", Type: "fivetran"}, data.Backends["fivetran_fivetran"])
	assert.Equal(t, BackendInfo{ID: "team_456", Name: "rover", Type: "rover"}, data.Backends["rover_rover"])
	snowflakeBackend := data.Backends["rhplatformtest_snowflake"]
	assert.Equal(t, BackendInfo{ID: "team_789", Name: "rhplatformtest", Type: "snowflake"}, snowflakeBackend)
}
