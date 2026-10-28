package controller

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-data-and-ai/usernaut/internal/controller/mocks"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFetchLDAPDataClassifiesMembers(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			ldapClient := mocks.NewMockLDAPClient(gomock.NewController(t))
			r := &GroupReconciler{
				LdapConn:  ldapClient,
				AppConfig: &config.AppConfig{LDAP: ldap.LDAP{ServiceAccountObjectFilter: "(objectClass=serviceAccount)"}},
			}
			members := []string{"person", "service", "missing"}
			ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), members).Return(map[string]map[string]interface{}{
				"person":  {"uid": "person", "mail": "person@example.com", "isServiceAccount": false},
				"service": {"uid": "service", "mail": "service@example.com", "isServiceAccount": true},
			}, nil)
			result, err := r.fetchLDAPData(t.Context(), members, enabled, logrus.NewEntry(logrus.New()))
			require.NoError(t, err)
			assert.Len(t, result.UserData, 2)
			assert.Equal(t, []string{"person"}, result.PersonMembers)
			if enabled {
				assert.Equal(t, []string{"person", "service"}, result.Members)
				assert.Empty(t, result.IgnoredServiceAccounts)
				assert.Equal(t, []string{"person@example.com", "service@example.com"}, result.CurrentMembers)
			} else {
				assert.Equal(t, []string{"person"}, result.Members)
				assert.Equal(t, []string{"service"}, result.IgnoredServiceAccounts)
				assert.Equal(t, []string{"person@example.com"}, result.CurrentMembers)
			}
			assert.ElementsMatch(t, result.Members, result.ActiveUserList)
		})
	}
}

func TestOptOutRemovesExistingServiceAccount(t *testing.T) {
	// A service account excluded from desired membership must remain in the fetched
	// backend set so reconciliation selects it for removal.
	r := &GroupReconciler{}
	add, remove, err := r.processUsers(t.Context(), nil,
		map[string]*structs.User{"service": {ID: "service"}},
		"rover", "rover", nil, logrus.NewEntry(logrus.New()), nil)
	assert.NoError(t, err)
	assert.Empty(t, add)
	assert.Equal(t, []string{"service"}, remove)
}
