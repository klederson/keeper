package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var editCmd = &cobra.Command{
	Use:   "edit <job>",
	Short: "Edit a backup job",
	Long:  "Interactive form to edit an existing backup job.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		job, idx := cfg.FindJob(args[0])
		if job == nil {
			return fmt.Errorf("job %q not found", args[0])
		}

		updated, err := ui.RunAddJobForm(job)
		if err != nil {
			return err
		}

		cfg.Jobs[idx] = *updated

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(ui.Success(fmt.Sprintf("Job %q updated", updated.Name)))
		return nil
	},
}
