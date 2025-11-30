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
	"sync"

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
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

const (
	// GroupControllerName is the name of the Group controller
	GroupControllerName = "group-controller"
	keyApiKey           = "apiKey"
	keyApiSecret        = "apiSecret"
)

var _ = Describe("Group Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource-group"

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
						Members: usernautdevv1alpha1.Members{
							Groups: []string{},
							Users:  []string{"test-user-1", "test-user-2"},
						},
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
					keyApiKey: "testKey",
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
				Client:     k8sClient,
				Scheme:     k8sClient.Scheme(),
				AppConfig:  &appConfig,
				Store:      store.New(cache),
				LdapConn:   ldapClient,
				CacheMutex: &sync.RWMutex{},
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

		It("should handle multiple same-type backends independently", func() {
			By("creating a resource with two backends of the same type but different names")

			const multiName = "test-resource-group-multi"
			multiNN := types.NamespacedName{Name: multiName, Namespace: "default"}

			multi := &usernautdevv1alpha1.Group{
				ObjectMeta: metav1.ObjectMeta{
					Name:      multiName,
					Namespace: "default",
				},
				Spec: usernautdevv1alpha1.GroupSpec{
					GroupName: multiName,
					Members: usernautdevv1alpha1.Members{
						Groups: []string{},
						Users:  []string{"test-user-1", "test-user-2"},
					},
					Backends: []usernautdevv1alpha1.Backend{
						{Name: "fivetran-a", Type: "fivetran"},
						{Name: "fivetran-b", Type: "fivetran"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, multi)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, multi) }()

			fivetranA := config.Backend{
				Name:    "fivetran-a",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					keyApiKey: "testKeyA",
					// Intentionally omit apiSecret to force client creation error
				},
			}
			fivetranB := config.Backend{
				Name:    "fivetran-b",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					keyApiKey: "testKeyB",
					// Intentionally omit apiSecret to force client creation error
				},
			}

			backendMap := make(map[string]map[string]config.Backend)
			backendMap[fivetranA.Type] = make(map[string]config.Backend)
			backendMap[fivetranA.Type][fivetranA.Name] = fivetranA
			backendMap[fivetranB.Type][fivetranB.Name] = fivetranB

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
				Backends:   []config.Backend{fivetranA, fivetranB},
				BackendMap: backendMap,
				Cache: cache.Config{
					Driver: "memory",
					InMemory: &inmemory.Config{
						DefaultExpiration: int32(-1),
						CleanupInterval:   int32(-1),
					},
				},
			}

			Cache, err := cache.New(&appConfig.Cache)
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

			reconciler := &GroupReconciler{
				Client:     k8sClient,
				Scheme:     k8sClient.Scheme(),
				AppConfig:  &appConfig,
				Store:      store.New(Cache),
				LdapConn:   ldapClient,
				CacheMutex: &sync.RWMutex{},
			}

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: multiNN})
			Expect(err).To(HaveOccurred())

			// Reload the resource to inspect status
			fresh := &usernautdevv1alpha1.Group{}
			Expect(k8sClient.Get(ctx, multiNN, fresh)).To(Succeed())
			Expect(fresh.Status.BackendsStatus).To(HaveLen(2))

			statuses := map[string]usernautdevv1alpha1.BackendStatus{}
			for _, s := range fresh.Status.BackendsStatus {
				statuses[s.Name] = s
			}

			Expect(statuses).To(HaveKey("fivetran-a"))
			Expect(statuses).To(HaveKey("fivetran-b"))
			Expect(statuses["fivetran-a"].Status).To(BeFalse())
			Expect(statuses["fivetran-b"].Status).To(BeFalse())
			Expect(statuses["fivetran-a"].Message).To(ContainSubstring("missing required connection parameters"))
			Expect(statuses["fivetran-b"].Message).To(ContainSubstring("missing required connection parameters"))
		})
	})
})
