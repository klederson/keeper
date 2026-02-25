package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Keeper configuration",
	Long:  "Interactive wizard to create the initial Keeper configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.Banner())

		// Check if config already exists
		if _, err := os.Stat(config.ConfigPath()); err == nil {
			fmt.Println(ui.Warn("Configuration already exists at " + config.ConfigPath()))
			fmt.Println(ui.Info("Use 'keeper add' to add backup jobs"))
			return nil
		}

		cfg, err := ui.RunInitForm()
		if err != nil {
			return err
		}

		if err := config.EnsureDataDir(); err != nil {
			return fmt.Errorf("creating data directory: %w", err)
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println(ui.Success("Config saved to " + config.ConfigPath()))
		return nil
	},
}
