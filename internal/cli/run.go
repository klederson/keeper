package cli

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/backend"
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

			results := orch.RunAll(ctx, cfg.Jobs, false, func(jobName string, evt backend.ProgressEvent) {
				printProgress(jobName, evt)
			})
			for name, result := range results {
				clearProgress()
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

		printJobHeader(job)

		result, err := orch.Run(ctx, job, false, func(evt backend.ProgressEvent) {
			printProgress(jobName, evt)
		})
		clearProgress()

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

func printJobHeader(job *config.Job) {
	src := ""
	if len(job.Sources) > 0 {
		src = job.Sources[0].Path
		if len(job.Sources) > 1 {
			src += fmt.Sprintf(" (+%d more)", len(job.Sources)-1)
		}
	}
	dest := fmt.Sprintf("%s@%s:%s", job.Destination.User, job.Destination.Host, job.Destination.Path)

	fmt.Println(ui.Info(fmt.Sprintf("Running job %q", job.Name)))
	fmt.Println(ui.Label("  Source", src))
	fmt.Println(ui.Label("  Dest", dest))
	fmt.Println()
}

var (
	progressMu   sync.Mutex
	spinnerIdx   int
	spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	lastLineLen  int
)

func printProgress(jobName string, evt backend.ProgressEvent) {
	progressMu.Lock()
	defer progressMu.Unlock()

	if evt.Phase == "done" {
		return
	}

	spinnerIdx = (spinnerIdx + 1) % len(spinnerChars)
	spinner := ui.AccentStyle.Render(spinnerChars[spinnerIdx])

	file := evt.CurrentFile
	if len(file) > 50 {
		file = "..." + file[len(file)-47:]
	}

	elapsed := time.Since(time.Now()).String() // placeholder
	_ = elapsed

	line := fmt.Sprintf("\r%s [%s] %s files | %s",
		spinner,
		ui.SubtitleStyle.Render(jobName),
		ui.TextStyle.Render(fmt.Sprintf("%d", evt.FilesCount)),
		ui.MutedStyle.Render(file),
	)

	// Pad with spaces to clear previous longer lines
	if len(line) < lastLineLen {
		line += strings.Repeat(" ", lastLineLen-len(line))
	}
	lastLineLen = len(line)

	fmt.Print(line)
}

func clearProgress() {
	progressMu.Lock()
	defer progressMu.Unlock()

	if lastLineLen > 0 {
		fmt.Print("\r" + strings.Repeat(" ", lastLineLen) + "\r")
		lastLineLen = 0
	}
}
