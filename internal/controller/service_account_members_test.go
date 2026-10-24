package controller

import (
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMembershipSearchFilter(t *testing.T) {
	cfg := ldap.LDAP{UserSearchFilter: "(objectClass=person)", ServiceAccountObjectClass: "serviceAccount"}
	assert.Equal(t, "(&(objectClass=person)(!(objectClass=serviceAccount)))", membershipSearchFilter(cfg, false))
	assert.Equal(t, "(|(objectClass=person)(objectClass=serviceAccount))", membershipSearchFilter(cfg, true))
	cfg.ServiceAccountObjectClass = ""
	assert.Equal(t, cfg.UserSearchFilter, membershipSearchFilter(cfg, false))
	assert.Equal(t, cfg.UserSearchFilter, membershipSearchFilter(cfg, true))
	cfg.ServiceAccountObjectClass = "service*Account"
	assert.Contains(t, membershipSearchFilter(cfg, true), `service\2aAccount`)
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
