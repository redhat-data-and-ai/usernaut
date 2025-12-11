package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserGroupsStore(t *testing.T) (*UserGroupsStore, cache.Cache) {
	t.Helper()
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)
	return newUserGroupsStore(c), c
}

func TestUserGroupsStore_GetGroups(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		setup       func(t *testing.T, store *UserGroupsStore, c cache.Cache)
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:  "user not found returns empty slice",
			email: "nonexistent@example.com",
			setup: func(t *testing.T, store *UserGroupsStore, c cache.Cache) {
				// No setup - user doesn't exist
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "user found with single group",
			email: "user@example.com",
			setup: func(t *testing.T, store *UserGroupsStore, c cache.Cache) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			want:    []string{"data-team"},
			wantErr: false,
		},
		{
			name:  "user found with multiple groups",
			email: "user@example.com",
			setup: func(t *testing.T, store *UserGroupsStore, c cache.Cache) {
				ctx := context.Background()
				err := store.AddGroup(ctx, "user@example.com", "data-team")
				require.NoError(t, err)
				err = store.AddGroup(ctx, "user@example.com", "platform-team")
				require.NoError(t, err)
				err = store.AddGroup(ctx, "user@example.com", "ml-team")
				require.NoError(t, err)
			},
			want:    []string{"data-team", "platform-team", "ml-team"},
			wantErr: false,
		},
		{
			name:  "invalid JSON returns error",
			email: "invalid@example.com",
			setup: func(t *testing.T, store *UserGroupsStore, c cache.Cache) {
				err := c.Set(context.Background(), "user:groups:invalid@example.com", "invalid json{{{", cache.NoExpiration)
				require.NoError(t, err)
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, c := setupUserGroupsStore(t)
			tt.setup(t, store, c)

			got, err := store.GetGroups(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUserGroupsStore_AddGroup(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		groupName string
		setup     func(t *testing.T, store *UserGroupsStore)
		verify    func(t *testing.T, store *UserGroupsStore)
		wantErr   bool
	}{
		{
			name:      "add group to new user",
			email:     "newuser@example.com",
			groupName: "data-team",
			setup:     func(t *testing.T, store *UserGroupsStore) {},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "newuser@example.com")
				require.NoError(t, err)
				assert.Equal(t, []string{"data-team"}, groups)
			},
			wantErr: false,
		},
		{
			name:      "add second group to existing user",
			email:     "user@example.com",
			groupName: "platform-team",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, 2, len(groups))
				assert.Contains(t, groups, "data-team")
				assert.Contains(t, groups, "platform-team")
			},
			wantErr: false,
		},
		{
			name:      "adding same group twice is idempotent",
			email:     "user@example.com",
			groupName: "data-team",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				// Should still only have one instance of the group
				assert.Equal(t, []string{"data-team"}, groups)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserGroupsStore(t)
			tt.setup(t, store)

			err := store.AddGroup(context.Background(), tt.email, tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, store)
			}
		})
	}
}

func TestUserGroupsStore_SetGroups(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		groups  []string
		setup   func(t *testing.T, store *UserGroupsStore)
		verify  func(t *testing.T, store *UserGroupsStore)
		wantErr bool
	}{
		{
			name:   "set groups for new user",
			email:  "newuser@example.com",
			groups: []string{"data-team", "platform-team"},
			setup:  func(t *testing.T, store *UserGroupsStore) {},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "newuser@example.com")
				require.NoError(t, err)
				assert.Equal(t, []string{"data-team", "platform-team"}, groups)
			},
			wantErr: false,
		},
		{
			name:   "set groups replaces existing groups",
			email:  "user@example.com",
			groups: []string{"new-team"},
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.SetGroups(context.Background(), "user@example.com", []string{"old-team-1", "old-team-2"})
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, []string{"new-team"}, groups)
			},
			wantErr: false,
		},
		{
			name:   "set empty groups list",
			email:  "user@example.com",
			groups: []string{},
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.SetGroups(context.Background(), "user@example.com", []string{"data-team"})
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, []string{}, groups)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserGroupsStore(t)
			tt.setup(t, store)

			err := store.SetGroups(context.Background(), tt.email, tt.groups)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, store)
			}
		})
	}
}

func TestUserGroupsStore_RemoveGroup(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		groupName string
		setup     func(t *testing.T, store *UserGroupsStore)
		verify    func(t *testing.T, store *UserGroupsStore)
		wantErr   bool
	}{
		{
			name:      "remove group from user with multiple groups",
			email:     "user@example.com",
			groupName: "data-team",
			setup: func(t *testing.T, store *UserGroupsStore) {
				ctx := context.Background()
				err := store.AddGroup(ctx, "user@example.com", "data-team")
				require.NoError(t, err)
				err = store.AddGroup(ctx, "user@example.com", "platform-team")
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, 1, len(groups))
				assert.Equal(t, "platform-team", groups[0])
				assert.NotContains(t, groups, "data-team")
			},
			wantErr: false,
		},
		{
			name:      "remove last group deletes entry",
			email:     "user@example.com",
			groupName: "data-team",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				exists, err := store.Exists(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.False(t, exists)
			},
			wantErr: false,
		},
		{
			name:      "remove from nonexistent user is no-op",
			email:     "nonexistent@example.com",
			groupName: "data-team",
			setup:     func(t *testing.T, store *UserGroupsStore) {},
			verify:    func(t *testing.T, store *UserGroupsStore) {},
			wantErr:   false,
		},
		{
			name:      "remove nonexistent group is no-op",
			email:     "user@example.com",
			groupName: "nonexistent-team",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *UserGroupsStore) {
				groups, err := store.GetGroups(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, []string{"data-team"}, groups)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserGroupsStore(t)
			tt.setup(t, store)

			err := store.RemoveGroup(context.Background(), tt.email, tt.groupName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, store)
			}
		})
	}
}

func TestUserGroupsStore_Delete(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		setup   func(t *testing.T, store *UserGroupsStore)
		wantErr bool
	}{
		{
			name:  "delete existing user",
			email: "user@example.com",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name:    "delete nonexistent user is no-op",
			email:   "nonexistent@example.com",
			setup:   func(t *testing.T, store *UserGroupsStore) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserGroupsStore(t)
			tt.setup(t, store)

			err := store.Delete(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify deletion
				exists, err := store.Exists(context.Background(), tt.email)
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestUserGroupsStore_Exists(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		setup     func(t *testing.T, store *UserGroupsStore)
		wantExist bool
	}{
		{
			name:  "exists returns true for existing user",
			email: "user@example.com",
			setup: func(t *testing.T, store *UserGroupsStore) {
				err := store.AddGroup(context.Background(), "user@example.com", "data-team")
				require.NoError(t, err)
			},
			wantExist: true,
		},
		{
			name:      "exists returns false for nonexistent user",
			email:     "nonexistent@example.com",
			setup:     func(t *testing.T, store *UserGroupsStore) {},
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserGroupsStore(t)
			tt.setup(t, store)

			exists, err := store.Exists(context.Background(), tt.email)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantExist, exists)
		})
	}
}

func TestUserGroupsStore_KeyPrefix(t *testing.T) {
	store, c := setupUserGroupsStore(t)
	ctx := context.Background()

	// Add a group
	err := store.AddGroup(ctx, "user@example.com", "data-team")
	require.NoError(t, err)

	// Verify key has correct prefix
	val, err := c.Get(ctx, "user:groups:user@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Verify key without prefix doesn't exist
	_, err = c.Get(ctx, "user@example.com")
	assert.Error(t, err)
}
