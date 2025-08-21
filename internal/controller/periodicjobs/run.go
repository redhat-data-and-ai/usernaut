package periodicjobs

import (
	"context"
	"errors"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	syncOnceInterval time.Duration = 0
)

// RunAll runs all the tasks in the PeriodicTaskManager
// it launches a new goroutine for each task
// and stops when the context is canceled
// it uses a ticker to run the tasks at the specified interval (from the task)
// and stops when the context is canceled
// If interval == -1, the task is run only once.
func (p *PeriodicTaskManager) RunAll(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Running periodic tasks", "time", time.Now())

	for _, task := range p.Tasks {
		go p.runTask(ctx, task)
	}
	return nil
}

func (*PeriodicTaskManager) runTask(ctx context.Context, task PeriodicTask) {
	logger := log.FromContext(ctx)
	interval := task.GetInterval()

	run := func() {
		logger.Info("Running periodic task", "name", task.GetName(), "interval", interval)
		if err := task.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error(err, "error running periodic task", "periodic-task-name", task.GetName())
		}
	}

	if interval == syncOnceInterval {
		run()
		logger.Info("Task configured to run only once, exiting", "name", task.GetName())
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping periodic task", "name", task.GetName())
			return
		case <-ticker.C:
			run()
		}
	}
}
