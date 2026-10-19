/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"slices"
	"strings"

	usernautv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

var (
	ldapFilterKeys = []string{
		"givenName", "displayName", "rhatJobTitle", "title", "employeeType", "manager",
		"rhatCostCenter", "rhatCostCenterDesc", "rhatGeo", "co", "st", "rhatLocation",
		"rhatOfficeLocation", "rhatOfficeFloor", "roomNumber",
	}
	ldapFilterCriteria = []string{"equals", "contains", "not"}
	ldapQueryOperators = []string{"and", "or"}
)

func validate(namespace string, group *usernautv1alpha1.Group, rules config.SpecValidationRulesConfig) error {
	if err := validateLDAPQuery(group.Spec.Members.LDAPQuery); err != nil {
		return err
	}

	nsRules, ok := rules[namespace]
	if !ok {
		return nil
	}
	if err := validateGroupName(group, nsRules.Group.GroupName); err != nil {
		return err
	}

	return validateBackends(group, nsRules.Group.Backends)
}

// validateLDAPQuery validates a root ldap_query (operator, filters, nesting).
func validateLDAPQuery(query *usernautv1alpha1.LDAPQuery) error {
	if query == nil {
		return nil
	}
	return validateLDAPQueryTree(query.Operator, query.Filters, "spec.members.ldap_query", 1)
}

func isManagerEqualsFilter(filter usernautv1alpha1.LDAPFilter) bool {
	return strings.EqualFold(strings.TrimSpace(filter.Key), "manager") &&
		strings.EqualFold(strings.TrimSpace(filter.Criteria), "equals")
}

func validateBackends(group *usernautv1alpha1.Group, rule config.BackendsValidationConfig) error {
	if len(rule.AllowedBackendTypes) == 0 {
		return nil
	}

	for _, backend := range group.Spec.Backends {
		if !slices.Contains(rule.AllowedBackendTypes, backend.Type) {
			return fmt.Errorf(
				"spec.backends type %q is not allowed; allowed backend types: %v",
				backend.Type,
				rule.AllowedBackendTypes,
			)
		}
	}

	return nil
}

func validateGroupName(group *usernautv1alpha1.Group, rule config.GroupNameValidationConfig) error {
	if rule.Prefix != "" && !strings.HasPrefix(group.Spec.GroupName, rule.Prefix) {
		return fmt.Errorf("spec.group_name must start with %q", rule.Prefix)
	}

	return nil
}

// validateLDAPNestedQuery walks a nested ldap_query tree. Nested ldap_query items are
// schemaless in the CRD, so this validates operators, filters, and filter items.
func validateLDAPNestedQuery(query *usernautv1alpha1.LDAPNestedQuery, path string, depth int) error {
	if query == nil {
		return nil
	}
	return validateLDAPQueryTree(query.Operator, query.Filters, path, depth)
}

// validateLDAPQueryTree validates operator, non-empty filters, depth, and each filter item.
func validateLDAPQueryTree(operator string, filters []usernautv1alpha1.LDAPFilter, path string, depth int) error {
	if depth > usernautv1alpha1.MaxLDAPQueryDepth {
		return fmt.Errorf("%s: ldap query nesting exceeds maximum depth of %d", path, usernautv1alpha1.MaxLDAPQueryDepth)
	}

	op := strings.ToLower(strings.TrimSpace(operator))
	if !slices.Contains(ldapQueryOperators, op) {
		return fmt.Errorf("%s: unsupported operator %q", path, operator)
	}

	if len(filters) == 0 {
		return fmt.Errorf("%s: filters are empty", path)
	}

	for i, filter := range filters {
		filterPath := fmt.Sprintf("%s.filters[%d]", path, i)
		if err := validateLDAPFilter(filter, filterPath, depth); err != nil {
			return err
		}
	}
	return nil
}

func validateLDAPFilter(filter usernautv1alpha1.LDAPFilter, path string, depth int) error {
	hasSimple := filter.Key != "" || filter.Criteria != "" || filter.Value != ""
	hasNested := filter.LDAPQuery != nil

	if hasSimple && hasNested {
		return fmt.Errorf("%s: filter item cannot have both key/criteria/value and ldap_query", path)
	}
	if !hasSimple && !hasNested {
		return fmt.Errorf("%s: filter item must have key/criteria/value or ldap_query", path)
	}

	if hasNested {
		if filter.Options != nil {
			return fmt.Errorf("%s: options is only allowed when key is manager and criteria is equals", path)
		}
		return validateLDAPNestedQuery(filter.LDAPQuery, path+".ldap_query", depth+1)
	}

	key := strings.TrimSpace(filter.Key)
	if key == "" {
		return fmt.Errorf("%s: filter key is empty", path)
	}
	if !slices.Contains(ldapFilterKeys, key) {
		return fmt.Errorf("%s: unsupported filter key %q", path, filter.Key)
	}

	criteria := strings.ToLower(strings.TrimSpace(filter.Criteria))
	if !slices.Contains(ldapFilterCriteria, criteria) {
		return fmt.Errorf("%s: unsupported filter criteria %q", path, filter.Criteria)
	}

	if filter.Options != nil && !isManagerEqualsFilter(filter) {
		return fmt.Errorf("%s: options is only allowed when key is manager and criteria is equals", path)
	}

	if strings.TrimSpace(filter.Value) == "" {
		return fmt.Errorf("%s: filter value is empty", path)
	}
	return nil
}
