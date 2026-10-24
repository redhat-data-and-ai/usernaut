package controller

import (
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMembersForBackend(t *testing.T) {
	r := &GroupReconciler{AppConfig: &config.AppConfig{LDAP: ldap.LDAP{ServiceAccountObjectClass: "serviceAccount"}}}
	members := []string{"person", "service", "missing"}
	data := map[string]*structs.LDAPUser{
		"person":  {UID: "person"},
		"service": {UID: "service", ObjectClass: "top serviceAccount"},
	}
	assert.Equal(t, []string{"person"}, r.membersForBackend(members, data, false))
	assert.Equal(t, []string{"person", "service"}, r.membersForBackend(members, data, true))
	assert.Len(t, members, 3)
	assert.Len(t, data, 2)
}

func TestOptOutRemovesExistingServiceAccount(t *testing.T) {
	// A service account excluded from desired membership must remain in the fetched
	// backend set so reconciliation selects it for removal.
	r := &GroupReconciler{}
	add, remove, err := r.processUsers(t.Context(), nil,
		map[string]*structs.User{"service": {ID: "service"}},
		"rover", "rover", logrus.NewEntry(logrus.New()), nil)
	assert.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"service"}, remove)
}
