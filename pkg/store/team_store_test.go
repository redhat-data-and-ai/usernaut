package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTeamStore(t *testing.T) (*TeamStore, cache.Cache) {
	t.Helper()
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)
	return newTeamStore(c), c
}

func TestTeamStore_GetBackends(t *testing.T) {
	tests := []GetBackendsTestCase{
		{
			Name:       "team not found returns empty map",
			Identifier: "nonexistent-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				// No setup - team doesn't exist
			},
			Want:    map[string]string{},
			WantErr: false,
		},
		{
			Name:       "team found with single backend",
			Identifier: "data-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				err := store.SetBackend(context.Background(), "data-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
			},
			Want: map[string]string{
				"fivetran_prod": "team_123",
			},
			WantErr: false,
		},
		{
			Name:       "team found with multiple backends",
			Identifier: "data-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "data-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "data-team", "snowflake_prod", "team_456")
				require.NoError(t, err)
			},
			Want: map[string]string{
				"fivetran_prod":  "team_123",
				"snowflake_prod": "team_456",
			},
			WantErr: false,
		},
		{
			Name:       "invalid JSON returns error",
			Identifier: "invalid-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface, c cache.Cache) {
				err := c.Set(context.Background(), "team:invalid-team", "invalid json{{{", cache.NoExpiration)
				require.NoError(t, err)
			},
			Want:        nil,
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunGetBackendsTests(t, tests, func() (EntityStoreInterface, cache.Cache) {
		store, c := setupTeamStore(t)
		return store, c
	})
}

func TestTeamStore_SetBackend(t *testing.T) {
	tests := []SetBackendTestCase{
		{
			Name:       "create new team with backend",
			Identifier: "new-team",
			BackendKey: "fivetran_prod",
			BackendID:  "team_789",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "new-team")
				require.NoError(t, err)
				assert.Equal(t, map[string]string{"fivetran_prod": "team_789"}, backends)
			},
			WantErr: false,
		},
		{
			Name:       "update existing backend ID",
			Identifier: "existing-team",
			BackendKey: "fivetran_prod",
			BackendID:  "team_new_123",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "existing-team", "fivetran_prod", "team_old_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "existing-team")
				require.NoError(t, err)
				assert.Equal(t, "team_new_123", backends["fivetran_prod"])
			},
			WantErr: false,
		},
		{
			Name:       "add second backend to existing team",
			Identifier: "multi-backend-team",
			BackendKey: "snowflake_prod",
			BackendID:  "team_456",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "multi-backend-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "multi-backend-team")
				require.NoError(t, err)
				assert.Equal(t, 2, len(backends))
				assert.Equal(t, "team_123", backends["fivetran_prod"])
				assert.Equal(t, "team_456", backends["snowflake_prod"])
			},
			WantErr: false,
		},
		{
			Name:       "handles invalid existing JSON",
			Identifier: "corrupt-team",
			BackendKey: "fivetran_prod",
			BackendID:  "team_123",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				ts := store.(*TeamStore)
				err := ts.cache.Set(context.Background(), "team:corrupt-team", "invalid json", cache.NoExpiration)
				require.NoError(t, err)
			},
			VerifyFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunSetBackendTests(t, tests, func() EntityStoreInterface {
		store, _ := setupTeamStore(t)
		return store
	})
}

func TestTeamStore_DeleteBackend(t *testing.T) {
	tests := []DeleteBackendTestCase{
		{
			Name:       "delete backend from team with multiple backends",
			Identifier: "multi-team",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				ctx := context.Background()
				err := store.SetBackend(ctx, "multi-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
				err = store.SetBackend(ctx, "multi-team", "snowflake_prod", "team_456")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				backends, err := store.GetBackends(context.Background(), "multi-team")
				require.NoError(t, err)
				assert.Equal(t, 1, len(backends))
				assert.Equal(t, "team_456", backends["snowflake_prod"])
				assert.NotContains(t, backends, "fivetran_prod")
			},
			WantErr: false,
		},
		{
			Name:       "delete last backend removes team entirely",
			Identifier: "single-backend-team",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "single-backend-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
			},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {
				exists, err := store.Exists(context.Background(), "single-backend-team")
				require.NoError(t, err)
				assert.False(t, exists)
			},
			WantErr: false,
		},
		{
			Name:       "delete from nonexistent team is no-op",
			Identifier: "nonexistent-team",
			BackendKey: "fivetran_prod",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			VerifyFunc: func(t *testing.T, store EntityStoreInterface) {},
			WantErr:    false,
		},
		{
			Name:       "handles invalid JSON",
			Identifier: "corrupt-team",
			BackendKey: "fivetran_prod",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				ts := store.(*TeamStore)
				err := ts.cache.Set(context.Background(), "team:corrupt-team", "invalid json", cache.NoExpiration)
				require.NoError(t, err)
			},
			VerifyFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:     true,
			ErrContains: "failed to unmarshal",
		},
	}

	RunDeleteBackendTests(t, tests, func() EntityStoreInterface {
		store, _ := setupTeamStore(t)
		return store
	})
}

func TestTeamStore_Delete(t *testing.T) {
	tests := []DeleteTestCase{
		{
			Name:       "delete existing team",
			Identifier: "data-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "data-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
			},
			WantErr: false,
		},
		{
			Name:       "delete nonexistent team is no-op",
			Identifier: "nonexistent-team",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantErr:    false,
		},
	}

	RunDeleteTests(t, tests, func() EntityStoreInterface {
		store, _ := setupTeamStore(t)
		return store
	})
}

func TestTeamStore_Exists(t *testing.T) {
	tests := []ExistsTestCase{
		{
			Name:       "exists returns true for existing team",
			Identifier: "data-team",
			SetupFunc: func(t *testing.T, store EntityStoreInterface) {
				err := store.SetBackend(context.Background(), "data-team", "fivetran_prod", "team_123")
				require.NoError(t, err)
			},
			WantExist: true,
		},
		{
			Name:       "exists returns false for nonexistent team",
			Identifier: "nonexistent-team",
			SetupFunc:  func(t *testing.T, store EntityStoreInterface) {},
			WantExist:  false,
		},
	}

	RunExistsTests(t, tests, func() EntityStoreInterface {
		store, _ := setupTeamStore(t)
		return store
	})
}

func TestTeamStore_KeyPrefix(t *testing.T) {
	store, c := setupTeamStore(t)
	ctx := context.Background()

	// Set a backend
	err := store.SetBackend(ctx, "data-team", "fivetran_prod", "team_123")
	require.NoError(t, err)

	// Verify key has correct prefix
	val, err := c.Get(ctx, "team:data-team")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Verify key without prefix doesn't exist
	_, err = c.Get(ctx, "data-team")
	assert.Error(t, err)
}
