package ldap

import (
	"context"
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LDAPTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	ctx  context.Context

	ldapClient *mocks.MockLDAPConnClient
}

func TestLdap(t *testing.T) {
	suite.Run(t, new(LDAPTestSuite))
}

func (suite *LDAPTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.ctrl = gomock.NewController(suite.T())

	suite.ldapClient = mocks.NewMockLDAPConnClient(suite.ctrl)
}

func (suite *LDAPTestSuite) TestGetUserLDAPData() {

	assertions := assert.New(suite.T())

	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "uid=testuser,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{
					{
						Name:   "mail",
						Values: []string{"testuser@gmail.com"},
					},
				},
			},
		},
	}
	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(searchResult, nil).Times(1)

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	assertions.Equal("uid=%s,ou=users,dc=example,dc=com", ldapConn.GetUserDN(), "Expected userDN to match the format")
	assertions.Equal("ou=adhoc,ou=managedGroups,dc=example,dc=com",
		ldapConn.GetBaseDN(), "Expected baseDN to match the format")

	resp, err := ldapConn.GetUserLDAPData(suite.ctx, "testuser")

	assertions.NoError(err)
	assertions.Equal("testuser@gmail.com", resp["mail"].(string))
}

func (suite *LDAPTestSuite) TestGetUserLDAPData_NoUserFound() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{Entries: []*ldap.Entry{}}, nil).Times(1)

	resp, err := ldapConn.GetUserLDAPData(suite.ctx, "nonexistentuser")

	assertions.ErrorIs(err, ErrNoUserFound)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPData_EmptyAttributes() {
	assertions := assert.New(suite.T())

	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "uid=testuser,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{},
			},
		},
	}
	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}
	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(searchResult, nil).Times(1)
	resp, err := ldapConn.GetUserLDAPData(suite.ctx, "testuser")
	assertions.NoError(err)
	assertions.Equal("", resp["mail"].(string), "Expected empty string for mail attribute")
}

func (suite *LDAPTestSuite) TestSearchError() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).
		Return(nil, ldap.NewError(ldap.LDAPResultOperationsError, errors.New("search error"))).Times(1)

	resp, err := ldapConn.GetUserLDAPData(suite.ctx, "testuser")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPData_NilConnection() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             nil, // Simulating a nil connection
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	resp, err := ldapConn.GetUserLDAPData(suite.ctx, "testuser")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetLdapConnection_Success() {
	// Test that when connection is not closing, the existing connection is returned
	assertions := assert.New(suite.T())
	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)

	conn := ldapConn.getConn()
	assertions.Equal(suite.ldapClient, conn, "Expected existing connection when not closing")
}

func (suite *LDAPTestSuite) TestGetLdapConnection_ReconnectionAttempt() {
	// Test that when connection is closing, reconnection is attempted
	// Note: This test verifies the reconnection attempt happens, but actual reconnection
	// requires a real LDAP server. The reconnection logic is tested indirectly through
	// integration tests or when using a real LDAP server.
	assertions := assert.New(suite.T())
	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://invalid-server:389", // Invalid server to test failure path
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(true).Times(1)

	conn := ldapConn.getConn()
	// Reconnection will fail with invalid server, verifying that reconnection attempt was made
	assertions.Nil(conn, "Expected nil when reconnection fails with invalid server")
}

func (suite *LDAPTestSuite) TestGetLdapConnection_Failure() {
	assertions := assert.New(suite.T())
	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(true).Times(1)

	conn := ldapConn.getConn()
	assertions.Nil(conn, "Failure to be returned when the existing one is closing and reconnecting")
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail() {
	assertions := assert.New(suite.T())

	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "uid=testuser,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{
					{
						Name:   "mail",
						Values: []string{"testuser@example.com"},
					},
					{
						Name:   "cn",
						Values: []string{"Test User"},
					},
					{
						Name:   "sn",
						Values: []string{"User"},
					},
				},
			},
		},
	}

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(searchResult, nil).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.NoError(err)
	assertions.Equal("testuser@example.com", resp["mail"].(string))
	assertions.Equal("Test User", resp["cn"].(string))
	assertions.Equal("User", resp["sn"].(string))
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_NoUserFound() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{Entries: []*ldap.Entry{}}, nil).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "nonexistent@example.com")

	assertions.ErrorIs(err, ErrNoUserFound)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_NoSuchObject() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).
		Return(nil, ldap.NewError(ldap.LDAPResultNoSuchObject, errors.New("no such object"))).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "nonexistent@example.com")

	assertions.ErrorIs(err, ErrNoUserFound)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_EmptyAttributes() {
	assertions := assert.New(suite.T())

	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "uid=testuser,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{},
			},
		},
	}

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(searchResult, nil).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.NoError(err)
	assertions.Equal("", resp["mail"].(string), "Expected empty string for mail attribute")
	assertions.Equal("", resp["cn"].(string), "Expected empty string for cn attribute")
	assertions.Equal("", resp["sn"].(string), "Expected empty string for sn attribute")
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_SearchError() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(nil).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).
		Return(nil, ldap.NewError(ldap.LDAPResultOperationsError, errors.New("search error"))).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_NilConnection() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             nil, // Simulating a nil connection
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPDataByEmail_BindError() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		BaseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().UnauthenticatedBind("").Return(errors.New("bind failed")).Times(1)

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.Error(err)
	assertions.Contains(err.Error(), "failed to bind before search")
	assertions.Nil(resp)
}
