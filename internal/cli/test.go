package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/backend"
	"github.com/klederson/keeper/internal/backup"
	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var testCmd = &cobra.Command{
	Use:   "test <job>",
	Short: "Test a backup job (dry-run)",
	Long:  "Execute rsync with --dry-run to verify the backup without transferring files.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		jobName := args[0]
		job, _ := cfg.FindJob(jobName)
		if job == nil {
			return fmt.Errorf("job %q not found", jobName)
		}

		printJobHeader(job)
		fmt.Println(ui.Warn("Dry-run mode — no files will be transferred"))
		fmt.Println()

		ctx := context.Background()
		orch := backup.NewOrchestrator()
		result, err := orch.Run(ctx, job, true, func(evt backend.ProgressEvent) {
			printProgress(jobName, evt)
		})
		clearProgress()

		if err != nil {
			return err
		}

		backup.PrintResult(jobName, result, true)
		return nil
	},
}
