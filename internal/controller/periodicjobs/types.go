package periodicjobs

import (
	"context"
	"time"
)

type PeriodicTask interface {
	Run(ctx context.Context) error
	GetInterval() time.Duration
	GetName() string
}

type PeriodicTaskManager struct {
	Tasks []PeriodicTask
}

// NewPeriodicTaskManager creates a new PeriodicTaskManager
// used manage interval based tasks
func NewPeriodicTaskManager() *PeriodicTaskManager {
	return &PeriodicTaskManager{
		Tasks: []PeriodicTask{},
	}
}

func (p *PeriodicTaskManager) AddTask(task PeriodicTask) {
	p.Tasks = append(p.Tasks, task)
}
