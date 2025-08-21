package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/redhat-data-and-ai/usernaut/internal/controller/periodicjobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	periodicTasksControllerName = "periodictasks"
	defaultRequeueDelay         = 10 * time.Second
)

type PeriodicTasksReconciler struct {
	client.Client
	SnowflakeEnvironment string
	taskManager          *periodicjobs.PeriodicTaskManager
}

func NewPeriodicTasksReconciler(
	k8sClient client.Client,
) (*PeriodicTasksReconciler, error) {
	periodicTaskManager := periodicjobs.NewPeriodicTaskManager()

	// Add jobs to the periodic task manager
	userOffboardingJob, err := periodicjobs.NewUserOffboardingJob()
	if err != nil {
		return nil, fmt.Errorf("failed to create user offboarding job: %w", err)
	}
	userOffboardingJob.AddToPeriodicTaskManager(periodicTaskManager)

	return &PeriodicTasksReconciler{
		Client:      k8sClient,
		taskManager: periodicTaskManager,
	}, nil
}

// AddToManager will add the reconciler for the configured obj to a manager.
func (ptr *PeriodicTasksReconciler) AddToManager(mgr manager.Manager) error {
	return mgr.Add(ptr)
}

// Start the periodic tasks controller
// not event triggered like a conventional controller
// does not watch any kuberntes resources
// this is the platform through with periodic jobs get passed to controller manager
func (ptr *PeriodicTasksReconciler) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting periodic tasks controller")

	defer func() {
		logger.Info("Finishing periodic tasks controller")
	}()

	logger.Info("Initializing periodic tasks controller")

	logger.Info("Periodic tasks controller is enabled. Proceeding with initialization")

	// Wait for initialization delay or context cancellation
	select {
	case <-ctx.Done():
		logger.Info("Context canceled during initialization")
		return ctx.Err()
	case <-time.After(10 * time.Second):
		logger.Info("Periodic tasks ready to start after initialization delay")
	}

	logger.Info("Invoking task manager to run all periodic tasks")
	err := ptr.taskManager.RunAll(ctx)
	if err != nil {
		logger.Error(err, "Error occurred while running periodic tasks")
		return err
	}

	logger.Info("All periodic tasks have been started successfully")
	return nil
}
