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
	// "encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	// "github.com/sirupsen/logrus"
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
	keyUrl              = "url"
	keyParentGroupId    = "parent_group_id"
)

var _ = Describe("Group Controller", func() {

	setupTestReconciler := func(backends []config.Backend) (*GroupReconciler, *mocks.MockLDAPClient) {
		backendMap := make(map[string]map[string]config.Backend)
		for _, backend := range backends {
			if _, ok := backendMap[backend.Type]; !ok {
				backendMap[backend.Type] = make(map[string]config.Backend)
			}
			backendMap[backend.Type][backend.Name] = backend
		}

		appConfig := &config.AppConfig{
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
			Backends:   backends,
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

		return &GroupReconciler{
			Client:     k8sClient,
			Scheme:     k8sClient.Scheme(),
			AppConfig:  appConfig,
			Store:      store.New(Cache),
			LdapConn:   ldapClient,
			CacheMutex: &sync.RWMutex{},
		}, ldapClient
	}

	setupSafeTestConfig := func() func() {
		tempDir, err := os.MkdirTemp("", "usernaut-test")
		Expect(err).NotTo(HaveOccurred())

		configDir := filepath.Join(tempDir, "appconfig")
		err = os.MkdirAll(configDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		// Create a temp default config
		configContent := ``
		err = os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte(configContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("WORKDIR", tempDir)
		Expect(err).NotTo(HaveOccurred())

		// Force reload of config to pick up the safe test config
		_, err = config.LoadConfig("default")
		Expect(err).NotTo(HaveOccurred())

		return func() {
			_ = os.Unsetenv("WORKDIR")
			_ = os.RemoveAll(tempDir)
		}
	}

	Context("When reconciling a resource", func() {
		const resourceName = "test-resource-group"

		// fetchLDAPData calls GetBulkUserLDAPData once; keys must match spec.members.users.
		bulkLDAPDataOK := map[string]map[string]interface{}{
			"test-user-1": {
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser@gmail.com",
				"uid":         "testuser",
			},
			"test-user-2": {
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser2@gmail.com",
				"uid":         "testuser2",
			},
		}

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
			controllerReconciler, ldapClient := setupTestReconciler([]config.Backend{fivetranBackend})

			ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), gomock.Any()).Return(bulkLDAPDataOK, nil).Times(1)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
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
			reconciler, ldapClient := setupTestReconciler([]config.Backend{fivetranA, fivetranB})

			ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), gomock.Any()).Return(bulkLDAPDataOK, nil).Times(1)

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: multiNN})
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

		It("should handle gitlab backend", func() {
			By("creating a resource with gitlab backend and group params")

			// Setup a temporary valid configuration to avoid panic in loader.go
			// when initializing backend clients which call config.GetConfig()
			cleanup := setupSafeTestConfig()
			defer cleanup()

			const gitlabResourceName = "test-resource-gitlab"
			gitlabNN := types.NamespacedName{Name: gitlabResourceName, Namespace: "default"}

			gitlabResource := &usernautdevv1alpha1.Group{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gitlabResourceName,
					Namespace: "default",
				},
				Spec: usernautdevv1alpha1.GroupSpec{
					GroupName: gitlabResourceName,
					Members: usernautdevv1alpha1.Members{
						Users: []string{"test-user-1", "test-user-2"},
					},
					Backends: []usernautdevv1alpha1.Backend{
						{
							Name: "gitlab-main",
							Type: "gitlab",
						},
					},
					GroupParams: []usernautdevv1alpha1.GroupParam{
						{
							Backend:  "gitlab",
							Name:     "gitlab-main",
							Property: "projects",
							Value:    []string{"my-group/my-project"},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, gitlabResource)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, gitlabResource) }()

			gitlabBackend := config.Backend{
				Name:    "gitlab-main",
				Type:    "gitlab",
				Enabled: true,
				Connection: map[string]interface{}{
					keyUrl:           "https://gitlab.com",
					keyParentGroupId: 123456,
				},
			}
			reconciler, ldapClient := setupTestReconciler([]config.Backend{gitlabBackend})

			ldapClient.EXPECT().GetBulkUserLDAPData(gomock.Any(), gomock.Any()).Return(bulkLDAPDataOK, nil).Times(1)

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: gitlabNN})
			Expect(err).To(HaveOccurred())

			// Reload the resource to inspect status
			fresh := &usernautdevv1alpha1.Group{}
			Expect(k8sClient.Get(ctx, gitlabNN, fresh)).To(Succeed())
			Expect(fresh.Status.BackendsStatus).To(HaveLen(1))

			status := fresh.Status.BackendsStatus[0]
			Expect(status.Name).To(Equal("gitlab-main"))
			Expect(status.Status).To(BeFalse())
			Expect(status.Message).To(ContainSubstring("missing required connection parameters"))
		})

		// DATA-3526: Tests for backend removal functionality
		Context("When backends are removed from the spec", func() {
			It("should detect removed backends correctly", func() {
				By("setting up a Group with existing backend status")

				const removedBackendsName = "test-removed-backends"

				// Create Group with multiple backends in status but fewer in spec
				groupWithRemovedBackends := &usernautdevv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{
						Name:      removedBackendsName,
						Namespace: "default",
					},
					Spec: usernautdevv1alpha1.GroupSpec{
						GroupName: removedBackendsName,
						Members: usernautdevv1alpha1.Members{
							Users: []string{"test-user"},
						},
						Backends: []usernautdevv1alpha1.Backend{
							{Name: "rover", Type: "rover"}, // Only rover in spec
						},
					},
					Status: usernautdevv1alpha1.GroupStatus{
						BackendsStatus: []usernautdevv1alpha1.BackendStatus{
							{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
							{Name: "fivetran", Type: "fivetran", Status: true, Message: "Successful"},   // Should be detected as removed
							{Name: "snowflake", Type: "snowflake", Status: true, Message: "Successful"}, // Should be detected as removed
						},
					},
				}

				appConfig := &config.AppConfig{}
				testCache, err := cache.New(&cache.Config{
					Driver: "memory",
					InMemory: &inmemory.Config{
						DefaultExpiration: int32(-1),
						CleanupInterval:   int32(-1),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				reconciler := &GroupReconciler{
					Client:     k8sClient,
					Scheme:     k8sClient.Scheme(),
					AppConfig:  appConfig,
					Store:      store.New(testCache),
					CacheMutex: &sync.RWMutex{},
					log:        logrus.NewEntry(logrus.New()),
				}

				By("detecting removed backends")
				removedBackends := reconciler.detectRemovedBackends(groupWithRemovedBackends)

				By("verifying that fivetran and snowflake are detected as removed")
				Expect(removedBackends).To(HaveLen(2))

				removedNames := make(map[string]bool)
				for _, backend := range removedBackends {
					removedNames[backend.Name] = true
				}

				Expect(removedNames).To(HaveKey("fivetran"))
				Expect(removedNames).To(HaveKey("snowflake"))
				Expect(removedNames).NotTo(HaveKey("rover")) // rover should NOT be detected as removed
			})

			It("should not detect any removed backends when specs match status", func() {
				By("setting up a Group where spec and status have same backends")

				groupMatching := &usernautdevv1alpha1.Group{
					Spec: usernautdevv1alpha1.GroupSpec{
						Backends: []usernautdevv1alpha1.Backend{
							{Name: "rover", Type: "rover"},
							{Name: "fivetran", Type: "fivetran"},
						},
					},
					Status: usernautdevv1alpha1.GroupStatus{
						BackendsStatus: []usernautdevv1alpha1.BackendStatus{
							{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
							{Name: "fivetran", Type: "fivetran", Status: true, Message: "Successful"},
						},
					},
				}

				reconciler := &GroupReconciler{
					log: logrus.NewEntry(logrus.New()),
				}

				By("detecting removed backends")
				removedBackends := reconciler.detectRemovedBackends(groupMatching)

				By("verifying no backends are detected as removed")
				Expect(removedBackends).To(BeEmpty())
			})

			It("should not detect removed backends when status is false", func() {
				By("setting up a Group with failed backends in status")

				groupWithFailedBackends := &usernautdevv1alpha1.Group{
					Spec: usernautdevv1alpha1.GroupSpec{
						Backends: []usernautdevv1alpha1.Backend{
							{Name: "rover", Type: "rover"},
						},
					},
					Status: usernautdevv1alpha1.GroupStatus{
						BackendsStatus: []usernautdevv1alpha1.BackendStatus{
							{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
							{Name: "fivetran", Type: "fivetran", Status: false, Message: "Failed"}, // Failed status, should not be removed
						},
					},
				}

				reconciler := &GroupReconciler{
					log: logrus.NewEntry(logrus.New()),
				}

				By("detecting removed backends")
				removedBackends := reconciler.detectRemovedBackends(groupWithFailedBackends)

				By("verifying no backends are detected as removed (failed backends are not considered for removal)")
				Expect(removedBackends).To(BeEmpty())
			})

			It("should handle empty status gracefully", func() {
				By("setting up a Group with empty status")

				groupEmptyStatus := &usernautdevv1alpha1.Group{
					Spec: usernautdevv1alpha1.GroupSpec{
						Backends: []usernautdevv1alpha1.Backend{
							{Name: "rover", Type: "rover"},
						},
					},
					Status: usernautdevv1alpha1.GroupStatus{
						BackendsStatus: []usernautdevv1alpha1.BackendStatus{}, // Empty status
					},
				}

				reconciler := &GroupReconciler{
					log: logrus.NewEntry(logrus.New()),
				}

				By("detecting removed backends")
				removedBackends := reconciler.detectRemovedBackends(groupEmptyStatus)

				By("verifying no backends are detected as removed")
				Expect(removedBackends).To(BeEmpty())
			})
		})

		// DATA-3526: Table-driven tests for detectRemovedBackends edge cases
		Context("When testing detectRemovedBackends with various scenarios", func() {
			type detectRemovedBackendsTestCase struct {
				name                 string
				specBackends         []usernautdevv1alpha1.Backend
				statusBackends       []usernautdevv1alpha1.BackendStatus
				expectedRemovedCount int
				expectedRemovedNames []string
			}

			DescribeTable("should detect removed backends correctly",
				func(tc detectRemovedBackendsTestCase) {
					group := &usernautdevv1alpha1.Group{
						Spec: usernautdevv1alpha1.GroupSpec{
							Backends: tc.specBackends,
						},
						Status: usernautdevv1alpha1.GroupStatus{
							BackendsStatus: tc.statusBackends,
						},
					}

					reconciler := &GroupReconciler{
						log: logrus.NewEntry(logrus.New()),
					}
					removedBackends := reconciler.detectRemovedBackends(group)

					Expect(removedBackends).To(HaveLen(tc.expectedRemovedCount))

					if tc.expectedRemovedCount > 0 {
						removedNames := make([]string, len(removedBackends))
						for i, backend := range removedBackends {
							removedNames[i] = backend.Name
						}
						Expect(removedNames).To(ConsistOf(tc.expectedRemovedNames))
					}
				},
				Entry("no backends in spec or status", detectRemovedBackendsTestCase{
					name:                 "empty spec and status",
					specBackends:         []usernautdevv1alpha1.Backend{},
					statusBackends:       []usernautdevv1alpha1.BackendStatus{},
					expectedRemovedCount: 0,
					expectedRemovedNames: []string{},
				}),
				Entry("single backend removed", detectRemovedBackendsTestCase{
					name: "single backend removed",
					specBackends: []usernautdevv1alpha1.Backend{
						{Name: "rover", Type: "rover"},
					},
					statusBackends: []usernautdevv1alpha1.BackendStatus{
						{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
						{Name: "fivetran", Type: "fivetran", Status: true, Message: "Successful"},
					},
					expectedRemovedCount: 1,
					expectedRemovedNames: []string{"fivetran"},
				}),
				Entry("multiple backends removed", detectRemovedBackendsTestCase{
					name: "multiple backends removed",
					specBackends: []usernautdevv1alpha1.Backend{
						{Name: "rover", Type: "rover"},
					},
					statusBackends: []usernautdevv1alpha1.BackendStatus{
						{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
						{Name: "fivetran", Type: "fivetran", Status: true, Message: "Successful"},
						{Name: "snowflake", Type: "snowflake", Status: true, Message: "Successful"},
					},
					expectedRemovedCount: 2,
					expectedRemovedNames: []string{"fivetran", "snowflake"},
				}),
				Entry("all backends removed", detectRemovedBackendsTestCase{
					name:         "all backends removed",
					specBackends: []usernautdevv1alpha1.Backend{},
					statusBackends: []usernautdevv1alpha1.BackendStatus{
						{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
						{Name: "fivetran", Type: "fivetran", Status: true, Message: "Successful"},
					},
					expectedRemovedCount: 2,
					expectedRemovedNames: []string{"rover", "fivetran"},
				}),
				Entry("failed backends not considered for removal", detectRemovedBackendsTestCase{
					name: "failed backends ignored",
					specBackends: []usernautdevv1alpha1.Backend{
						{Name: "rover", Type: "rover"},
					},
					statusBackends: []usernautdevv1alpha1.BackendStatus{
						{Name: "rover", Type: "rover", Status: true, Message: "Successful"},
						{Name: "fivetran", Type: "fivetran", Status: false, Message: "Failed"},
					},
					expectedRemovedCount: 0,
					expectedRemovedNames: []string{},
				}),
				Entry("mixed successful and failed backends", detectRemovedBackendsTestCase{
					name:         "mixed success and failure",
					specBackends: []usernautdevv1alpha1.Backend{},
					statusBackends: []usernautdevv1alpha1.BackendStatus{
						{Name: "rover", Type: "rover", Status: true, Message: "Successful"},      // Should be removed
						{Name: "fivetran", Type: "fivetran", Status: false, Message: "Failed"},   // Should NOT be removed
						{Name: "snowflake", Type: "snowflake", Status: true, Message: "Success"}, // Should be removed
					},
					expectedRemovedCount: 2,
					expectedRemovedNames: []string{"rover", "snowflake"},
				}),
			)
		})
	})
})
