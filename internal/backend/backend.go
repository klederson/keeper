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
	Run(ctx context.Context, job *config.Job, dryRun bool) (*Result, error)
	Validate(job *config.Job) error
	Name() string
}

type ProgressFunc func(transferred int64, total int64, currentFile string)
