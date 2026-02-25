package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new backup job",
	Long:  "Interactive form to configure and add a new backup job.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		job, err := ui.RunAddJobForm(nil)
		if err != nil {
			return err
		}

		if err := cfg.AddJob(*job); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(ui.Success(fmt.Sprintf("Job %q added successfully", job.Name)))
		fmt.Println(ui.Label("  Source", job.Sources[0].Path))
		fmt.Println(ui.Label("  Destination", fmt.Sprintf("%s@%s:%s", job.Destination.User, job.Destination.Host, job.Destination.Path)))
		fmt.Println(ui.Label("  Schedule", job.Schedule))
		fmt.Println()
		fmt.Println(ui.Info("Run 'keeper test " + job.Name + "' to verify the connection"))
		return nil
	},
}
