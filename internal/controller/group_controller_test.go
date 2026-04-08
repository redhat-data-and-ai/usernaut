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
	"os"
	"path/filepath"
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
	keyUrl              = "url"
	keyParentGroupId    = "parent_group_id"
)

var _ = Describe("Group Controller", func() {

	setupTestReconciler := func(backends []config.Backend, cfgMutators ...func(*config.AppConfig)) (*GroupReconciler, *mocks.MockLDAPClient) {
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

		for _, m := range cfgMutators {
			m(appConfig)
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
					keyApiKey:    "testKey",
					keyApiSecret: "testSecret",
				},
			}
			controllerReconciler, ldapClient := setupTestReconciler([]config.Backend{fivetranBackend})

			// No backend name patterns: group is non-configurable, reconciler returns before LDAP fetch
			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Times(0)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should fetch LDAP user data when the group is configurable", func() {
			By("Reconciling with a pattern so isGroupConfigurable is true; LDAP runs before backend processing")

			// Use a dedicated CR so we do not race with the shared example's async delete (terminating object + reconcile would run handleDeletion).
			const ldapConfigurableName = "test-resource-group-ldap-configurable"
			ldapNN := types.NamespacedName{Name: ldapConfigurableName, Namespace: "default"}
			ldapGroup := &usernautdevv1alpha1.Group{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ldapConfigurableName,
					Namespace: "default",
				},
				Spec: usernautdevv1alpha1.GroupSpec{
					GroupName: "test-resource-group",
					Members: usernautdevv1alpha1.Members{
						Groups: []string{},
						Users:  []string{"test-user-1", "test-user-2"},
					},
					Backends: []usernautdevv1alpha1.Backend{
						{Name: "fivetran", Type: "fivetran"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, ldapGroup)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, ldapGroup) }()

			fivetranBackend := config.Backend{
				Name:    "fivetran",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					keyApiKey:    "testKey",
					keyApiSecret: "testSecret",
				},
			}
			withTestResourceGroupPattern := func(c *config.AppConfig) {
				c.Pattern = map[string][]config.PatternEntry{
					"fivetran": {{
						Input:  `^test-resource-group$`,
						Output: "test_resource_group",
					}},
				}
			}

			controllerReconciler, ldapClient := setupTestReconciler([]config.Backend{fivetranBackend}, withTestResourceGroupPattern)

			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Return(map[string]interface{}{
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser@gmail.com",
				"uid":         "testuser",
			}, nil).Times(2)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: ldapNN,
			})
			// gomock Times(2) asserts LDAP integration; backend may succeed (real API) or fail without connectivity
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("failed to reconcile all backends"))
			}
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

			// clients.New reads lowercase "apikey"/"apisecret" (see pkg/clients/client.go).
			// Omit apisecret so client creation fails after the group is deemed configurable.
			fivetranA := config.Backend{
				Name:    "fivetran-a",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					"apikey": "testKeyA",
				},
			}
			fivetranB := config.Backend{
				Name:    "fivetran-b",
				Type:    "fivetran",
				Enabled: true,
				Connection: map[string]interface{}{
					"apikey": "testKeyB",
				},
			}
			withMultiGroupPattern := func(c *config.AppConfig) {
				c.Pattern = map[string][]config.PatternEntry{
					"fivetran": {{
						Input:  `^test-resource-group-multi$`,
						Output: "test_resource_group_multi",
					}},
				}
			}
			reconciler, ldapClient := setupTestReconciler([]config.Backend{fivetranA, fivetranB}, withMultiGroupPattern)

			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Return(map[string]interface{}{
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser@gmail.com",
				"uid":         "testuser",
			}, nil).Times(2)

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: multiNN})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to reconcile all backends"))

			// Reload the resource to inspect per-backend status (clear errors for operators)
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
					keyParentGroupId: 111111,
				},
			}
			reconciler, ldapClient := setupTestReconciler([]config.Backend{gitlabBackend})

			// Since there are no matching patterns for gitlab backend, the group is non-configurable
			// and reconciliation returns without processing backends, so no LDAP calls expected
			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Times(0)

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: gitlabNN})
			Expect(err).NotTo(HaveOccurred())

			// Reload the resource to inspect status
			fresh := &usernautdevv1alpha1.Group{}
			Expect(k8sClient.Get(ctx, gitlabNN, fresh)).To(Succeed())
			// Since there are no matching patterns for gitlab backend, the group is non-configurable
			// and reconciliation returns without processing backends
			Expect(fresh.Status.ReconciledUsers).To(BeEmpty())

			// Verify the status condition is set correctly (print columns / API consumers rely on GroupReadyCondition)
			Expect(fresh.Status.Conditions).To(HaveLen(1))
			condition := fresh.Status.Conditions[0]
			Expect(condition.Type).To(Equal(usernautdevv1alpha1.GroupReadyCondition))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("NonConfigurable"))
			Expect(condition.Message).To(ContainSubstring("Group is not configurable"))
		})

		It("should surface gitlab connection validation on backend status when configurable", func() {
			By("GitLab NewClient requires url and token; missing token must appear on BackendsStatus")

			cleanup := setupSafeTestConfig()
			defer cleanup()

			const gitlabValName = "test-resource-gitlab-validation"
			gitlabValNN := types.NamespacedName{Name: gitlabValName, Namespace: "default"}

			gitlabValResource := &usernautdevv1alpha1.Group{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gitlabValName,
					Namespace: "default",
				},
				Spec: usernautdevv1alpha1.GroupSpec{
					GroupName: gitlabValName,
					Members: usernautdevv1alpha1.Members{
						Users: []string{"test-user-1", "test-user-2"},
					},
					Backends: []usernautdevv1alpha1.Backend{
						{Name: "gitlab-main", Type: "gitlab"},
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
			Expect(k8sClient.Create(ctx, gitlabValResource)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, gitlabValResource) }()

			// url set but token omitted — pkg/clients/gitlab.NewClient enforces both
			gitlabBackend := config.Backend{
				Name:    "gitlab-main",
				Type:    "gitlab",
				Enabled: true,
				Connection: map[string]interface{}{
					keyUrl:           "https://gitlab.com",
					keyParentGroupId: 111111,
				},
			}
			withGitlabValPattern := func(c *config.AppConfig) {
				c.Pattern = map[string][]config.PatternEntry{
					"gitlab": {{
						Input:  `^test-resource-gitlab-validation$`,
						Output: "test_resource_gitlab_validation",
					}},
				}
			}
			reconciler, ldapClient := setupTestReconciler([]config.Backend{gitlabBackend}, withGitlabValPattern)

			ldapClient.EXPECT().GetUserLDAPData(gomock.Any(), gomock.Any()).Return(map[string]interface{}{
				"cn":          "Test",
				"sn":          "User",
				"displayName": "Test User",
				"mail":        "testuser@gmail.com",
				"uid":         "testuser",
			}, nil).Times(2)

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: gitlabValNN})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to reconcile all backends"))

			fresh := &usernautdevv1alpha1.Group{}
			Expect(k8sClient.Get(ctx, gitlabValNN, fresh)).To(Succeed())
			Expect(fresh.Status.BackendsStatus).To(HaveLen(1))
			status := fresh.Status.BackendsStatus[0]
			Expect(status.Name).To(Equal("gitlab-main"))
			Expect(status.Status).To(BeFalse())
			Expect(status.Message).To(ContainSubstring("missing required connection parameters"))
		})
	})
})
