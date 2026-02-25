package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "keeper",
	Short: "Backup daemon & CLI tool",
	Long:  ui.Banner(),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(doctorCmd)
}
