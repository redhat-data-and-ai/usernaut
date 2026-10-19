package groupmembers

import (
	"strings"

	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap/types"
)

// replaceManagerInFilters returns a copy of filters where manager filters that have
// include_indirect_reports are rewritten to managerUID. Other filters are left unchanged.
// Duplicate rewritten manager filters with the same criteria and value are collapsed.
func replaceManagerInFilters(
	filters []v1alpha1.LDAPFilter,
	managerUID string,
	inherited *v1alpha1.LDAPQueryOptions,
) []v1alpha1.LDAPFilter {
	if len(filters) == 0 {
		return []v1alpha1.LDAPFilter{}
	}
	result := make([]v1alpha1.LDAPFilter, 0, len(filters))
	seenManagers := make(map[string]struct{})
	for _, filter := range filters {
		if filter.LDAPQuery != nil {
			// Root options fallbacks do not apply inside nested ldap_query.
			nested := replaceManagerInNestedQuery(filter.LDAPQuery, managerUID)
			result = append(result, v1alpha1.LDAPFilter{
				LDAPQuery: nested,
			})
			continue
		}

		copied := copyLDAPFilter(filter)
		if !managerFilterHasIndirectReports(filter, inherited) {
			result = append(result, copied)
			continue
		}

		copied.Value = managerUID
		dedupeKey := strings.ToLower(strings.TrimSpace(filter.Criteria)) + "|" + managerUID
		if _, ok := seenManagers[dedupeKey]; ok {
			continue
		}
		seenManagers[dedupeKey] = struct{}{}
		result = append(result, copied)
	}
	return result
}

func replaceManagerInNestedQuery(
	query *v1alpha1.LDAPNestedQuery,
	managerUID string,
) *v1alpha1.LDAPNestedQuery {
	if query == nil {
		return nil
	}
	return &v1alpha1.LDAPNestedQuery{
		Operator: query.Operator,
		Filters:  replaceManagerInFilters(query.Filters, managerUID, nil),
	}
}

func copyLDAPFilter(filter v1alpha1.LDAPFilter) v1alpha1.LDAPFilter {
	copied := v1alpha1.LDAPFilter{
		Key:      filter.Key,
		Criteria: filter.Criteria,
		Value:    filter.Value,
	}
	if filter.Options != nil {
		opts := *filter.Options
		copied.Options = &opts
	}
	return copied
}

func isManagerFilter(filter v1alpha1.LDAPFilter) bool {
	return strings.EqualFold(strings.TrimSpace(filter.Key), "manager")
}

func managerFilterHasIndirectReports(filter v1alpha1.LDAPFilter, inherited *v1alpha1.LDAPQueryOptions) bool {
	if !isManagerFilter(filter) {
		return false
	}
	if filter.Options != nil {
		return filter.Options.IncludeIndirectReports
	}
	return inherited != nil && inherited.IncludeIndirectReports
}

func managerFilterHasIncludeManagers(filter v1alpha1.LDAPFilter, inherited *v1alpha1.LDAPQueryOptions) bool {
	if !isManagerFilter(filter) {
		return false
	}
	if filter.Options != nil {
		return filter.Options.IncludeManager
	}
	return inherited != nil && inherited.IncludeManager
}

func queryHasIndirectReports(query *v1alpha1.LDAPQuery) bool {
	if query == nil {
		return false
	}
	return filtersHaveIndirectReports(query.Filters, query.Options)
}

func queryResultOptions(query *v1alpha1.LDAPQuery) v1alpha1.LDAPQueryResultOptions {
	if query == nil || query.ResultOptions == nil {
		return v1alpha1.LDAPQueryResultOptions{}
	}
	return *query.ResultOptions
}

func filtersHaveIndirectReports(filters []v1alpha1.LDAPFilter, inherited *v1alpha1.LDAPQueryOptions) bool {
	for _, filter := range filters {
		if managerFilterHasIndirectReports(filter, inherited) {
			return true
		}
		if filter.LDAPQuery != nil && filtersHaveIndirectReports(filter.LDAPQuery.Filters, nil) {
			return true
		}
	}
	return false
}

func filtersHaveManager(filters []v1alpha1.LDAPFilter) bool {
	for _, filter := range filters {
		if strings.EqualFold(strings.TrimSpace(filter.Key), "manager") {
			return true
		}
		if filter.LDAPQuery != nil && filtersHaveManager(filter.LDAPQuery.Filters) {
			return true
		}
	}
	return false
}

// queryHasManagerFilter reports whether a manager filter exists anywhere in the query tree.
func queryHasManagerFilter(query *v1alpha1.LDAPQuery) bool {
	if query == nil {
		return false
	}
	return filtersHaveManager(query.Filters)
}

// includeManagerUIDsFromQuery returns UIDs from manager filters with include_manager
// (filter options or root options fallback). Nested filters do not inherit root options.
func extractManagersInfoFromQuery(query *v1alpha1.LDAPQuery) types.QueryMembers {
	if query == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out types.QueryMembers
	extractManagersInfoFromFilters(query.Filters, query.Options, seen, &out)
	return out
}

func extractManagersInfoFromFilters(
	filters []v1alpha1.LDAPFilter,
	inherited *v1alpha1.LDAPQueryOptions,
	seen map[string]struct{},
	result *types.QueryMembers,
) {
	for _, filter := range filters {
		if filter.LDAPQuery != nil {
			extractManagersInfoFromFilters(filter.LDAPQuery.Filters, nil, seen, result)
			continue
		}
		if !managerFilterHasIncludeManagers(filter, inherited) {
			continue
		}
		uid := strings.TrimSpace(filter.Value)
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		*result = append(*result, types.QueryMember{UID: uid})
	}
}
