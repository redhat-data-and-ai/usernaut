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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	usernautv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

var _ = Describe("Group spec validation", func() {
	const (
		aifNamespace      = "ddis-asteroid--usernaut-aif-rhplatformtest"
		standardNamespace = "ddis-asteroid--usernaut-rhplatformtest"
	)

	var specRules config.SpecValidationRulesConfig

	allowedBackendTypes := []string{"rover"}

	BeforeEach(func() {
		specRules = config.SpecValidationRulesConfig{
			aifNamespace: {
				Group: config.GroupSpecValidationRules{
					GroupName: config.GroupNameValidationConfig{
						Prefix: "aif-",
					},
					Backends: config.BackendsValidationConfig{
						AllowedBackendTypes: allowedBackendTypes,
					},
				},
			},
		}
	})

	newGroup := func(groupName string, backends ...usernautv1alpha1.Backend) *usernautv1alpha1.Group {
		return &usernautv1alpha1.Group{
			ObjectMeta: metav1.ObjectMeta{Name: groupName},
			Spec: usernautv1alpha1.GroupSpec{
				GroupName: groupName,
				Backends:  backends,
			},
		}
	}

	Describe("validateGroupName", func() {
		It("accepts a valid group_name prefix", func() {
			group := newGroup("aif-mygroup")

			err := validateGroupName(group, specRules[aifNamespace].Group.GroupName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("rejects an invalid group_name prefix", func() {
			group := newGroup("other-mygroup")

			err := validateGroupName(group, specRules[aifNamespace].Group.GroupName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`spec.group_name must start with "aif-"`))
		})

		It("allows any group_name when prefix is empty", func() {
			group := newGroup("other-mygroup")

			err := validateGroupName(group, config.GroupNameValidationConfig{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("validateBackends", func() {
		It("accepts backends in the allowed list", func() {
			group := newGroup("aif-mygroup",
				usernautv1alpha1.Backend{Name: "rover", Type: "rover"},
			)

			err := validateBackends(group, specRules[aifNamespace].Group.Backends)
			Expect(err).NotTo(HaveOccurred())
		})

		It("rejects backends not in the allowed list", func() {
			group := newGroup("aif-mygroup",
				usernautv1alpha1.Backend{Name: "snowflake", Type: "snowflake"},
			)

			err := validateBackends(group, specRules[aifNamespace].Group.Backends)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`spec.backends type "snowflake" is not allowed`))
		})

		It("allows any backends when allowedBackendTypes is empty", func() {
			group := newGroup("aif-mygroup",
				usernautv1alpha1.Backend{Name: "rover", Type: "rover"},
			)

			err := validateBackends(group, config.BackendsValidationConfig{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows an empty backends list", func() {
			group := newGroup("aif-mygroup")

			err := validateBackends(group, specRules[aifNamespace].Group.Backends)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("validate", func() {
		It("skips namespaces without configured rules", func() {
			group := newGroup("not-mygroup")

			err := validate(standardNamespace, group, specRules)
			Expect(err).NotTo(HaveOccurred())
		})

		It("rejects invalid group names in a configured namespace", func() {
			group := newGroup("not-mygroup")

			err := validate(aifNamespace, group, specRules)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`spec.group_name must start with "aif-"`))
		})

		It("accepts valid group names in a configured namespace", func() {
			group := newGroup("aif-mygroup")

			err := validate(aifNamespace, group, specRules)
			Expect(err).NotTo(HaveOccurred())
		})

		It("rejects disallowed backends in a configured namespace", func() {
			group := newGroup("aif-mygroup",
				usernautv1alpha1.Backend{Name: "snowflake", Type: "snowflake"},
			)

			err := validate(aifNamespace, group, specRules)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`spec.backends type "snowflake" is not allowed`))
		})
	})
})
