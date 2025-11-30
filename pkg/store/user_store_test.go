package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserStore(t *testing.T) (*UserStore, cache.Cache) {
	t.Helper()
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)
	return newUserStore(c), c
}

func TestUserStore_GetBackends(t *testing.T) {
	tests := []GetBackendsTestCase{
		{
			Name:       "user not found returns empty map",
			Identifier: "nonexistent@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				// No setup - user doesn't exist
			},
			Want:    map[string]string{},
			WantErr: false,
		},
		{
			Name:       "user found with single backend",
			Identifier: "user@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			Want: map[string]string{
				"fivetran_prod": "user_123",
			},
			WantErr: false,
		},
		{
			Name:       "user found with multiple backends",
			Identifier: "user@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "user@example.com", "snowflake_prod", "user_456")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "user@example.com", "gitlab_prod", "user_789")
				require.NoError(t, err)
			},
			Want: map[string]string{
				"fivetran_prod":  "user_123",
				"snowflake_prod": "user_456",
				"gitlab_prod":    "user_789",
			},
			WantErr: false,
		},
		{
			Name:       "invalid JSON returns error",
			Identifier: "invalid@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				err := c.Set(context.Background(), "user:invalid@example.com", "invalid json{{{", cache.NoExpiration)
				require.NoError(t, err)
			},
			Want:        nil,
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunGetBackendsTests(t, tests, func() (EntityStoreInterface, cache.Cache) {
		store, c := setupUserStore(t)
		return store, c
	})
}

func TestUserStore_SetBackend(t *testing.T) {
	tests := []SetBackendTestCase{
		{
			Name:       "create new user with backend",
			Identifier: "newuser@example.com",
			BackendKey: "fivetran_prod",
			BackendID:  "user_789",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "newuser@example.com")
				require.NoError(t, err)
				assert.Equal(t, map[string]string{"fivetran_prod": "user_789"}, backends)
			},
			WantErr: false,
		},
		{
			Name:       "update existing backend ID",
			Identifier: "user@example.com",
			BackendKey: "fivetran_prod",
			BackendID:  "user_new_123",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_old_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, "user_new_123", backends["fivetran_prod"])
			},
			WantErr: false,
		},
		{
			Name:       "add second backend to existing user",
			Identifier: "user@example.com",
			BackendKey: "snowflake_prod",
			BackendID:  "user_456",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, 2, len(backends))
				assert.Equal(t, "user_123", backends["fivetran_prod"])
				assert.Equal(t, "user_456", backends["snowflake_prod"])
			},
			WantErr: false,
		},
		{
			Name:       "handles invalid existing JSON",
			Identifier: "corrupt@example.com",
			BackendKey: "fivetran_prod",
			BackendID:  "user_123",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				us := store.(*UserStore)
				err := us.cache.Set(context.Background(), "user:corrupt@example.com", "invalid json", cache.NoExpiration)
				require.NoError(t, err)
			},
			VerifyFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunSetBackendTests(t, tests, func() EntityStoreInterface {
		store, _ := setupUserStore(t)
		return store
	})
}

func TestUserStore_DeleteBackend(t *testing.T) {
	tests := []DeleteBackendTestCase{
		{
			Name:       "delete backend from user with multiple backends",
			Identifier: "user@example.com",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "user@example.com", "snowflake_prod", "user_456")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.Equal(t, 1, len(backends))
				assert.Equal(t, "user_456", backends["snowflake_prod"])
				assert.NotContains(t, backends, "fivetran_prod")
			},
			WantErr: false,
		},
		{
			Name:       "delete last backend removes user entirely",
			Identifier: "user@example.com",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				exists, err := store.Exists(context.Background(), "user@example.com")
				require.NoError(t, err)
				assert.False(t, exists)
			},
			WantErr: false,
		},
		{
			Name:       "delete from nonexistent user is no-op",
			Identifier: "nonexistent@example.com",
			BackendKey: "fivetran_prod",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {},
			WantErr:    false,
		},
		{
			Name:       "handles invalid JSON",
			Identifier: "corrupt@example.com",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				us := store.(*UserStore)
				err := us.cache.Set(context.Background(), "user:corrupt@example.com", "invalid json", cache.NoExpiration)
				require.NoError(t, err)
			},
			VerifyFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunDeleteBackendTests(t, tests, func() EntityStoreInterface {
		store, _ := setupUserStore(t)
		return store
	})
}

