package controller

import (
	"context"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

const presetBackendKey = "preset_preset"

func newPresetReconciler(t *testing.T) *GroupReconciler {
	t.Helper()
	return newBackendReconciler(t, "preset", "preset", false)
}

func newBackendReconciler(t *testing.T, backendName, backendType string, searchUserByEmail bool) *GroupReconciler {
	t.Helper()
	cacheClient, err := inmemory.NewCache(&inmemory.Config{DefaultExpiration: -1, CleanupInterval: -1})
	require.NoError(t, err)
	connection := map[string]interface{}{}
	if searchUserByEmail {
		connection["search_user_by_email"] = true
	}
	return &GroupReconciler{
		Store:      store.New(cacheClient),
		CacheMutex: &sync.RWMutex{},
		AppConfig: &config.AppConfig{
			BackendMap: map[string]map[string]config.Backend{
				backendType: {
					backendName: {Name: backendName, Type: backendType, Connection: connection},
				},
			},
		},
	}
}

func presetLDAPUser() map[string]*structs.LDAPUser {
	return map[string]*structs.LDAPUser{
		"ruagrawa": {Email: "ruagrawa@redhat.com", UID: "ruagrawa"},
	}
}

// plainPresetBackend does not implement email lookup.
type plainPresetBackend struct {
	clients.Client
	created []*structs.User
}

func (b *plainPresetBackend) CreateUser(_ context.Context, user *structs.User) (*structs.User, error) {
	b.created = append(b.created, user)
	return &structs.User{ID: "created-id", Email: user.Email, UserName: user.UserName}, nil
}

// emailPresetBackend implements clients.EmailUserLookup.
type emailPresetBackend struct {
	plainPresetBackend
	found *structs.User
	calls int
}

func (b *emailPresetBackend) FindUserByEmail(_ context.Context, _ string) (*structs.User, error) {
	b.calls++
	return b.found, nil
}

func TestProcessUsers_staleIDNotOnGroupIsReplacedAndAdded(t *testing.T) {
	r := newBackendReconciler(t, "preset", "preset", true)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &emailPresetBackend{found: &structs.User{ID: "oidc|new", Email: "ruagrawa@redhat.com"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"oidc|new"}, add)
	assert.Empty(t, remove)
	assert.Empty(t, backend.created)
	assert.Equal(t, 1, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_staleIDAlreadyOnGroupUpdatesCacheOnly(t *testing.T) {
	r := newBackendReconciler(t, "preset", "preset", true)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{"oidc|new": {ID: "oidc|new"}}
	backend := &emailPresetBackend{found: &structs.User{ID: "oidc|new", Email: "ruagrawa@redhat.com"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_removedMemberWithNewSCIMIDIsRemoved(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{"oidc|new": {ID: "oidc|new"}}
	backend := &emailPresetBackend{found: &structs.User{ID: "oidc|new"}}
	add, remove, err := r.processUsers(ctx, nil, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"oidc|new"}, remove)
	assert.Equal(t, 0, backend.calls)
}

func TestProcessUsers_emailMissDoesNotCreate(t *testing.T) {
	r := newBackendReconciler(t, "preset", "preset", true)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{"oidc|new": {ID: "oidc|new"}}
	backend := &emailPresetBackend{}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, backend.created)
	assert.Equal(t, []string{"oidc|new"}, remove)
}

func TestProcessUsers_flagOffSkipsEmailLookup(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{"oidc|new": {ID: "oidc|new"}}
	backend := &emailPresetBackend{found: &structs.User{ID: "oidc|new"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, add)
	assert.Equal(t, []string{"oidc|new"}, remove)
	assert.Equal(t, 0, backend.calls)
}

func TestProcessUsers_flagOnWithoutLookupErrors(t *testing.T) {
	r := newBackendReconciler(t, "preset", "preset", true)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	_, _, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", &plainPresetBackend{}, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support email lookup")
}

func TestProcessUsers_backendWithoutLookupKeepsCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", &plainPresetBackend{}, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, add)
	assert.Empty(t, remove)
}

func TestProcessUsers_flagOnCallsEmailLookup(t *testing.T) {
	r := newBackendReconciler(t, "gitlab", "gitlab", true)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", "gitlab_gitlab", "old-id"))

	backend := &emailPresetBackend{found: &structs.User{ID: "new-id", Email: "ruagrawa@redhat.com"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"gitlab", "gitlab", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"new-id"}, add)
	assert.Empty(t, remove)
	assert.Equal(t, 1, backend.calls)
}

func TestCreateUsers_missingCacheCreatesUser(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	backend := &plainPresetBackend{}
	err := r.createUsersInBackendAndCache(ctx, []string{"ruagrawa"}, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	require.Len(t, backend.created, 1)
	assert.Equal(t, "ruagrawa@redhat.com", backend.created[0].Email)
	assert.Equal(t, "ruagrawa", backend.created[0].UserName)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "created-id", cached[presetBackendKey])
}

func TestCreateUsers_cachedIDSkipsCreate(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &plainPresetBackend{}
	err := r.createUsersInBackendAndCache(ctx, []string{"ruagrawa"}, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, backend.created)
}
