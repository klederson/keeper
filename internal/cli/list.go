package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
	"github.com/klederson/keeper/internal/ui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backup jobs",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Jobs) == 0 {
			fmt.Println(ui.Info("No backup jobs configured"))
			fmt.Println(ui.Info("Run 'keeper add' to create one"))
			return nil
		}

		store := reporter.NewStore()

		columns := []ui.TableColumn{
			{Title: "Name", Width: 18},
			{Title: "Source", Width: 28},
			{Title: "Destination", Width: 30},
			{Title: "Schedule", Width: 16},
			{Title: "Last Run", Width: 14},
		}

		rows := make([][]string, 0, len(cfg.Jobs))
		for _, job := range cfg.Jobs {
			source := ""
			if len(job.Sources) > 0 {
				source = job.Sources[0].Path
				if len(job.Sources) > 1 {
					source += fmt.Sprintf(" (+%d)", len(job.Sources)-1)
				}
			}

			dest := fmt.Sprintf("%s@%s:%s",
				job.Destination.User,
				job.Destination.Host,
				job.Destination.Path,
			)

			// Truncate long paths
			if len(source) > 26 {
				source = "..." + source[len(source)-23:]
			}
			if len(dest) > 28 {
				dest = "..." + dest[len(dest)-25:]
			}

			lastRun := ui.MutedStyle.Render("never")
			records := store.GetJobRecords(job.Name, 1)
			if len(records) > 0 {
				r := records[0]
				icon := ui.StatusIcon(r.Success)
				lastRun = icon + " " + formatTimeAgo(r.CompletedAt)
			}

			rows = append(rows, []string{
				job.Name,
				source,
				dest,
				job.Schedule,
				lastRun,
			})
		}

		fmt.Println(ui.Section("Backup Jobs"))
		fmt.Println(ui.Table(columns, rows))
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  %d job(s) configured", len(cfg.Jobs))))

		return nil
	},
}
