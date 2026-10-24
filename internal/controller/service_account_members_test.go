package controller

import (
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMembershipSearchFilter(t *testing.T) {
	cfg := ldap.LDAP{UserSearchFilter: "(objectClass=person)", ServiceAccountObjectClass: "serviceAccount"}
	filter, err := membershipSearchFilter(cfg, false)
	require.NoError(t, err)
	assert.Equal(t, "(&(objectClass=person)(!(objectClass=serviceAccount)))", filter)
	filter, err = membershipSearchFilter(cfg, true)
	require.NoError(t, err)
	assert.Equal(t, "(|(objectClass=person)(objectClass=serviceAccount))", filter)
	cfg.ServiceAccountObjectClass = "service*Account"
	filter, err = membershipSearchFilter(cfg, true)
	require.NoError(t, err)
	assert.Contains(t, filter, `service\2aAccount`)
}

func TestFetchLDAPDataRejectsMissingServiceAccountClass(t *testing.T) {
	for _, allowSA := range []bool{false, true} {
		for _, class := range []string{"", " ", "\t"} {
			r := &GroupReconciler{AppConfig: &config.AppConfig{LDAP: ldap.LDAP{
				UserSearchFilter:          "(|(objectClass=person)(objectClass=serviceAccount))",
				ServiceAccountObjectClass: class,
			}}}
			// No LDAP client: invalid configuration must fail before any LDAP/backend work.
			result, users, err := r.fetchLDAPData(t.Context(), []string{"service"}, allowSA, logrus.NewEntry(logrus.New()))
			require.ErrorContains(t, err, "ldap.serviceAccountObjectClass is required")
			assert.Nil(t, result)
			assert.Nil(t, users)
		}
	}
}

func TestReconcilableMembers(t *testing.T) {
	members := []string{"person", "service", "missing"}
	data := map[string]*structs.LDAPUser{"person": {UID: "person"}}
	log := logrus.NewEntry(logrus.New())
	assert.Equal(t, []string{"person"}, reconcilableMembers(members, data, log))
	assert.Len(t, members, 3)
	data["service"] = &structs.LDAPUser{UID: "service", ObjectClass: "serviceAccount"}
	assert.Equal(t, []string{"person", "service"}, reconcilableMembers(members, data, log))
}

func TestOptOutRemovesExistingServiceAccount(t *testing.T) {
	// A service account excluded by LDAP must still be present in the fetched
	// backend set so reconciliation selects it for removal.
	r := &GroupReconciler{}
	add, remove, err := r.processUsers(t.Context(), nil,
		map[string]*structs.User{"service": {ID: "service"}},
		"rover", "rover", logrus.NewEntry(logrus.New()), nil)
	assert.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"service"}, remove)
}
