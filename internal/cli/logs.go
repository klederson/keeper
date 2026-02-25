package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
	"github.com/klederson/keeper/internal/ui"
)

var logsTail bool

var logsCmd = &cobra.Command{
	Use:   "logs [job]",
	Short: "Show backup logs",
	Long:  "Show recent backup run logs. Specify a job name to filter.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := reporter.NewStore()

		if len(args) > 0 {
			jobName := args[0]
			if job, _ := cfg.FindJob(jobName); job == nil {
				return fmt.Errorf("job %q not found", jobName)
			}
			return showJobLogs(store, jobName)
		}

		return showAllLogs(store)
	},
}

func showJobLogs(store *reporter.Store, jobName string) error {
	records := store.GetJobRecords(jobName, 20)

	if len(records) == 0 {
		fmt.Println(ui.Info(fmt.Sprintf("No logs for job %q", jobName)))
		return nil
	}

	fmt.Println(ui.Section("Logs: " + jobName))

	columns := []ui.TableColumn{
		{Title: "Date", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Duration", Width: 10},
		{Title: "Files", Width: 8},
		{Title: "Transferred", Width: 14},
		{Title: "Errors", Width: 24},
	}

	rows := make([][]string, 0, len(records))
	for _, r := range records {
		status := ui.AccentStyle.Render("✓ success")
		if !r.Success {
			status = ui.ErrorStyle.Render("✗ failed")
		}
		if r.DryRun {
			status = ui.MutedStyle.Render("~ dry-run")
		}

		errMsg := ""
		if len(r.Errors) > 0 {
			errMsg = r.Errors[0]
			if len(errMsg) > 22 {
				errMsg = errMsg[:22] + "..."
			}
		}

		rows = append(rows, []string{
			r.StartedAt.Format(time.DateTime),
			status,
			formatDuration(r.CompletedAt.Sub(r.StartedAt)),
			fmt.Sprintf("%d", r.FilesTransferred),
			formatBytes(r.BytesTransferred),
			errMsg,
		})
	}

	fmt.Println(ui.Table(columns, rows))
	return nil
}

func showAllLogs(store *reporter.Store) error {
	records := store.GetRecentRecords(30)

	if len(records) == 0 {
		fmt.Println(ui.Info("No backup logs yet"))
		return nil
	}

	fmt.Println(ui.Section("Recent Activity"))

	for _, r := range records {
		icon := ui.StatusIcon(r.Success)
		if r.DryRun {
			icon = ui.MutedStyle.Render("~")
		}

		timeStr := ui.MutedStyle.Render(formatTimeAgo(r.CompletedAt))
		name := ui.SubtitleStyle.Render(fmt.Sprintf("[%s]", r.JobName))
		detail := ""

		if r.Success {
			duration := formatDuration(r.CompletedAt.Sub(r.StartedAt))
			detail = ui.TextStyle.Render(fmt.Sprintf("%s in %s", formatBytes(r.BytesTransferred), duration))
		} else if len(r.Errors) > 0 {
			detail = ui.ErrorStyle.Render(r.Errors[0])
		}

		fmt.Printf("  %s %s %s %s\n", timeStr, name, icon, detail)
	}

	fmt.Println()
	return nil
}

func init() {
	logsCmd.Flags().BoolVar(&logsTail, "tail", false, "Follow logs in real time")
}
