package workers

import (
	"context"
	"time"
)

// Worker defines the interface for all background workers
type Worker interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

// BaseWorker provides common functionality for all workers
type BaseWorker struct {
	name     string
	interval time.Duration
}

func NewBaseWorker(name string, interval time.Duration) BaseWorker {
	return BaseWorker{
		name:     name,
		interval: interval,
	}
}

func (w BaseWorker) Name() string {
	return w.name
}

func (w BaseWorker) Interval() time.Duration {
	return w.interval
}
