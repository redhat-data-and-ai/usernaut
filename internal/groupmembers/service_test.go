package groupmembers

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap/types"
)

func irOptions() *usernautdevv1alpha1.LDAPFilterOptions {
	return &usernautdevv1alpha1.LDAPFilterOptions{IncludeIndirectReports: true}
}

func includeManagerOptions() *usernautdevv1alpha1.LDAPFilterOptions {
	return &usernautdevv1alpha1.LDAPFilterOptions{IncludeManager: true}
}

func TestReplaceManagerInFilters(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		{Key: "title", Criteria: "contains", Value: "engineer"},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)

	require.Len(t, got, 2)
	assert.Equal(t, "newMgr", got[0].Value)
	require.NotNil(t, got[0].Options)
	assert.True(t, got[0].Options.IncludeIndirectReports)
	assert.Equal(t, "engineer", got[1].Value)
	assert.Equal(t, "title", got[1].Key)
}

func TestReplaceManagerInFilters_LeavesManagerWithoutIndirectReports(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
		{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
		{Key: "title", Criteria: "contains", Value: "engineer"},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)

	require.Len(t, got, 3)
	assert.Equal(t, "mgrAlpha", got[0].Value)
	assert.Nil(t, got[0].Options)
	assert.Equal(t, "newMgr", got[1].Value)
	require.NotNil(t, got[1].Options)
	assert.True(t, got[1].Options.IncludeIndirectReports)
	assert.Equal(t, "engineer", got[2].Value)
}

func TestReplaceManagerInFilters_RootOptionsFallback(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
		{Key: "title", Criteria: "contains", Value: "engineer"},
	}

	got := replaceManagerInFilters(filters, "newMgr", &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true})

	require.Len(t, got, 2)
	assert.Equal(t, "newMgr", got[0].Value)
	assert.Equal(t, "engineer", got[1].Value)
}

func TestRootIncludeManagerDoesNotFallBackIntoNested(t *testing.T) {
	t.Parallel()

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeManager: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{
				LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
					Operator: "or",
					Filters: []usernautdevv1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
					},
				},
			},
		},
	}

	assert.Empty(t, extractManagersInfoFromQuery(query).UIDs())
	assert.False(t, queryHasIndirectReports(query))
}

func TestReplaceManagerInFilters_RootIRDoesNotRewriteNestedManager(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
				},
			},
		},
	}

	got := replaceManagerInFilters(filters, "newMgr", &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true})

	require.Len(t, got, 1)
	require.NotNil(t, got[0].LDAPQuery)
	require.Len(t, got[0].LDAPQuery.Filters, 1)
	assert.Equal(t, "mgrBeta", got[0].LDAPQuery.Filters[0].Value)
}

func TestReplaceManagerInFilters_DeduplicatesMatchingManagers(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)

	require.Len(t, got, 1)
	assert.Equal(t, "manager", got[0].Key)
	assert.Equal(t, "equals", got[0].Criteria)
	assert.Equal(t, "newMgr", got[0].Value)
}

func TestReplaceManagerInFilters_PreservesDistinctManagerCriteria(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		// contains is not eligible for filter options / IR rewrite
		{Key: "manager", Criteria: "contains", Value: "mgrAlpha"},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)

	require.Len(t, got, 2)
	assert.Equal(t, "equals", got[0].Criteria)
	assert.Equal(t, "contains", got[1].Criteria)
	assert.Equal(t, "newMgr", got[0].Value)
	assert.Equal(t, "mgrAlpha", got[1].Value)
}

func TestReplaceManagerInFilters_NestedQuery(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
				},
			},
		},
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrGamma", Options: irOptions()},
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrDelta", Options: irOptions()},
								{Key: "co", Criteria: "equals", Value: "US"},
							},
						},
					},
				},
			},
		},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)
	require.Len(t, got, 2)

	require.NotNil(t, got[0].LDAPQuery)
	assert.Len(t, got[0].LDAPQuery.Filters, 1)
	assert.Equal(t, "newMgr", got[0].LDAPQuery.Filters[0].Value)

	require.NotNil(t, got[1].LDAPQuery)
	assert.Equal(t, "newMgr", got[1].LDAPQuery.Filters[0].Value)
	require.NotNil(t, got[1].LDAPQuery.Filters[1].LDAPQuery)
	assert.Equal(t, "newMgr", got[1].LDAPQuery.Filters[1].LDAPQuery.Filters[0].Value)
	assert.Equal(t, "US", got[1].LDAPQuery.Filters[1].LDAPQuery.Filters[1].Value)
}

func TestReplaceManagerInFilters_CollapsesDuplicateManagersInNestedQuery(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
				},
			},
		},
	}

	got := replaceManagerInFilters(filters, "newMgr", nil)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].LDAPQuery)
	require.Len(t, got[0].LDAPQuery.Filters, 1)
	assert.Equal(t, "newMgr", got[0].LDAPQuery.Filters[0].Value)
}

func TestQueryHasManagerFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query *usernautdevv1alpha1.LDAPQuery
		want  bool
	}{
		{
			name:  "nil query",
			query: nil,
			want:  false,
		},
		{
			name: "top level manager",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
				},
			},
			want: true,
		},
		{
			name: "nested manager only",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "title", Criteria: "contains", Value: "engineer"},
								{
									LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
										Operator: "and",
										Filters: []usernautdevv1alpha1.LDAPFilter{
											{
												LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
													Operator: "or",
													Filters: []usernautdevv1alpha1.LDAPFilter{
														{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
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
			},
			want: true,
		},
		{
			name: "no manager anywhere",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "title", Criteria: "contains", Value: "engineer"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, queryHasManagerFilter(tt.query))
		})
	}
}

func TestQueryHasIndirectReports(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query *usernautdevv1alpha1.LDAPQuery
		want  bool
	}{
		{
			name:  "nil query",
			query: nil,
			want:  false,
		},
		{
			name: "manager without options",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
				},
			},
			want: false,
		},
		{
			name: "manager with include_indirect_reports",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
				},
			},
			want: true,
		},
		{
			name: "nested manager with include_indirect_reports",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "root include_indirect_reports when filter options are empty",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
				},
			},
			want: true,
		},
		{
			name: "filter options override empty root include_indirect_reports",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: &usernautdevv1alpha1.LDAPFilterOptions{}},
				},
			},
			want: false,
		},
		{
			name: "root include_indirect_reports does not fall back into nested manager filters",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "manager contains uses root include_indirect_reports fallback",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "contains", Value: "mgr"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, queryHasIndirectReports(tt.query))
		})
	}
}

func TestIncludeManagerUIDsFromQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query *usernautdevv1alpha1.LDAPQuery
		want  []string
	}{
		{
			name:  "nil query",
			query: nil,
			want:  []string{},
		},
		{
			name: "skips manager without include_manager",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: includeManagerOptions()},
				},
			},
			want: []string{"mgrBeta"},
		},
		{
			name: "top level only",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: includeManagerOptions()},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: includeManagerOptions()},
				},
			},
			want: []string{"mgrAlpha", "mgrBeta"},
		},
		{
			name: "nested and deduplicated",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: includeManagerOptions()},
								{
									LDAPQuery: &usernautdevv1alpha1.LDAPNestedQuery{
										Operator: "and",
										Filters: []usernautdevv1alpha1.LDAPFilter{
											{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: includeManagerOptions()},
											{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: includeManagerOptions()},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []string{"mgrAlpha", "mgrBeta"},
		},
		{
			name: "root include_manager when filter options are empty",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Options:  &usernautdevv1alpha1.LDAPQueryOptions{IncludeManager: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
				},
			},
			want: []string{"mgrAlpha", "mgrBeta"},
		},
		{
			name: "result include_only_managers does not include filter manager",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator:      "and",
				Options:       &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
				ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagers: true},
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "pbhattac"},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractManagersInfoFromQuery(tt.query).UIDs())
		})
	}
}

// TestBackwardCompatibleRootQueryOptionsCompat ensures legacy ldap_query.options still work:
// include_indirect_reports / include_manager on root options with a root manager
// filter that omits filter options.
func TestBackwardCompatibleRootQueryOptionsCompat(t *testing.T) {
	t.Parallel()

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Options: &usernautdevv1alpha1.LDAPQueryOptions{
			IncludeIndirectReports: true,
			IncludeManager:         true,
		},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}

	assert.True(t, queryHasIndirectReports(query))
	assert.Equal(t, []string{"pbhattac"}, extractManagersInfoFromQuery(query).UIDs())

	got := replaceManagerInFilters(query.Filters, "alice", query.Options)
	require.Len(t, got, 1)
	assert.Equal(t, "alice", got[0].Value)

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
			"alice":    {"carol"},
		},
	})
	members, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "bob", "carol", "pbhattac"}, members)
}

type orgLDAPStub struct {
	reports        map[string][]string
	managers       map[string]string
	bulkMissingUID map[string]struct{} // UIDs omitted from GetBulkUserLDAPData
}

func (s *orgLDAPStub) BuildLDAPQueryString(_ context.Context, query *usernautdevv1alpha1.LDAPQuery) (string, error) {
	return "mgr:" + managerUIDFromQuery(query), nil
}

func (s *orgLDAPStub) GetQueryMembers(_ context.Context, query string) (types.QueryMembers, error) {
	uid := strings.TrimPrefix(query, "mgr:")
	members, ok := s.reports[uid]
	if !ok {
		return types.QueryMembers{}, nil
	}
	result := make(types.QueryMembers, 0, len(members))
	for _, member := range members {
		qm := types.QueryMember{UID: member}
		if mgr, ok := s.managers[member]; ok {
			qm.Manager = mgr
		}
		result = append(result, qm)
	}
	return result, nil
}

