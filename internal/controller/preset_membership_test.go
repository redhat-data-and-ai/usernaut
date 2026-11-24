package controller

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
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

func TestProcessUsers_memberEmailReplacesStaleID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	members := map[string]*structs.User{
		"oidc|new": {ID: "oidc|new", Email: "ruagrawa@redhat.com"},
	}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "preset", "preset",
		logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)

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
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "preset", "preset",
		logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_noMemberEmailKeepsCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, map[string]*structs.User{}, "preset", "preset",
		logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"samlp|old"}, addedIDs(add))
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_matchingMemberWithoutEmailKeepsCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", presetBackendKey, "samlp|old"))

	members := map[string]*structs.User{
		"samlp|old": {ID: "samlp|old"},
	}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "preset", "preset",
		logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "samlp|old", cached[presetBackendKey])
}

func TestProcessUsers_memberEmailReplacesCachedID(t *testing.T) {
	r := newPresetReconciler(t)
	ctx := t.Context()
	require.NoError(t, r.Store.User.SetBackend(ctx, "ruagrawa@redhat.com", "gitlab_gitlab", "100"))

	members := map[string]*structs.User{
		"200": {ID: "200", Email: "ruagrawa@redhat.com"},
	}
	add, remove, err := r.processUsers(ctx, []string{"ruagrawa"}, members, "gitlab", "gitlab",
		logrus.NewEntry(logrus.New()), presetLDAPUser())
	require.NoError(t, err)
	assert.Empty(t, add)
	assert.Empty(t, remove)

	cached, err := r.Store.User.GetBackends(ctx, "ruagrawa@redhat.com")
	require.NoError(t, err)
	assert.Equal(t, "200", cached["gitlab_gitlab"])
}

func addedIDs(pairs []*structs.Member) []string {
	ids := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		if pair == nil {
			continue
		}
		ids = append(ids, pair.GetUserID())
	}
	return ids
}

func assertIDs(t *testing.T, want, got []string) {
	t.Helper()
	if len(want) == 0 {
		assert.Empty(t, got)
		return
	}
	assert.Equal(t, want, got)
}

func TestPresetGroupScenarios(t *testing.T) {
	const email = "ruagrawa@redhat.com"
	ldapUsers := presetLDAPUser()

	tests := []struct {
		name       string
		desired    []string
		members    map[string]*structs.User
		wantAdd    []string
		wantRemove []string
		wantCache  string
	}{
		{
			name:    "member without email keeps cached ID",
			desired: []string{"ruagrawa"},
			members: map[string]*structs.User{
				"oidc|new": {ID: "oidc|new"},
			},
			wantAdd:    []string{"samlp|old"},
			wantRemove: []string{"oidc|new"},
			wantCache:  "samlp|old",
		},
		{
			name:      "user not on group keeps cached ID",
			desired:   []string{"ruagrawa"},
			members:   map[string]*structs.User{},
			wantAdd:   []string{"samlp|old"},
			wantCache: "samlp|old",
		},
		{
			name:    "user on group and scim id unchanged",
			desired: []string{"ruagrawa"},
			members: map[string]*structs.User{
				"samlp|old": {ID: "samlp|old"},
			},
			wantCache: "samlp|old",
		},
		{
			name: "user removed from group and scim id changed",
			members: map[string]*structs.User{
				"oidc|new": {ID: "oidc|new", Email: email},
			},
			wantRemove: []string{"oidc|new"},
			wantCache:  "oidc|new",
		},
		{
			name: "user removed from group and scim id unchanged",
			members: map[string]*structs.User{
				"samlp|old": {ID: "samlp|old", Email: email},
			},
			wantRemove: []string{"samlp|old"},
			wantCache:  "samlp|old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newPresetReconciler(t)
			ctx := t.Context()
			require.NoError(t, r.Store.User.SetBackend(ctx, email, presetBackendKey, "samlp|old"))

			add, remove, err := r.processUsers(ctx, tt.desired, tt.members, "preset", "preset",
				logrus.NewEntry(logrus.New()), ldapUsers)
			require.NoError(t, err)
			assertIDs(t, tt.wantAdd, addedIDs(add))
			assertIDs(t, tt.wantRemove, remove)

			cached, err := r.Store.User.GetBackends(ctx, email)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCache, cached[presetBackendKey])
		})
	}
}

func TestProcessUsersForAllBackends(t *testing.T) {
	const (
		email    = "ruagrawa@redhat.com"
		cachedID = "cached-id"
		liveID   = "live-id"
	)
	ldapUsers := presetLDAPUser()
	backends := []struct {
		name string
		typ  string
	}{
		{name: "preset", typ: "preset"},
		{name: "gitlab", typ: "gitlab"},
		{name: "snowflake", typ: "snowflake"},
		{name: "fivetran", typ: "fivetran"},
		{name: "astro", typ: "astro"},
		{name: "rover", typ: "rover"},
	}

	tests := []struct {
		name       string
		desired    []string
		members    map[string]*structs.User
		wantAdd    []string
		wantRemove []string
		wantCache  string
	}{
		{
			name:    "user in group and id unchanged",
			desired: []string{"ruagrawa"},
			members: map[string]*structs.User{
				cachedID: {ID: cachedID, Email: email},
			},
			wantCache: cachedID,
		},
		{
			name:      "user not in group and needs to be added",
			desired:   []string{"ruagrawa"},
			members:   map[string]*structs.User{},
			wantAdd:   []string{cachedID},
			wantCache: cachedID,
		},
		{
			name: "user removed from group",
			members: map[string]*structs.User{
				cachedID: {ID: cachedID, Email: email},
			},
			wantRemove: []string{cachedID},
			wantCache:  cachedID,
		},
		{
			name:    "member email replaces stale id",
			desired: []string{"ruagrawa"},
			members: map[string]*structs.User{
				liveID: {ID: liveID, Email: email},
			},
			wantCache: liveID,
		},
		{
			name:    "member without email keeps stale id",
			desired: []string{"ruagrawa"},
			members: map[string]*structs.User{
				liveID: {ID: liveID},
			},
			wantAdd:    []string{cachedID},
			wantRemove: []string{liveID},
			wantCache:  cachedID,
		},
	}

	for _, backend := range backends {
		t.Run(backend.name, func(t *testing.T) {
			backendKey := backend.name + "_" + backend.typ
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					r := newPresetReconciler(t)
					ctx := t.Context()
					require.NoError(t, r.Store.User.SetBackend(ctx, email, backendKey, cachedID))

					add, remove, err := r.processUsers(ctx, tt.desired, tt.members, backend.name, backend.typ,
						logrus.NewEntry(logrus.New()), ldapUsers)
					require.NoError(t, err)
					assertIDs(t, tt.wantAdd, addedIDs(add))
					assertIDs(t, tt.wantRemove, remove)

					cached, err := r.Store.User.GetBackends(ctx, email)
					require.NoError(t, err)
					assert.Equal(t, tt.wantCache, cached[backendKey])
				})
			}
		})
	}
}
