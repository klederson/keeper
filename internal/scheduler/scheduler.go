package scheduler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/klederson/keeper/internal/backup"
	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
)

type Scheduler struct {
	cron         *cron.Cron
	orchestrator *backup.Orchestrator
	store        *reporter.Store
	mu           sync.Mutex
	entries      map[string]cron.EntryID
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		orchestrator: backup.NewOrchestrator(),
		store:        reporter.NewStore(),
		entries:      make(map[string]cron.EntryID),
	}
}

func (s *Scheduler) AddJob(job config.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobCopy := job
	entryID, err := s.cron.AddFunc(job.Schedule, func() {
		s.runJob(&jobCopy)
	})
	if err != nil {
		return err
	}

	s.entries[job.Name] = entryID
	slog.Info("scheduled job", "job", job.Name, "schedule", job.Schedule)
	return nil
}

func (s *Scheduler) RemoveJob(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.entries[name]; ok {
		s.cron.Remove(id)
		delete(s.entries, name)
		slog.Info("unscheduled job", "job", name)
	}
}

func (s *Scheduler) LoadFromConfig(cfg *config.Config) error {
	for _, job := range cfg.Jobs {
		if job.Schedule == "" {
			continue
		}
		if err := s.AddJob(job); err != nil {
			slog.Error("failed to schedule job", "job", job.Name, "error", err)
		}
	}
	return nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
	slog.Info("scheduler started", "jobs", len(s.entries))
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("scheduler stopped")
}

func (s *Scheduler) runJob(job *config.Job) {
	slog.Info("scheduler triggered job", "job", job.Name)

	ctx := context.Background()
	result, err := s.orchestrator.Run(ctx, job, false, nil)
	if err != nil {
		slog.Error("scheduled job failed", "job", job.Name, "error", err)
		return
	}

	record := reporter.ResultToRecord(job.Name, result, false)
	s.store.Append(record)

	if result.Success {
		slog.Info("scheduled job completed",
			"job", job.Name,
			"files", result.FilesTransferred,
			"bytes", result.BytesTransferred,
		)
	} else {
		slog.Warn("scheduled job completed with errors",
			"job", job.Name,
			"errors", result.Errors,
		)
	}
}
