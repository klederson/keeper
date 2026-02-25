package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
	"github.com/klederson/keeper/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all backup jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := reporter.NewStore()
		allRecords := store.LoadAll()
		stats30d := reporter.CalculateStats(allRecords, time.Now().AddDate(0, 0, -30))

		fmt.Println(ui.Section("Keeper Status"))

		// Overall stats
		fmt.Println(ui.KeyValue([][2]string{
			{"Success rate (30d)", fmt.Sprintf("%.1f%%", stats30d.SuccessRate)},
			{"Total transferred", formatBytes(stats30d.TotalBytes)},
			{"Avg duration", formatDuration(stats30d.AvgDuration)},
			{"Jobs run (30d)", fmt.Sprintf("%d", stats30d.TotalRuns)},
		}))

		fmt.Println(ui.Section("Jobs"))

		columns := []ui.TableColumn{
			{Title: "Name", Width: 16},
			{Title: "Schedule", Width: 14},
			{Title: "Last Run", Width: 12},
			{Title: "Status", Width: 12},
			{Title: "Duration", Width: 10},
			{Title: "Transferred", Width: 14},
		}

		rows := make([][]string, 0, len(cfg.Jobs))
		for _, job := range cfg.Jobs {
			records := store.GetJobRecords(job.Name, 1)

			lastRun := ui.MutedStyle.Render("never")
			status := ui.MutedStyle.Render("—")
			duration := ui.MutedStyle.Render("—")
			transferred := ui.MutedStyle.Render("—")

			if len(records) > 0 {
				r := records[0]
				lastRun = formatTimeAgo(r.CompletedAt)
				if r.Success {
					status = ui.AccentStyle.Render("✓ success")
				} else {
					status = ui.ErrorStyle.Render("✗ failed")
				}
				duration = formatDuration(r.CompletedAt.Sub(r.StartedAt))
				transferred = formatBytes(r.BytesTransferred)
			}

			rows = append(rows, []string{
				job.Name,
				job.Schedule,
				lastRun,
				status,
				duration,
				transferred,
			})
		}

		fmt.Println(ui.Table(columns, rows))
		return nil
	},
}