func (s *orgLDAPStub) GetBulkUserLDAPData(_ context.Context, userIDs []string) (map[string]map[string]interface{}, error) {
	result := make(map[string]map[string]interface{}, len(userIDs))
	for _, uid := range userIDs {
		if _, skip := s.bulkMissingUID[uid]; skip {
			continue
		}
		userData := map[string]interface{}{"uid": uid}
		if mgr, ok := s.managers[uid]; ok {
			userData["manager"] = mgr
		}
		result[uid] = userData
	}
	return result, nil
}

func managerUIDFromQuery(query *usernautdevv1alpha1.LDAPQuery) string {
	if query == nil {
		return ""
	}
	return managerUIDFromFilters(query.Filters)
}

func managerUIDFromFilters(filters []usernautdevv1alpha1.LDAPFilter) string {
	for _, filter := range filters {
		if strings.EqualFold(strings.TrimSpace(filter.Key), "manager") {
			return strings.TrimSpace(filter.Value)
		}
		if filter.LDAPQuery != nil {
			if uid := managerUIDFromFilters(filter.LDAPQuery.Filters); uid != "" {
				return uid
			}
		}
	}
	return ""
}

func testOrgQuery(opts *usernautdevv1alpha1.LDAPFilterOptions) *usernautdevv1alpha1.LDAPQuery {
	return &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac", Options: opts},
		},
	}
}

func TestMembers_IncludeOnlyManagers(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
			"alice":    {"carol", "dave"},
			"dave":     {"eve"},
		},
	})

	allMembers, err := service.Members(context.Background(), testOrgQuery(irOptions()))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "bob", "carol", "dave", "eve"}, allMembers)

	managersOnlyQuery := &usernautdevv1alpha1.LDAPQuery{
		Operator:      "and",
		Options:       &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagers: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}
	managersOnly, err := service.Members(context.Background(), managersOnlyQuery)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "dave"}, managersOnly)
}

func TestMembers_IncludeOnlyManagersWithoutIR(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
			"alice":    {"carol", "dave"},
			"dave":     {"eve"},
		},
	})

	// Without IR: only direct hits (alice, bob). Among those, only alice has reports.
	query := &usernautdevv1alpha1.LDAPQuery{
		Operator:      "and",
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagers: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}
	got, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice"}, got)
}

func TestMembers_InheritedOnlyManagersExcludesFilterManager(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
			"alice":    {"carol", "dave"},
			"dave":     {"eve"},
		},
	})

	// result.include_only_managers alone does not add the filter manager UID.
	query := &usernautdevv1alpha1.LDAPQuery{
		Operator:      "and",
		Options:       &usernautdevv1alpha1.LDAPQueryOptions{IncludeIndirectReports: true},
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagers: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}

	got, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "dave"}, got)
}

func TestMembers_RootOptionsFallback(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
			"alice":    {"carol", "dave"},
			"dave":     {"eve"},
		},
	})

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Options: &usernautdevv1alpha1.LDAPQueryOptions{
			IncludeIndirectReports: true,
			IncludeManager:         true,
		},
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagers: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}

	got, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "dave", "pbhattac"}, got)
}

func TestMembers_IncludeOnlyManagersOfResult(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
		},
		managers: map[string]string{
			"alice": "uid=pbhattac,ou=users,dc=redhat,dc=com",
			"bob":   "uid=otherboss,ou=users,dc=redhat,dc=com",
		},
	})

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator:      "and",
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeOnlyManagersOfResult: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}

	got, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"pbhattac", "otherboss"}, got)
}

func TestMembers_IncludeManagersOfResult(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		reports: map[string][]string{
			"pbhattac": {"alice", "bob"},
		},
		managers: map[string]string{
			"alice": "uid=pbhattac,ou=users,dc=redhat,dc=com",
			"bob":   "uid=otherboss,ou=users,dc=redhat,dc=com",
		},
	})

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator:      "and",
		ResultOptions: &usernautdevv1alpha1.LDAPQueryResultOptions{IncludeManagersOfResult: true},
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "manager", Criteria: "equals", Value: "pbhattac"},
		},
	}

	got, err := service.Members(context.Background(), query)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "bob", "pbhattac", "otherboss"}, got)
}

func TestFetchUsersByUID_DropsUsersAbsentFromBulk(t *testing.T) {
	t.Parallel()

	service := New(&orgLDAPStub{
		managers: map[string]string{
			"alice": "uid=pbhattac,ou=users,dc=redhat,dc=com",
		},
		bulkMissingUID: map[string]struct{}{
			"ghost": {},
		},
	})

	got, err := service.fetchUsersByUID(context.Background(), []string{"alice", "ghost"})
	require.NoError(t, err)
	assert.Equal(t, types.QueryMembers{
		{UID: "alice", Manager: "uid=pbhattac,ou=users,dc=redhat,dc=com"},
	}, got)
}
