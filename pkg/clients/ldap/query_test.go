package ldap

import (
	"errors"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func (suite *LDAPTestSuite) TestGetQueryMembers() {
	assertions := assert.New(suite.T())

	query := "(objectClass=groupOfNames)"
	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "uid=example-user,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"example-user"}},
				},
			},
		},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	var capturedReq *ldap.SearchRequest
	suite.ldapClient.EXPECT().
		Search(gomock.Any()).
		DoAndReturn(func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
			capturedReq = req
			return searchResult, nil
		}).
		Times(1)

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		userDN:     "uid=%s,ou=users,dc=example,dc=com",
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn", "description"},
	}

	resp, err := ldapConn.GetQueryMembers(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal("example-user", resp[0])

	// Verify the search request is constructed correctly.
	if assertions.NotNil(capturedReq) {
		assertions.Equal(ldapConn.baseUserDN, capturedReq.BaseDN)
		assertions.Equal(ldap.ScopeWholeSubtree, capturedReq.Scope)
		assertions.Equal(query, capturedReq.Filter)
		assertions.Equal([]string{"uid"}, capturedReq.Attributes)
	}
}

func (suite *LDAPTestSuite) TestGetQueryMembers_NoEntriesFound() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{Entries: []*ldap.Entry{}}, nil).Times(1)

	resp, err := ldapConn.GetQueryMembers(suite.ctx, "(objectClass=groupOfNames)")

	assertions.NoError(err)
	assertions.Empty(resp)
}

func (suite *LDAPTestSuite) TestGetQueryMembers_EmptyAttributes() {
	assertions := assert.New(suite.T())

	searchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "uid=dn-only-user,ou=users,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{},
			},
		},
	}

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).Return(searchResult, nil).Times(1)

	resp, err := ldapConn.GetQueryMembers(suite.ctx, "(objectClass=groupOfNames)")

	assertions.NoError(err)
	assertions.Equal("dn-only-user", resp[0])
}

func (suite *LDAPTestSuite) TestGetQueryMembers_SearchError() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn"},
	}

	suite.ldapClient.EXPECT().IsClosing().Return(false).Times(1)
	suite.ldapClient.EXPECT().Search(gomock.Any()).
		Return(nil, ldap.NewError(ldap.LDAPResultOperationsError, errors.New("search error"))).Times(1)

	resp, err := ldapConn.GetQueryMembers(suite.ctx, "(objectClass=groupOfNames)")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetQueryMembers_NilConnection() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:       nil, // Simulating a nil connection
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn"},
	}

	resp, err := ldapConn.GetQueryMembers(suite.ctx, "(objectClass=groupOfNames)")

	assertions.Error(err)
	assertions.Nil(resp)
}

func (suite *LDAPTestSuite) TestGetQueryMembers_EmptyQuery() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
	}

	resp, err := ldapConn.GetQueryMembers(suite.ctx, "")

	assertions.NoError(err)
	assertions.Empty(resp)
}
