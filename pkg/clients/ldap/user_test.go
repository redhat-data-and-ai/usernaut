package ldap

import (
	"context"
	"errors"
	"fmt"
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
	assertions.Equal("ou=adhoc,ou=managedGroups,dc=example,dc=com", ldapConn.GetBaseDN(), "Expected baseDN to match the format")

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
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(nil, ldap.NewError(ldap.LDAPResultOperationsError, errors.New("search error"))).Times(1)

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

	addr, stop := startMockLDAPServer(suite.T())
	defer stop()

	assertions := assert.New(suite.T())
	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           fmt.Sprintf("ldap://%s", addr),
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(true).Times(1)

	conn := ldapConn.getConn()
	assertions.NotNil(conn, "Expected a new LDAP connection to be returned when the existing one is closing")
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
