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

func validate(namespace string, group *usernautv1alpha1.Group, rules config.SpecValidationRulesConfig) error {
	nsRules, ok := rules[namespace]
	if !ok {
		return nil
	}
	if err := validateGroupName(group, nsRules.Group.GroupName); err != nil {
		return err
	}

	return validateBackends(group, nsRules.Group.Backends)
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
