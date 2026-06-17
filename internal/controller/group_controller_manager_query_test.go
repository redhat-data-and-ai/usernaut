package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
)

func TestReplaceManagerInFilters(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
		{Key: "title", Criteria: "contains", Value: "engineer"},
	}

	got := replaceManagerInFilters(filters, "newMgr")

	require.Len(t, got, 2)
	assert.Equal(t, "newMgr", got[0].Value)
	assert.Equal(t, "engineer", got[1].Value)
	assert.Equal(t, "title", got[1].Key)
}

func TestReplaceManagerInFilters_DeduplicatesMatchingManagers(t *testing.T) {
	t.Parallel()

	filters := []usernautdevv1alpha1.LDAPFilter{
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
		{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
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
		{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
		{Key: "manager", Criteria: "contains", Value: "mgrAlpha"},
	}

	got := replaceManagerInFilters(filters, "newMgr")

	require.Len(t, got, 2)
	assert.Equal(t, "equals", got[0].Criteria)
	assert.Equal(t, "contains", got[1].Criteria)
	assert.Equal(t, "newMgr", got[0].Value)
	assert.Equal(t, "newMgr", got[1].Value)
}

func TestReplaceManagerInSubQueries(t *testing.T) {
	t.Parallel()

	queries := []usernautdevv1alpha1.LDAPSubQuery{
		{
			Operator: "or",
			Filters: []usernautdevv1alpha1.LDAPFilter{
				{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
				{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
			},
			Queries: []usernautdevv1alpha1.LDAPLeafQuery{
				{
					Operator: "and",
					Filters: []usernautdevv1alpha1.LDAPFilter{
						{Key: "manager", Criteria: "equals", Value: "mgrGamma"},
					},
					Queries: []usernautdevv1alpha1.LDAPLeafSubQuery{
						{
							Operator: "or",
							Filters: []usernautdevv1alpha1.LDAPFilter{
								{Key: "manager", Criteria: "equals", Value: "mgrDelta"},
								{Key: "co", Criteria: "equals", Value: "US"},
							},
						},
					},
				},
			},
		},
	}

	got := replaceManagerInSubQueries(queries, "newMgr")
	require.Len(t, got, 1)

	assert.Len(t, got[0].Filters, 1)
	assert.Equal(t, "newMgr", got[0].Filters[0].Value)

	require.Len(t, got[0].Queries, 1)
	assert.Equal(t, "newMgr", got[0].Queries[0].Filters[0].Value)

	require.Len(t, got[0].Queries[0].Queries, 1)
	assert.Equal(t, "newMgr", got[0].Queries[0].Queries[0].Filters[0].Value)
	assert.Equal(t, "US", got[0].Queries[0].Queries[0].Filters[1].Value)
}

func TestReplaceManagerInSubQueries_CollapsesDuplicateManagers(t *testing.T) {
	t.Parallel()

	query := &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Queries: []usernautdevv1alpha1.LDAPSubQuery{
			{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
				},
			},
		},
	}

	got := replaceManagerInSubQueries(query.Queries, "newMgr")
	require.Len(t, got, 1)
	require.Len(t, got[0].Filters, 1)
	assert.Equal(t, "newMgr", got[0].Filters[0].Value)
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
				Queries: []usernautdevv1alpha1.LDAPSubQuery{
					{
						Operator: "or",
						Filters: []usernautdevv1alpha1.LDAPFilter{
							{Key: "title", Criteria: "contains", Value: "engineer"},
						},
						Queries: []usernautdevv1alpha1.LDAPLeafQuery{
							{
								Operator: "and",
								Queries: []usernautdevv1alpha1.LDAPLeafSubQuery{
									{
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
			want:  nil,
		},
		{
			name: "top level only",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "or",
				Filters: []usernautdevv1alpha1.LDAPFilter{
					{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
					{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
				},
			},
			want: []string{"mgrAlpha", "mgrBeta"},
		},
		{
			name: "nested and deduplicated",
			query: &usernautdevv1alpha1.LDAPQuery{
				Operator: "and",
				Queries: []usernautdevv1alpha1.LDAPSubQuery{
					{
						Operator: "or",
						Filters: []usernautdevv1alpha1.LDAPFilter{
							{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
						},
						Queries: []usernautdevv1alpha1.LDAPLeafQuery{
							{
								Operator: "and",
								Filters: []usernautdevv1alpha1.LDAPFilter{
									{Key: "manager", Criteria: "equals", Value: "mgrAlpha"},
									{Key: "manager", Criteria: "equals", Value: "mgrBeta"},
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
