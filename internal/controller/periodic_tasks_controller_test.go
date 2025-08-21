package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-data-and-ai/usernaut/internal/controller/periodicjobs"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("PeriodicTasks Controller", func() {
	Context("When starting the controller", func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
			mgr    manager.Manager
		)

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())

			By("setting up a new manager")
			var err error
			mgr, err = manager.New(cfg, manager.Options{})
			Expect(err).NotTo(HaveOccurred())

			By("setting up in-memory cache")
			cacheConfig := &inmemory.Config{
				DefaultExpiration: 60,
				CleanupInterval:   120,
			}
			inMemCache, err := inmemory.NewCache(cacheConfig)
			Expect(err).NotTo(HaveOccurred())

			By("creating a new PeriodicTasksReconciler with an empty task manager")
			reconciler := &PeriodicTasksReconciler{
				Client:      k8sClient,
				taskManager: periodicjobs.NewPeriodicTaskManager(), // Empty task manager
				cacheClient: inMemCache,
			}
			Expect(reconciler).NotTo(BeNil())

			By("adding the reconciler to the manager")
			err = reconciler.AddToManager(mgr)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			cancel()
		})

		It("should start without errors", func() {
			By("starting the manager")
			go func() {
				defer GinkgoRecover()
				err := mgr.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()
		})
	})
})
