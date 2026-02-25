package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var removeCmd = &cobra.Command{
	Use:     "remove <job>",
	Short:   "Remove a backup job",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		jobName := args[0]
		if _, idx := cfg.FindJob(jobName); idx < 0 {
			return fmt.Errorf("job %q not found", jobName)
		}

		confirmed, err := ui.ConfirmRemove(jobName)
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Println(ui.Info("Cancelled"))
			return nil
		}

		if err := cfg.RemoveJob(jobName); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println(ui.Success(fmt.Sprintf("Job %q removed", jobName)))
		return nil
	},
}
