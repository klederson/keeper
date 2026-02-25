package cli

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
	"github.com/klederson/keeper/internal/ui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the interactive TUI dashboard",
	Aliases: []string{"dash"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := reporter.NewStore()
		model := ui.NewDashboard(cfg, store)

		p := tea.NewProgram(model)
		_, err = p.Run()
		return err
	},
}
