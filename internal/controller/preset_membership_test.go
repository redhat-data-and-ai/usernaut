package controller

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

const presetBackendKey = "preset_preset"

func newPresetReconciler(t *testing.T) *GroupReconciler {
	t.Helper()
	cacheClient, err := inmemory.NewCache(&inmemory.Config{DefaultExpiration: -1, CleanupInterval: -1})
	require.NoError(t, err)
	return &GroupReconciler{
		Store:      store.New(cacheClient),
		CacheMutex: &sync.RWMutex{},
	}
}

func presetLDAPUser() map[string]*structs.LDAPUser {
	return map[string]*structs.LDAPUser{
		"ruagrawa": {Email: "ruagrawa@redhat.com", UID: "ruagrawa"},
	}
}

type plainPresetBackend struct {
	clients.Client
	created []*structs.User
}

func (b *plainPresetBackend) CreateUser(_ context.Context, user *structs.User) (*structs.User, error) {
	b.created = append(b.created, user)
	return &structs.User{ID: "created-id", Email: user.Email, UserName: user.UserName}, nil
}

type emailLookupBackend struct {
	clients.Client
	found *structs.User
	err   error
	calls int
}

func (b *emailLookupBackend) FindUserByEmail(context.Context, string) (*structs.User, error) {
	b.calls++
	return b.found, b.err
}

func TestProcessUsers_staleIDNotOnGroupKeepsCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", nil, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, add)
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_noMemberMatchUsesEmailLookup(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &emailLookupBackend{found: &structs.User{ID: "oidc|new", Email: "ruagrawa@redhat.com"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"oidc|new"}, add)
	assert.Empty(t, remove)
	assert.Equal(t, 1, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_emailLookupMissKeepsCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &emailLookupBackend{}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, add)
	assert.Empty(t, remove)
	assert.Equal(t, 1, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_emailLookupErrorStopsBackend(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &emailLookupBackend{err: errors.New("preset unavailable")}
	_, _, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{},
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.Error(t, err)
	assert.Equal(t, 1, backend.calls)
}

func TestProcessUsers_memberEmailReplacesStaleID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{
		"oidc|new": {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	backend := &emailLookupBackend{found: &structs.User{ID: "should-not-be-used"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)
	assert.Zero(t, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_emailMatchReplacesCachedMember(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{
		"samlp|old": {ID: "samlp|old"},
		"oidc|new":  {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", nil, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"samlp|old"}, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_duplicateEmailDropsTheOtherID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{
		"samlp|old": {ID: "samlp|old", Email: "ruagrawa@redhat.com"},
		"oidc|new":  {ID: "oidc|new", Email: "Ruagrawa@redhat.com"},
	}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", nil, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"oidc|new"}, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_nonPresetIgnoresMemberEmail(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", "gitlab_gitlab", "100"))

	onGroup := map[string]*structs.User{
		"200": {ID: "200", Email: "ruagrawa@redhat.com"},
	}
	backend := &emailLookupBackend{found: &structs.User{ID: "should-not-be-used"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"gitlab", "gitlab", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"100"}, add)
	assert.Equal(t, []string{"200"}, remove)
	assert.Zero(t, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "100", cached["gitlab_gitlab"])
}

func TestProcessUsers_removedMemberWithNewSCIMIDIsRemoved(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{
		"oidc|new": {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	add, remove, err := r.processUsers(ctx, nil, onGroup,
		"preset", "preset", nil, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"oidc|new"}, remove)
}

func TestProcessUsers_memberWithoutEmailUsesCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	onGroup := map[string]*structs.User{"oidc|new": {ID: "oidc|new"}}
	backend := &emailLookupBackend{}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, onGroup,
		"preset", "preset", backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, add)
	assert.Equal(t, []string{"oidc|new"}, remove)
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