func TestUserStore_Delete(t *testing.T) {
	tests := []DeleteTestCase{
		{
			Name:       "delete existing user",
			Identifier: "user@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			WantErr: false,
		},
		{
			Name:       "delete nonexistent user is no-op",
			Identifier: "nonexistent@example.com",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:    false,
		},
	}

	RunDeleteTests(t, tests, func() EntityStoreInterface {
		store, _ := setupUserStore(t)
		return store
	})
}

func TestUserStore_Exists(t *testing.T) {
	tests := []ExistsTestCase{
		{
			Name:       "exists returns true for existing user",
			Identifier: "user@example.com",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			WantExist: true,
		},
		{
			Name:       "exists returns false for nonexistent user",
			Identifier: "nonexistent@example.com",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantExist:  false,
		},
	}

	RunExistsTests(t, tests, func() EntityStoreInterface {
		store, _ := setupUserStore(t)
		return store
	})
}

func TestUserStore_GetByPattern(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		setup        func(t *testing.T, store *UserStore)
		wantEmails   []string
		wantNotFound []string
		wantCount    int
		wantErr      bool
	}{
		{
			name:    "pattern matches domain",
			pattern: "*@example.com",
			setup: func(t *testing.T, store *UserStore) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "alice@example.com", "fivetran_prod", "user_1")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "bob@example.com", "fivetran_prod", "user_2")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "charlie@other.com", "fivetran_prod", "user_3")
				require.NoError(t, err)
			},
			wantEmails:   []string{"alice@example.com", "bob@example.com"},
			wantNotFound: []string{"charlie@other.com"},
			wantCount:    2,
			wantErr:      false,
		},
		{
			name:    "pattern matches username",
			pattern: "*alice*",
			setup: func(t *testing.T, store *UserStore) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "alice@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "bob@example.com", "fivetran_prod", "user_456")
				require.NoError(t, err)
			},
			wantEmails:   []string{"alice@example.com"},
			wantNotFound: []string{"bob@example.com"},
			wantCount:    1,
			wantErr:      false,
		},
		{
			name:    "no matches returns empty map",
			pattern: "*nonexistent*",
			setup: func(t *testing.T, store *UserStore) {
				err := store.SetBackend(context.Background(), "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
			},
			wantEmails:   []string{},
			wantNotFound: []string{"user@example.com"},
			wantCount:    0,
			wantErr:      false,
		},
		{
			name:    "skips invalid JSON entries",
			pattern: "*@example.com",
			setup: func(t *testing.T, store *UserStore) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "valid@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
				// Add invalid JSON
				err = store.cache.Set(ctx, "user:invalid@example.com", "invalid json", cache.NoExpiration)
				require.NoError(t, err)
			},
			wantEmails:   []string{"valid@example.com"},
			wantNotFound: []string{"invalid@example.com"},
			wantCount:    1,
			wantErr:      false,
		},
		{
			name:    "returns backend data correctly",
			pattern: "*user@example.com",
			setup: func(t *testing.T, store *UserStore) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "user@example.com", "fivetran_prod", "user_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "user@example.com", "snowflake_prod", "user_456")
				require.NoError(t, err)
			},
			wantEmails: []string{"user@example.com"},
			wantCount:  1,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupUserStore(t)
			tt.setup(t, store)

			got, err := store.GetByPattern(context.Background(), tt.pattern)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantCount, len(got))

				for _, email := range tt.wantEmails {
					assert.Contains(t, got, email)
				}
				for _, email := range tt.wantNotFound {
					assert.NotContains(t, got, email)
				}
			}
		})
	}
}

func TestUserStore_KeyPrefix(t *testing.T) {
	store, c := setupUserStore(t)
	ctx := context.Background()

	// Set a backend
	err := store.SetBackend(ctx, "user@example.com", "fivetran_prod", "user_123")
	require.NoError(t, err)

	// Verify key has correct prefix
	val, err := c.Get(ctx, "user:user@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Verify key without prefix doesn't exist
	_, err = c.Get(ctx, "user@example.com")
	assert.Error(t, err)
}
