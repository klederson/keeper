package backup

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/klederson/keeper/internal/backend"
	"github.com/klederson/keeper/internal/config"
)

type Orchestrator struct {
	mu       sync.Mutex
	running  map[string]context.CancelFunc
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		running: make(map[string]context.CancelFunc),
	}
}

func (o *Orchestrator) IsRunning(jobName string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	_, ok := o.running[jobName]
	return ok
}

func (o *Orchestrator) Run(ctx context.Context, job *config.Job, dryRun bool) (*backend.Result, error) {
	o.mu.Lock()
	if _, running := o.running[job.Name]; running {
		o.mu.Unlock()
		return nil, fmt.Errorf("job %q is already running", job.Name)
	}

	ctx, cancel := context.WithCancel(ctx)
	o.running[job.Name] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.running, job.Name)
		o.mu.Unlock()
	}()

	return RunJob(ctx, job, dryRun)
}

func (o *Orchestrator) RunAll(ctx context.Context, jobs []config.Job, dryRun bool) map[string]*backend.Result {
	results := make(map[string]*backend.Result)
	var mu sync.Mutex

	for i := range jobs {
		job := &jobs[i]

		slog.Info("running job", "job", job.Name)

		result, err := o.Run(ctx, job, dryRun)
		if err != nil {
			slog.Error("job failed", "job", job.Name, "error", err)
			result = &backend.Result{
				Errors:  []string{err.Error()},
				Success: false,
			}
		}

		mu.Lock()
		results[job.Name] = result
		mu.Unlock()
	}

	return results
}

func (o *Orchestrator) Cancel(jobName string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if cancel, ok := o.running[jobName]; ok {
		cancel()
	}
}
