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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
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
		baseUserDN:       "ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "cn", "sn"},
	}

	resp, err := ldapConn.GetUserLDAPDataByEmail(suite.ctx, "testuser@example.com")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetUserLDAPData_ContextCanceledBeforeSearch() {
	assertions := assert.New(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		userDN:           "uid=%s,ou=users,dc=example,dc=com",
		baseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=uid)",
		attributes:       []string{"mail"},
	}

	resp, err := ldapConn.GetUserLDAPData(ctx, "testuser")

	assertions.ErrorIs(err, context.Canceled)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetBulkUserLDAPData_ContextCanceledBeforeLDAP() {
	assertions := assert.New(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		baseUserDN:       "ou=users,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "uid"},
	}

	out, err := ldapConn.GetBulkUserLDAPData(ctx, []string{"u1", "u2"})

	assertions.ErrorIs(err, context.Canceled)
	assertions.Empty(out)
}

func (suite *LDAPTestSuite) TestGetBulkUserLDAPData_ContextCanceledStopsAfterFirstBatch() {
	assertions := assert.New(suite.T())

	prev := bulkLDAPBatchSize
	bulkLDAPBatchSize = 2
	defer func() { bulkLDAPBatchSize = prev }()

	ctx, cancel := context.WithCancel(context.Background())

	entry := func(uid string) *ldap.Entry {
		return &ldap.Entry{
			DN: "uid=" + uid + ",ou=users,dc=example,dc=com",
			Attributes: []*ldap.EntryAttribute{
				{Name: "uid", Values: []string{uid}},
				{Name: "mail", Values: []string{uid + "@example.com"}},
			},
		}
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).DoAndReturn(
		func(_ *ldap.SearchRequest) (*ldap.SearchResult, error) {
			cancel()
			return &ldap.SearchResult{Entries: []*ldap.Entry{entry("a1")}}, nil
		},
	).Times(1)

	ldapConn := &LDAPConn{
		conn:             suite.ldapClient,
		baseUserDN:       "ou=users,dc=example,dc=com",
		server:           "ldap://ldap.com:389",
		userSearchFilter: "(objectClass=person)",
		attributes:       []string{"mail", "uid"},
	}

	out, err := ldapConn.GetBulkUserLDAPData(ctx, []string{"a1", "a2", "b1", "b2"})

	assertions.ErrorIs(err, context.Canceled)
	assertions.Len(out, 1)
	assertions.Contains(out, "a1")
}

func (suite *LDAPTestSuite) TestBulkLookupClassifiesByServiceAccountFilter() {
	const serviceFilter = "(&(employeeType=service)(accountStatus=active))"
	conn := &LDAPConn{
		conn: suite.ldapClient, userSearchFilter: "(objectClass=person)",
		serviceAccountObjectFilter: serviceFilter, attributes: []string{"mail"},
	}
	ids := []string{"person", "shared", "svc"}
	uidFilter := "(|(uid=person)(uid=shared)(uid=svc))"
	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(2)
	calls := 0
	suite.ldapClient.EXPECT().Search(gomock.Any()).DoAndReturn(func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
		suite.Contains(req.Attributes, "uid")
		_, err := ldap.CompileFilter(req.Filter)
		suite.NoError(err)
		calls++
		if calls == 1 {
			suite.Equal("(&(objectClass=person)"+uidFilter+")", req.Filter)
			return &ldap.SearchResult{Entries: []*ldap.Entry{
				ldap.NewEntry("uid=person", map[string][]string{"uid": {"person"}, "mail": {"person@example.com"}}),
				ldap.NewEntry("uid=shared", map[string][]string{"uid": {"shared"}, "mail": {"shared@example.com"}}),
			}}, nil
		}
		suite.Equal("(&"+serviceFilter+uidFilter+")", req.Filter)
		// No objectClass is returned: classification comes from the matching search.
		return &ldap.SearchResult{Entries: []*ldap.Entry{
			ldap.NewEntry("uid=shared", map[string][]string{"uid": {"shared"}, "mail": {"shared@example.com"}}),
			ldap.NewEntry("uid=svc", map[string][]string{"uid": {"svc"}, "mail": {"svc@example.com"}}),
		}}, nil
	}).Times(2)
	users, err := conn.GetBulkUserLDAPData(suite.ctx, ids)
	suite.Require().NoError(err)
	suite.Len(users, 3)
	suite.Equal(false, users["person"]["isServiceAccount"])
	suite.Equal(true, users["shared"]["isServiceAccount"])
	suite.Equal(true, users["svc"]["isServiceAccount"])
}

