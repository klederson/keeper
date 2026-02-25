package reporter

import (
	"github.com/klederson/keeper/internal/backend"
	"github.com/klederson/keeper/internal/config"
)

func ResultToRecord(jobName string, result *backend.Result, dryRun bool) config.RunRecord {
	return config.RunRecord{
		JobName:          jobName,
		StartedAt:        result.StartedAt,
		CompletedAt:      result.CompletedAt,
		Success:          result.Success,
		FilesTotal:       result.FilesTotal,
		FilesTransferred: result.FilesTransferred,
		BytesTotal:       result.BytesTotal,
		BytesTransferred: result.BytesTransferred,
		Errors:           result.Errors,
		DryRun:           dryRun,
	}
}
