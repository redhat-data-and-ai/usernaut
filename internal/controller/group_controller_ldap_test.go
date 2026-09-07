package controller

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/internal/controller/mocks"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

func TestFetchLDAPData_SkippedUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ldapClient := mocks.NewMockLDAPClient(ctrl)
	r := &GroupReconciler{
		LdapConn: ldapClient,
		log:      logger.Logger(context.Background()),
	}

	members := []string{"found-user", "missing-user"}
	ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), members).Return(map[string]map[string]interface{}{
		"found-user": {
			"cn":          "Found",
			"sn":          "User",
			"displayName": "Found User",
			"mail":        "found@example.com",
			"uid":         "found-user",
		},
	}, nil)

	result, err := r.fetchLDAPData(context.Background(), members)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"found@example.com"}, result.CurrentMembers)
	assert.Equal(t, []string{"found-user"}, result.ActiveUserList)
	assert.Equal(t, []string{"missing-user"}, result.SkippedUsers)
	require.Contains(t, r.allLdapUserData, "found-user")
	assert.NotContains(t, r.allLdapUserData, "missing-user")
}

func TestFetchLDAPData_AllUsersFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ldapClient := mocks.NewMockLDAPClient(ctrl)
	r := &GroupReconciler{
		LdapConn: ldapClient,
		log:      logger.Logger(context.Background()),
	}

	members := []string{"found-user"}
	ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), members).Return(map[string]map[string]interface{}{
		"found-user": {
			"cn":          "Found",
			"sn":          "User",
			"displayName": "Found User",
			"mail":        "found@example.com",
			"uid":         "found-user",
		},
	}, nil)

	result, err := r.fetchLDAPData(context.Background(), members)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.SkippedUsers)
	assert.Equal(t, []string{"found@example.com"}, result.CurrentMembers)
}
