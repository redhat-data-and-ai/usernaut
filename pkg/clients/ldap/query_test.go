package ldap

import (
	"context"
	"errors"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
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
	assertions.Equal("example-user", resp[0].UID)

	// Verify the search request is constructed correctly.
	if assertions.NotNil(capturedReq) {
		assertions.Equal(ldapConn.baseUserDN, capturedReq.BaseDN)
		assertions.Equal(ldap.ScopeWholeSubtree, capturedReq.Scope)
		assertions.Equal(query, capturedReq.Filter)
		assertions.Equal(defaultAttributes, capturedReq.Attributes)
	}
}

func (suite *LDAPTestSuite) TestGetQueryMembers_ContextCanceled() {
	assertions := assert.New(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ldapConn := &LDAPConn{
		conn:       suite.ldapClient,
		baseUserDN: "ou=users,dc=example,dc=com",
		server:     "ldap://ldap.com:389",
		attributes: []string{"cn"},
	}

	resp, err := ldapConn.GetQueryMembers(ctx, "(objectClass=groupOfNames)")

	assertions.ErrorIs(err, context.Canceled)
	assertions.Nil(resp)
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

func (suite *LDAPTestSuite) TestGetQueryMembers_UIDFromDN() {
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
	assertions.Equal("dn-only-user", resp[0].UID)
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

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_AndOperator() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "manager",
				Criteria: "equals",
				Value:    "mgrAlpha",
			},
			{
				Key:      "title",
				Criteria: "contains",
				Value:    "senior",
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal("(&(manager=uid=mgrAlpha,ou=users,dc=redhat,dc=com)(title=*senior*))", filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_OrOperator() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "or",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "manager",
				Criteria: "equals",
				Value:    "mgrBeta",
			},
			{
				Key:      "manager",
				Criteria: "equals",
				Value:    "mgrGamma",
			},
		},
	}
	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal(
		"(|(manager=uid=mgrBeta,ou=users,dc=redhat,dc=com)(manager=uid=mgrGamma,ou=users,dc=redhat,dc=com))",
		filter,
	)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_ManagerContains() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "manager",
				Criteria: "contains",
				Value:    "foobar",
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal("(&(manager=uid=*foobar*,ou=users,dc=redhat,dc=com))", filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_MixOperator() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "manager",
				Criteria: "equals",
				Value:    "mgrDelta",
			},
			{
				Key:      "employeeType",
				Criteria: "not",
				Value:    "external employee",
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal("(&(manager=uid=mgrDelta,ou=users,dc=redhat,dc=com)(!(employeeType=external employee)))", filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_NestedOrInsideAnd() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "employeeType",
				Criteria: "not",
				Value:    "external employee",
			},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "engineer"},
						{Key: "title", Criteria: "contains", Value: "developer"},
						{Key: "title", Criteria: "contains", Value: "architect"},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal(
		"(&(!(employeeType=external employee))(|(title=*engineer*)(title=*developer*)(title=*architect*)))",
		filter,
	)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_NestedAndInsideOr() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "or",
		Filters: []v1alpha1.LDAPFilter{
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "and",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "engineer"},
						{Key: "co", Criteria: "equals", Value: "US"},
					},
				},
			},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "and",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "developer"},
						{Key: "co", Criteria: "equals", Value: "IND"},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal(
		"(|(&(title=*engineer*)(co=US))(&(title=*developer*)(co=IND)))",
		filter,
	)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_MultipleNestedQueries() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{Key: "employeeType", Criteria: "not", Value: "external employee"},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "engineer"},
						{Key: "title", Criteria: "contains", Value: "developer"},
					},
				},
			},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
						{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	expected := "(&(!(employeeType=external employee))" +
		"(|(title=*engineer*)(title=*developer*))" +
		"(|(manager=uid=mgrAlpha,ou=users,dc=redhat,dc=com)" +
		"(manager=uid=mgrBeta,ou=users,dc=redhat,dc=com)))"
	assertions.Equal(expected, filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_QueriesOnly() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "or",
		Filters: []v1alpha1.LDAPFilter{
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "and",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					},
				},
			},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "and",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)

	assertions.NoError(err)
	assertions.Equal(
		"(|(&(manager=uid=mgrAlpha,ou=users,dc=redhat,dc=com))(&(manager=uid=mgrBeta,ou=users,dc=redhat,dc=com)))",
		filter,
	)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_EmptyFiltersAndQueries() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
	}

	_, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.Error(err)
	assertions.Contains(err.Error(), "filters are empty")
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_InvalidNestedOperator() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "xor",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "engineer"},
					},
				},
			},
		},
	}

	_, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.Error(err)
	assertions.Contains(err.Error(), "unsupported operator")
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_EmptyNestedFilters() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters:  []v1alpha1.LDAPFilter{},
				},
			},
		},
	}

	_, err := ldapConn.BuildLDAPQueryString(context.Background(), query)
	assertions.Error(err)
	assertions.Contains(err.Error(), "filters are empty")
}