func (suite *LDAPTestSuite) TestServiceAccountSearchFailureRejectsPartialResults() {
	conn := &LDAPConn{
		conn: suite.ldapClient, userSearchFilter: "(objectClass=person)",
		serviceAccountObjectFilter: "(employeeType=service)",
	}
	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(2)
	gomock.InOrder(
		suite.ldapClient.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{Entries: []*ldap.Entry{
			ldap.NewEntry("uid=shared", map[string][]string{"uid": {"shared"}}),
		}}, nil),
		suite.ldapClient.EXPECT().Search(gomock.Any()).Return(nil, assert.AnError),
	)
	users, err := conn.GetBulkUserLDAPData(suite.ctx, []string{"shared"})
	suite.ErrorIs(err, assert.AnError)
	suite.Nil(users)
}

func (suite *LDAPTestSuite) TestIdentityLookupsIncludeServiceAccounts() {
	conn := &LDAPConn{
		conn:                       suite.ldapClient,
		userDN:                     "uid=%s,ou=users,dc=example,dc=com",
		userSearchFilter:           "(objectClass=person)",
		serviceAccountObjectFilter: "(&(employeeType=service)(accountStatus=active))",
		attributes:                 []string{"uid", "mail"},
	}
	identityFilter := "(|(objectClass=person)(&(employeeType=service)(accountStatus=active)))"
	expectedFilters := []string{identityFilter, "(&" + identityFilter + "(mail=svc@example.com))"}
	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(2)
	calls := 0
	suite.ldapClient.EXPECT().Search(gomock.Any()).DoAndReturn(func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
		suite.Equal(expectedFilters[calls], req.Filter)
		_, err := ldap.CompileFilter(req.Filter)
		suite.NoError(err)
		calls++
		return &ldap.SearchResult{Entries: []*ldap.Entry{ldap.NewEntry("uid=svc", map[string][]string{
			"uid": {"svc"}, "mail": {"svc@example.com"}, "objectClass": {"top", "serviceAccount"},
		})}}, nil
	}).Times(2)
	ctx := suite.ctx
	byID, err := conn.GetUserLDAPData(ctx, "svc")
	suite.NoError(err)
	suite.Equal("svc", byID["uid"])
	byEmail, err := conn.GetUserLDAPDataByEmail(ctx, "svc@example.com")
	suite.NoError(err)
	suite.Equal("svc", byEmail["uid"])
}

func (suite *LDAPTestSuite) TestMissingServiceAccountIsStillAbsent() {
	conn := &LDAPConn{
		conn:                       suite.ldapClient,
		userSearchFilter:           "(objectClass=person)",
		serviceAccountObjectFilter: "(objectClass=serviceAccount)",
	}
	suite.ldapClient.EXPECT().IsClosing().Return(false)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{}, nil)
	_, err := conn.GetUserLDAPDataByEmail(suite.ctx, "missing@example.com")
	suite.ErrorIs(err, ErrNoUserFound)
}

func TestIdentitySearchFilterParentheses(t *testing.T) {
	for _, tc := range []struct{ input, normalized string }{
		{"objectClass=person", "(objectClass=person)"},
		{"(objectClass=person)", "(objectClass=person)"},
		{"  objectClass=person  ", "(objectClass=person)"},
		{"  (objectClass=person)  ", "(objectClass=person)"},
		{"|(objectClass=person)(objectClass=inetOrgPerson)", "(|(objectClass=person)(objectClass=inetOrgPerson))"},
		{"(|(objectClass=person)(objectClass=inetOrgPerson))", "(|(objectClass=person)(objectClass=inetOrgPerson))"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			for _, serviceFilter := range []string{"", "objectClass=serviceAccount", "(objectClass=serviceAccount)"} {
				conn := &LDAPConn{userSearchFilter: tc.input, serviceAccountObjectFilter: serviceFilter}
				want := tc.normalized
				if serviceFilter != "" {
					want = "(|" + want + "(objectClass=serviceAccount))"
				}
				got := conn.identitySearchFilter()
				assert.Equal(t, want, got)
				_, err := ldap.CompileFilter(got)
				assert.NoError(t, err)
			}
		})
	}
}
