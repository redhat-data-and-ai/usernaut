package groupmembers

import (
	"context"

	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap/types"
)

// addManagersFromQuery pulls include_manager UIDs from the query and merges them
// into the member and manager result sets. When of-result options are set, UIDs
// are verified via bulk LDAP first (absent users are dropped).
func (s *Service) addManagersFromQuery(
	ctx context.Context,
	query *v1alpha1.LDAPQuery,
	members, managers types.QueryMembers,
) (types.QueryMembers, types.QueryMembers, error) {
	managersInQuery := extractManagersInfoFromQuery(query)
	if len(managersInQuery) == 0 {
		return members, managers, nil
	}

	resultOptions := queryResultOptions(query)

	if resultOptions.IncludeOnlyManagersOfResult || resultOptions.IncludeManagersOfResult {
		var err error
		managersInQuery, err = s.fetchUsersByUID(ctx, managersInQuery.UIDs())
		if err != nil {
			return nil, nil, err
		}
	}

	return append(members, managersInQuery...), append(managers, managersInQuery...), nil
}

// applyResultOptions applies include_only_managers and of-result result_options.
func applyResultOptions(members, managers types.QueryMembers, query *v1alpha1.LDAPQuery) types.QueryMembers {
	opts := queryResultOptions(query)

	if opts.IncludeOnlyManagers {
		members = managers
	}

	if opts.IncludeManagersOfResult {
		members = append(members, members.Managers()...)
	}

	if opts.IncludeOnlyManagersOfResult {
		return members.Managers()
	}

	return members
}
