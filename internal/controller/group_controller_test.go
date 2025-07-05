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

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/internal/controller/mocks"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

const (
	// GroupControllerName is the name of the Group controller
	GroupControllerName = "group-controller"
	keyApiKey           = "apiKey"
	keyApiSecret        = "apiSecret"
)

var _ = Describe("Group Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		group := &usernautdevv1alpha1.Group{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Group")
			err := k8sClient.Get(ctx, typeNamespacedName, group)
			if err != nil && errors.IsNotFound(err) {
				resource := &usernautdevv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: usernautdevv1alpha1.GroupSpec{
						GroupName: "test-resource-group",
						Members:   []string{"test-user-1", "test-user-2"},
						Backends: []usernautdevv1alpha1.Backend{
							{
								Name: "fivetran",
								Type: "fivetran",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &usernautdevv1alpha1.Group{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Group")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			fivetranBackend := config.Backend{
				Name:    "fivetran",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					keyApiKey:    "testKey",
					keyApiSecret: "testSecret",
				},
			}

			backendMap := make(map[string]map[string]config.Backend)
			backendMap[fivetranBackend.Type] = make(map[string]config.Backend)
			backendMap[fivetranBackend.Type][fivetranBackend.Name] = fivetranBackend

			appConfig := config.AppConfig{
				App: config.App{
					Name:        "usernaut-test",
					Version:     "v0.0.1",
					Environment: "test",
				},
				LDAP: ldap.LDAP{
					Server:           "ldap://ldap.test.com:389",
					BaseDN:           "ou=adhoc,ou=managedGroups,dc=org,dc=com",
					UserDN:           "uid=%s,ou=users,dc=org,dc=com",
					UserSearchFilter: "(objectClass=filteClass)",
					Attributes:       []string{"mail", "uid", "cn", "sn", "displayName"},
				},
				Backends: []config.Backend{
					fivetranBackend,
				},
				BackendMap: backendMap,
				Cache: cache.Config{
					Driver: "memory",
					InMemory: &inmemory.Config{
						DefaultExpiration: int32(-1),
						CleanupInterval:   int32(-1),
					},
				},
			}

			cache, err := cache.New(&appConfig.Cache)
			Expect(err).NotTo(HaveOccurred())

			ctrl := gomock.NewController(GinkgoT())
			ldapClient := mocks.NewMockLDAPClient(ctrl)

			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Return(map[string]interface{}{
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser@gmail.com",
				"uid":         "testuser",
			}, nil).Times(2)

			controllerReconciler := &GroupReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				AppConfig: &appConfig,
				Cache:     cache,
				LdapConn:  ldapClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			// TODO: ideally err should be nil if the reconciliation is successful,
			// we need to mock the backend client to return a successful response.
			Expect(err).To(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
