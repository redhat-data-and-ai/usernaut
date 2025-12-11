package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

func TestNew(t *testing.T) {
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)

	store := New(c)

	// Verify all sub-stores are initialized
	assert.NotNil(t, store)
	assert.NotNil(t, store.User)
	assert.NotNil(t, store.Team)
	assert.NotNil(t, store.Group)
	assert.NotNil(t, store.Meta)
	assert.NotNil(t, store.UserGroups)
}

func TestStore_InterfaceCompliance(t *testing.T) {
	// This test verifies that our concrete types implement the interfaces
	// The compile-time checks in store.go should catch this, but this test
	// provides runtime verification and documentation

	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)

	store := New(c)

	// Verify User implements UserStoreInterface
	var _ UserStoreInterface = store.User

	// Verify Team implements TeamStoreInterface
	var _ TeamStoreInterface = store.Team

	// Verify Group implements GroupStoreInterface
	var _ GroupStoreInterface = store.Group

	// Verify Meta implements MetaStoreInterface
	var _ MetaStoreInterface = store.Meta

	// Verify UserGroups implements UserGroupsStoreInterface
	var _ UserGroupsStoreInterface = store.UserGroups
}

func TestStore_IndependentOperations(t *testing.T) {
	// Test that User, Group, Meta, and UserGroups operations don't interfere with each other
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)

	store := New(c)
	ctx := testContext(t)

	// Create user
	err = store.User.SetBackend(ctx, "user@example.com", "fivetran_prod", "user_123")
	require.NoError(t, err)

	// Create group with backend
	err = store.Group.SetBackend(ctx, "data-team", "fivetran", "fivetran", "team_456")
	require.NoError(t, err)

	// Create meta
	err = store.Meta.SetUserList(ctx, []string{"user1", "user2"})
	require.NoError(t, err)

	// Create user groups
	err = store.UserGroups.AddGroup(ctx, "user@example.com", "data-team")
	require.NoError(t, err)

	// Verify all exist independently
	userBackends, err := store.User.GetBackends(ctx, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user_123", userBackends["fivetran_prod"])

	groupBackends, err := store.Group.GetBackends(ctx, "data-team")
	require.NoError(t, err)
	assert.Equal(t, "team_456", groupBackends["fivetran_fivetran"].ID)

	userList, err := store.Meta.GetUserList(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"user1", "user2"}, userList)

	userGroups, err := store.UserGroups.GetGroups(ctx, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"data-team"}, userGroups)
}

func TestStore_KeyNamespacing(t *testing.T) {
	// Test that key prefixing prevents collisions
	c, err := inmemory.NewCache(&inmemory.Config{
		DefaultExpiration: 300,
		CleanupInterval:   600,
	})
	require.NoError(t, err)

	store := New(c)
	ctx := testContext(t)

	// Use the same name for user email and group name
	sameName := "test@example.com"

	err = store.User.SetBackend(ctx, sameName, "backend1", "id1")
	require.NoError(t, err)

	err = store.Group.SetBackend(ctx, sameName, "backend1", "backend1", "id2")
	require.NoError(t, err)

	// Verify they don't collide
	userBackends, err := store.User.GetBackends(ctx, sameName)
	require.NoError(t, err)
	assert.Equal(t, "id1", userBackends["backend1"])

	groupBackends, err := store.Group.GetBackends(ctx, sameName)
	require.NoError(t, err)
	assert.Equal(t, "id2", groupBackends["backend1_backend1"].ID)
}
