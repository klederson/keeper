package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/backup"
	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
	"github.com/klederson/keeper/internal/ui"
)

var runAll bool

var runCmd = &cobra.Command{
	Use:   "run [job]",
	Short: "Run a backup job now",
	Long:  "Execute a backup job immediately. Use --all to run all jobs.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		orch := backup.NewOrchestrator()
		store := reporter.NewStore()
		ctx := context.Background()

		if runAll {
			fmt.Println(ui.Info(fmt.Sprintf("Running all %d jobs...", len(cfg.Jobs))))
			fmt.Println()

			results := orch.RunAll(ctx, cfg.Jobs, false)
			for name, result := range results {
				backup.PrintResult(name, result, false)
				record := reporter.ResultToRecord(name, result, false)
				store.Append(record)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify a job name or use --all")
		}

		jobName := args[0]
		job, _ := cfg.FindJob(jobName)
		if job == nil {
			return fmt.Errorf("job %q not found", jobName)
		}

		fmt.Println(ui.Info(fmt.Sprintf("Running job %q...", jobName)))

		result, err := orch.Run(ctx, job, false)
		if err != nil {
			return err
		}

		backup.PrintResult(jobName, result, false)

		record := reporter.ResultToRecord(jobName, result, false)
		store.Append(record)

		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&runAll, "all", false, "Run all backup jobs")
}
