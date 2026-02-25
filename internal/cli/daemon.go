package cli

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/scheduler"
	"github.com/klederson/keeper/internal/ui"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the Keeper daemon",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the backup daemon",
	Long:  "Start the daemon in foreground. Use systemd or similar to run in background.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := config.EnsureDataDir(); err != nil {
			return err
		}

		// Write PID file
		pidPath := pidFilePath()
		if err := writePIDFile(pidPath); err != nil {
			return fmt.Errorf("writing PID file: %w", err)
		}
		defer os.Remove(pidPath)

		slog.Info("starting keeper daemon", "pid", os.Getpid())
		fmt.Println(ui.Success("Keeper daemon starting"))
		fmt.Println(ui.Label("  PID", fmt.Sprintf("%d", os.Getpid())))
		fmt.Println(ui.Label("  Jobs", fmt.Sprintf("%d", len(cfg.Jobs))))

		sched := scheduler.New()
		if err := sched.LoadFromConfig(cfg); err != nil {
			return fmt.Errorf("loading scheduler: %w", err)
		}

		sched.Start()
		fmt.Println(ui.Success("Scheduler running"))

		// Wait for signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGHUP:
				slog.Info("received SIGHUP, reloading config")
				fmt.Println(ui.Info("Reloading configuration..."))

				newCfg, err := config.Load()
				if err != nil {
					slog.Error("failed to reload config", "error", err)
					fmt.Println(ui.Error("Failed to reload: " + err.Error()))
					continue
				}

				sched.Stop()
				sched = scheduler.New()
				if err := sched.LoadFromConfig(newCfg); err != nil {
					slog.Error("failed to reload scheduler", "error", err)
				}
				sched.Start()
				fmt.Println(ui.Success("Configuration reloaded"))

			case syscall.SIGINT, syscall.SIGTERM:
				slog.Info("received shutdown signal", "signal", sig)
				fmt.Println()
				fmt.Println(ui.Info("Shutting down..."))
				sched.Stop()
				fmt.Println(ui.Success("Daemon stopped"))
				return nil
			}
		}
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := readPIDFile()
		if err != nil {
			return fmt.Errorf("daemon not running (no PID file)")
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("process %d not found", pid)
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("sending SIGTERM to %d: %w", pid, err)
		}

		fmt.Println(ui.Success(fmt.Sprintf("Sent stop signal to daemon (PID %d)", pid)))
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := readPIDFile()
		if err != nil {
			fmt.Println(ui.MutedStyle.Render("Daemon is not running"))
			return nil
		}

		// Check if process is alive
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println(ui.MutedStyle.Render("Daemon is not running (stale PID file)"))
			return nil
		}

		if err := process.Signal(syscall.Signal(0)); err != nil {
			fmt.Println(ui.MutedStyle.Render("Daemon is not running (stale PID file)"))
			os.Remove(pidFilePath())
			return nil
		}

		fmt.Println(ui.AccentStyle.Render("Daemon is running"))
		fmt.Println(ui.Label("  PID", fmt.Sprintf("%d", pid)))
		return nil
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
}

func pidFilePath() string {
	return filepath.Join(config.DataDir(), "keeper.pid")
}

func writePIDFile(path string) error {
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0644)
}

func readPIDFile() (int, error) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}