// TestBuildLDAPQueryString_BothSimpleAndNestedOnSameFilter verifies buildFilterItem rejects
// a filter item that sets both key/criteria/value and ldap_query.
func (suite *LDAPTestSuite) TestBuildLDAPQueryString_BothSimpleAndNestedOnSameFilter() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "title",
				Criteria: "contains",
				Value:    "engineer",
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "rhatGeo", Criteria: "equals", Value: "APAC"},
					},
				},
			},
		},
	}

	_, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.Error(err)
	assertions.Contains(err.Error(), "filter item cannot have both key/criteria/value and ldap_query")
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_FourLevelNesting() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{Key: "employeeType", Criteria: "not", Value: "external employee"},
			{
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "title", Criteria: "contains", Value: "engineer"},
						{
							LDAPQuery: &v1alpha1.LDAPNestedQuery{
								Operator: "and",
								Filters: []v1alpha1.LDAPFilter{
									{Key: "co", Criteria: "equals", Value: "US"},
									{
										LDAPQuery: &v1alpha1.LDAPNestedQuery{
											Operator: "or",
											Filters: []v1alpha1.LDAPFilter{
												{Key: "rhatCostCenter", Criteria: "equals", Value: "123"},
												{Key: "rhatCostCenter", Criteria: "equals", Value: "456"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.NoError(err)
	assertions.Equal(
		"(&(!(employeeType=external employee))(|(title=*engineer*)(&(co=US)(|(rhatCostCenter=123)(rhatCostCenter=456)))))",
		filter,
	)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_ExceedsMaxDepth() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	level4 := &v1alpha1.LDAPNestedQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{Key: "co", Criteria: "equals", Value: "US"},
		},
	}
	level3 := &v1alpha1.LDAPNestedQuery{
		Operator: "or",
		Filters:  []v1alpha1.LDAPFilter{{LDAPQuery: level4}},
	}
	level2 := &v1alpha1.LDAPNestedQuery{
		Operator: "and",
		Filters:  []v1alpha1.LDAPFilter{{LDAPQuery: level3}},
	}
	level1 := &v1alpha1.LDAPNestedQuery{
		Operator: "or",
		Filters:  []v1alpha1.LDAPFilter{{LDAPQuery: level2}},
	}
	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters:  []v1alpha1.LDAPFilter{{LDAPQuery: level1}},
	}

	_, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.Error(err)
	assertions.Contains(err.Error(), "exceeds maximum depth")
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_ManagerOptionsIgnoredInFilterString() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "manager",
				Criteria: "equals",
				Value:    "mgrAlpha",
				Options: &v1alpha1.LDAPFilterOptions{
					IncludeIndirectReports: true,
					IncludeManager:         true,
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.NoError(err)
	assertions.Equal("(&(manager=uid=mgrAlpha,ou=users,dc=redhat,dc=com))", filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_NonManagerOptionsIgnoredInFilterString() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Key:      "title",
				Criteria: "contains",
				Value:    "engineer",
				Options: &v1alpha1.LDAPFilterOptions{
					IncludeIndirectReports: true,
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.NoError(err)
	assertions.Equal("(&(title=*engineer*))", filter)
}

func (suite *LDAPTestSuite) TestBuildLDAPQueryString_NestedQueryItemOptionsIgnoredInFilterString() {
	assertions := assert.New(suite.T())

	ldapConn := &LDAPConn{
		baseUserDN: "ou=users,dc=redhat,dc=com",
	}

	query := &v1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []v1alpha1.LDAPFilter{
			{
				Options: &v1alpha1.LDAPFilterOptions{IncludeManager: true},
				LDAPQuery: &v1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []v1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					},
				},
			},
		},
	}

	filter, err := ldapConn.BuildLDAPQueryString(suite.ctx, query)
	assertions.NoError(err)
	assertions.Equal("(&(|(manager=uid=mgrAlpha,ou=users,dc=redhat,dc=com)))", filter)
}
