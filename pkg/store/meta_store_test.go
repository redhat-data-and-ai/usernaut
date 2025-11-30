package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMetaStore(t *testing.T) (*MetaStore, cache.Cache) {
	t.Helper()
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)
	return newMetaStore(c), c
}

func TestMetaStore_GetUserList(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, store *MetaStore)
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "user list not found returns empty slice",
			setup:   func(t *testing.T, store *MetaStore) {},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "user list found with single user",
			setup: func(t *testing.T, store *MetaStore) {
				err := store.SetUserList(context.Background(), []string{"user1"})
				require.NoError(t, err)
			},
			want:    []string{"user1"},
			wantErr: false,
		},
		{
			name: "user list found with multiple users",
			setup: func(t *testing.T, store *MetaStore) {
				err := store.SetUserList(context.Background(), []string{"user1", "user2", "user3"})
				require.NoError(t, err)
			},
			want:    []string{"user1", "user2", "user3"},
			wantErr: false,
		},
		{
			name: "empty user list returns empty slice",
			setup: func(t *testing.T, store *MetaStore) {
				err := store.SetUserList(context.Background(), []string{})
				require.NoError(t, err)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "invalid JSON returns error",
			setup: func(t *testing.T, store *MetaStore) {
				err := store.cache.Set(context.Background(), "meta:user_list", "invalid json{{{", cache.NoExpiration)
				require.NoError(t, err)
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupMetaStore(t)
			tt.setup(t, store)

			got, err := store.GetUserList(context.Background())

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

func TestMetaStore_SetUserList(t *testing.T) {
	tests := []struct {
		name    string
		users   []string
		verify  func(t *testing.T, store *MetaStore)
		wantErr bool
	}{
		{
			name:  "set user list with single user",
			users: []string{"user1"},
			verify: func(t *testing.T, store *MetaStore) {
				users, err := store.GetUserList(context.Background())
				require.NoError(t, err)
				assert.Equal(t, []string{"user1"}, users)
			},
			wantErr: false,
		},
		{
			name:  "set user list with multiple users",
			users: []string{"user1", "user2", "user3", "user4"},
			verify: func(t *testing.T, store *MetaStore) {
				users, err := store.GetUserList(context.Background())
				require.NoError(t, err)
				assert.Equal(t, []string{"user1", "user2", "user3", "user4"}, users)
			},
			wantErr: false,
		},
		{
			name:  "set empty user list",
			users: []string{},
			verify: func(t *testing.T, store *MetaStore) {
				users, err := store.GetUserList(context.Background())
				require.NoError(t, err)
				assert.Equal(t, []string{}, users)
			},
			wantErr: false,
		},
		{
			name:  "set nil user list stores empty list",
			users: nil,
			verify: func(t *testing.T, store *MetaStore) {
				users, err := store.GetUserList(context.Background())
				require.NoError(t, err)
				// nil marshals to null in JSON, but we should handle it gracefully
				assert.NotNil(t, users)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupMetaStore(t)

			err := store.SetUserList(context.Background(), tt.users)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, store)
			}
		})
	}
}

func TestMetaStore_UpdateUserList(t *testing.T) {
	tests := []struct {
		name      string
		initial   []string
		update    []string
		wantFinal []string
	}{
		{
			name:      "update empty list",
			initial:   []string{},
			update:    []string{"user1", "user2"},
			wantFinal: []string{"user1", "user2"},
		},
		{
			name:      "update existing list",
			initial:   []string{"user1", "user2"},
			update:    []string{"user3", "user4"},
			wantFinal: []string{"user3", "user4"},
		},
		{
			name:      "update to empty list",
			initial:   []string{"user1", "user2", "user3"},
			update:    []string{},
			wantFinal: []string{},
		},
		{
			name:      "update with overlapping users",
			initial:   []string{"user1", "user2"},
			update:    []string{"user2", "user3"},
			wantFinal: []string{"user2", "user3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupMetaStore(t)
			ctx := context.Background()

			// Set initial state
			err := store.SetUserList(ctx, tt.initial)
			require.NoError(t, err)

			// Update
			err = store.SetUserList(ctx, tt.update)
			require.NoError(t, err)

			// Verify final state
			got, err := store.GetUserList(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFinal, got)
		})
	}
}

func TestMetaStore_KeyPrefix(t *testing.T) {
	store, c := setupMetaStore(t)
	ctx := context.Background()

	// Set user list
	err := store.SetUserList(ctx, []string{"user1", "user2"})
	require.NoError(t, err)

	// Verify key has correct prefix
	val, err := c.Get(ctx, "meta:user_list")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Verify key without prefix doesn't exist
	_, err = c.Get(ctx, "user_list")
	assert.Error(t, err)
}

func TestMetaStore_LargeUserList(t *testing.T) {
	store, _ := setupMetaStore(t)
	ctx := context.Background()

	// Create a large user list
	largeList := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeList[i] = "user" + string(rune(i))
	}

	// Set the large list
	err := store.SetUserList(ctx, largeList)
	require.NoError(t, err)

	// Retrieve and verify
	got, err := store.GetUserList(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1000, len(got))
	assert.Equal(t, largeList, got)
}
