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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
)

// spec.members is optional. An empty members block is a valid Group.
var _ = Describe("Group spec.members validation", func() {
	ctx := context.Background()

	minimalBackend := []usernautdevv1alpha1.Backend{
		{Name: "fivetran", Type: "fivetran"},
	}

	validLDAPQuery := &usernautdevv1alpha1.LDAPQuery{
		Operator: "and",
		Filters: []usernautdevv1alpha1.LDAPFilter{
			{Key: "employeeType", Criteria: "equals", Value: "employee"},
		},
	}

	newGroup := func(name string, members usernautdevv1alpha1.Members) *usernautdevv1alpha1.Group {
		return &usernautdevv1alpha1.Group{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: usernautdevv1alpha1.GroupSpec{
				GroupName: name,
				Members:   members,
				Backends:  minimalBackend,
			},
		}
	}

	It("reconciles successfully with no ldap_query and empty users list", func() {
		name := "group-members-accept-empty-users"
		g := newGroup(name, usernautdevv1alpha1.Members{
			Users: []string{},
		})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})

	It("reconciles successfully with no ldap_query, users, or groups", func() {
		name := "group-members-accept-all-empty"
		g := newGroup(name, usernautdevv1alpha1.Members{})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})

	It("accepts non-empty groups without users or ldap_query", func() {
		name := "group-members-accept-groups-only"
		g := newGroup(name, usernautdevv1alpha1.Members{
			Groups: []string{"some-group"},
		})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})

	It("accepts ldap_query with no users (query-only membership)", func() {
		name := "group-members-accept-query-only"
		g := newGroup(name, usernautdevv1alpha1.Members{
			LDAPQuery: validLDAPQuery,
		})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})

	It("accepts ldap_query with empty users slice", func() {
		name := "group-members-accept-query-empty-users"
		g := newGroup(name, usernautdevv1alpha1.Members{
			Users:     []string{},
			LDAPQuery: validLDAPQuery,
		})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})

	It("accepts non-empty users without ldap_query", func() {
		name := "group-members-accept-users-only"
		g := newGroup(name, usernautdevv1alpha1.Members{
			Users: []string{"alice"},
		})
		Expect(k8sClient.Create(ctx, g)).To(Succeed())
		DeferCleanup(func() {
			_ = k8sClient.Delete(ctx, g)
		})
	})
})
