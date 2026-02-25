package backup

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/klederson/keeper/internal/backend"
	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

func RunJob(ctx context.Context, job *config.Job, dryRun bool) (*backend.Result, error) {
	b, err := backend.New(job.Destination.Type)
	if err != nil {
		return nil, fmt.Errorf("creating backend: %w", err)
	}

	if err := b.Validate(job); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	mode := "backup"
	if dryRun {
		mode = "dry-run"
	}

	slog.Info("starting job",
		"job", job.Name,
		"mode", mode,
		"backend", b.Name(),
	)

	result, err := b.Run(ctx, job, dryRun)
	if err != nil {
		return result, fmt.Errorf("backup failed: %w", err)
	}

	return result, nil
}

func PrintResult(jobName string, result *backend.Result, dryRun bool) {
	duration := result.CompletedAt.Sub(result.StartedAt).Round(time.Second)

	fmt.Println()
	if dryRun {
		fmt.Println(ui.Section("Dry Run Results: " + jobName))
	} else {
		fmt.Println(ui.Section("Backup Results: " + jobName))
	}

	status := ui.Success("completed successfully")
	if !result.Success {
		status = ui.Error("completed with errors")
	}
	fmt.Println("  " + status)
	fmt.Println()

	fmt.Println(ui.KeyValue([][2]string{
		{"Duration", duration.String()},
		{"Files total", fmt.Sprintf("%d", result.FilesTotal)},
		{"Files transferred", fmt.Sprintf("%d", result.FilesTransferred)},
		{"Total size", formatBytes(result.BytesTotal)},
		{"Transferred", formatBytes(result.BytesTransferred)},
	}))

	if len(result.Errors) > 0 {
		fmt.Println(ui.Section("Errors"))
		for _, e := range result.Errors {
			fmt.Println("  " + ui.Error(e))
		}
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
