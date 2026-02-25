package backend

import (
	"context"
	"time"

	"github.com/klederson/keeper/internal/config"
)

type Result struct {
	StartedAt        time.Time
	CompletedAt      time.Time
	FilesTotal       int
	FilesTransferred int
	BytesTotal       int64
	BytesTransferred int64
	Errors           []string
	Success          bool
}

type BackupBackend interface {
	Run(ctx context.Context, job *config.Job, dryRun bool, onProgress func(ProgressEvent)) (*Result, error)
	Validate(job *config.Job) error
	Name() string
}

type ProgressEvent struct {
	CurrentFile string
	FilesCount  int
	Phase       string // "transferring", "stats", "done"
}
