package periodicjobs_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-data-and-ai/usernaut/internal/controller/periodicjobs"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type MockPeriodicTask struct {
	name     string
	interval time.Duration
	runCount int
}

func (m *MockPeriodicTask) GetName() string {
	return m.name
}

func (m *MockPeriodicTask) GetInterval() time.Duration {
	return m.interval
}

func (m *MockPeriodicTask) Run(_ context.Context) error {
	m.runCount++
	return nil
}

var _ = Describe("PeriodicTaskManager", func() {
	var (
		manager *periodicjobs.PeriodicTaskManager
		ctx     context.Context
		cancel  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		manager = &periodicjobs.PeriodicTaskManager{
			Tasks: []periodicjobs.PeriodicTask{
				&MockPeriodicTask{name: "task1", interval: 100 * time.Millisecond},
				&MockPeriodicTask{name: "task2", interval: 200 * time.Millisecond},
			},
		}
	})

	AfterEach(func() {
		cancel()
	})

	It("should run all tasks at their specified intervals", func() {
		logger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
		ctx = log.IntoContext(ctx, logger)

		go func() {
			err := manager.RunAll(ctx)
			Expect(err).NotTo(HaveOccurred())
		}()

		// Allow some time for tasks to run
		time.Sleep(500 * time.Millisecond)

		for _, task := range manager.Tasks {
			mockTask, ok := task.(*MockPeriodicTask)
			Expect(ok).To(BeTrue())
			Expect(mockTask.runCount).To(BeNumerically(">", 0), "Task should have run at least once")
		}
	})

	It("should stop running tasks when context is canceled", func() {
		logger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
		ctx = log.IntoContext(ctx, logger)

		go func() {
			err := manager.RunAll(ctx)
			Expect(err).NotTo(HaveOccurred())
		}()

		// Allow some time for tasks to run
		time.Sleep(300 * time.Millisecond)

		// Cancel the context to stop tasks
		cancel()

		// Allow some time for tasks to cancel and stop
		time.Sleep(300 * time.Millisecond)

		// Capture the run count after cancellation
		runCounts := make(map[string]int)
		for _, task := range manager.Tasks {
			mockTask, ok := task.(*MockPeriodicTask)
			Expect(ok).To(BeTrue())
			runCounts[mockTask.GetName()] = mockTask.runCount
		}

		// Wait a bit to ensure no more runs occur
		time.Sleep(300 * time.Millisecond)

		for _, task := range manager.Tasks {
			mockTask, ok := task.(*MockPeriodicTask)
			Expect(ok).To(BeTrue())
			Expect(mockTask.runCount).To(Equal(runCounts[mockTask.GetName()]), "Task should not run after context is canceled")
		}
	})
})
