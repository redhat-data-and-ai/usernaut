package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
)

func irOptions() *usernautdevv1alpha1.LDAPOptions {
	return &usernautdevv1alpha1.LDAPOptions{IncludeIndirectReports: true}
}

func includeManagerOptions() *usernautdevv1alpha1.LDAPOptions {
	return &usernautdevv1alpha1.LDAPOptions{IncludeManager: true}
}

func TestReplaceManagerInFilters(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		{Key: "title", Criteria: "contains", Value: "engineer"},
	}

	got := replaceManagerInFilters(filters, "newMgr")

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

	got := replaceManagerInFilters(filters, "newMgr")

	require.Len(t, got, 3)
	assert.Equal(t, "mgrAlpha", got[0].Value)
	assert.Nil(t, got[0].Options)
	assert.Equal(t, "newMgr", got[1].Value)
	require.NotNil(t, got[1].Options)
	assert.True(t, got[1].Options.IncludeIndirectReports)
	assert.Equal(t, "engineer", got[2].Value)
}

func TestReplaceManagerInFilters_DeduplicatesMatchingManagers(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
	}

	got := replaceManagerInFilters(filters, "newMgr")

	require.Len(t, got, 1)
	assert.Equal(t, "manager", got[0].Key)
	assert.Equal(t, "equals", got[0].Criteria)
	assert.Equal(t, "newMgr", got[0].Value)
}

func TestReplaceManagerInFilters_PreservesDistinctManagerCriteria(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
		{Key: "manager", Criteria: "contains", Value: "mgrAlpha", Options: irOptions()},
	}

	got := replaceManagerInFilters(filters, "newMgr")

	require.Len(t, got, 2)
	assert.Equal(t, "equals", got[0].Criteria)
	assert.Equal(t, "contains", got[1].Criteria)
	assert.Equal(t, "newMgr", got[0].Value)
	assert.Equal(t, "newMgr", got[1].Value)
}

func TestReplaceManagerInFilters_NestedQuery(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
				},
			},
		},
		{
			LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrGamma", Options: irOptions()},
					{
						LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
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

	got := replaceManagerInFilters(filters, "newMgr")
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
			LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: irOptions()},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta", Options: irOptions()},
				},
			},
		},
	}

	got := replaceManagerInFilters(filters, "newMgr")
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
						LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "title", Criteria: "contains", Value: "engineer"},
								{
									LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
										Operator: "and",
										Filters: []usernautdevv1alpha1.LDAPFilter{
											{
												LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
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
						LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, queryHasIndirectReports(tt.query))
		})
	}
}

func TestExtractManagerUIDsFromQuery(t *testing.T) {
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
						LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrAlpha", Options: includeManagerOptions()},
								{
									LDAPQuery: &usernautdevv1alpha1.LDAPQuery{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractManagerUIDsFromQuery(tt.query))
		})
	}
}
