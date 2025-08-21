package periodicjobs

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	UpdatePlatformAdminRoleName     = "update_platform_admin_role"
	UpdatePlatformAdminRoleInterval = 60 * time.Minute
)

type UpdatePlatformAdminRoleJob struct {
	snowflakeEnvironment string
}

func NewUpdatePlatformAdminRoleJob(snowflakeEnvironment string) *UpdatePlatformAdminRoleJob {
	return &UpdatePlatformAdminRoleJob{
		snowflakeEnvironment: snowflakeEnvironment,
	}
}

// add the job to the periodic task manager
func (upar *UpdatePlatformAdminRoleJob) AddToPeriodicTaskManager(mgr *PeriodicTaskManager) {
	mgr.AddTask(upar)
}

func (*UpdatePlatformAdminRoleJob) GetInterval() time.Duration {
	return UpdatePlatformAdminRoleInterval
}

func (*UpdatePlatformAdminRoleJob) GetName() string {
	return UpdatePlatformAdminRoleName
}

func (*UpdatePlatformAdminRoleJob) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("")

	// add databases to Platform Admin Role
	return nil
}
