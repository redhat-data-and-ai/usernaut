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
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
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
		AppConfig: &config.AppConfig{
			BackendMap: map[string]map[string]config.Backend{
				"preset": {"preset": {CompareUsersByEmail: true}},
			},
		},
	}
}

func presetLDAPUser() map[string]*structs.LDAPUser {
	return map[string]*structs.LDAPUser{
		"ruagrawa": {Email: "ruagrawa@redhat.com", UID: "ruagrawa"},
	}
}

type emailLookup struct {
	clients.Client
	found *structs.User
	err   error
	calls int
}

func (b *emailLookup) FetchUserDetailsByEmail(context.Context, string) (*structs.User, error) {
	b.calls++
	return b.found, b.err
}

func TestMemberIDByEmail(t *testing.T) {
	r := newPresetReconciler(t)
	members := map[string]*structs.User{
		"oidc|new": {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	assert.Equal(t, "oidc|new", r.memberIDByEmail(members, "ruagrawa@redhat.com"))
	assert.Empty(t, r.memberIDByEmail(members, "other@redhat.com"))
}

func TestProcessUsers_memberEmailReplacesStaleID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	members := map[string]*structs.User{
		"oidc|new": {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	backend := &emailLookup{found: &structs.User{ID: "should-not-be-used"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)
	assert.Zero(t, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "oidc|new", cached[presetBackendKey])
}

func TestProcessUsers_keepsCachedMemberWhenEmailMatches(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	members := map[string]*structs.User{
		"samlp|old": {ID: "samlp|old", Email: "ruagrawa@redhat.com"},
	}
	backend := &emailLookup{found: &structs.User{ID: "should-not-be-used"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)
	assert.Zero(t, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_noMemberEmailUsesLookedUpID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	backend := &emailLookup{found: &structs.User{ID: "oidc|new", Email: "ruagrawa@redhat.com"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{}, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
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

	backend := &emailLookup{}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{}, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
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

	backend := &emailLookup{err: errors.New("preset unavailable")}
	_, _, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{}, "preset", "preset",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.Error(t, err)
	assert.Equal(t, 1, backend.calls)
}

func TestProcessUsers_flagOffIgnoresMemberEmail(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", "gitlab_gitlab", "100"))

	members := map[string]*structs.User{
		"200": {ID: "200", Email: "ruagrawa@redhat.com"},
	}
	backend := &emailLookup{found: &structs.User{ID: "should-not-be-used"}}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "gitlab", "gitlab",
		backend, logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"100"}, add)
	assert.Equal(t, []string{"200"}, remove)
	assert.Zero(t, backend.calls)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "100", cached["gitlab_gitlab"])
}
