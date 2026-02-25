package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/ui"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.Banner())
		fmt.Println(ui.Section("System Check"))

		allOk := true

		// Check rsync
		if path, err := exec.LookPath("rsync"); err != nil {
			fmt.Println(ui.Error("rsync not found in PATH"))
			fmt.Println(ui.Info("  Install: sudo apt install rsync (Debian/Ubuntu)"))
			allOk = false
		} else {
			out, _ := exec.Command("rsync", "--version").Output()
			version := firstLine(string(out))
			fmt.Println(ui.Success(fmt.Sprintf("rsync found: %s", path)))
			fmt.Println(ui.MutedStyle.Render("  " + version))
		}

		// Check ssh
		if path, err := exec.LookPath("ssh"); err != nil {
			fmt.Println(ui.Error("ssh not found in PATH"))
			allOk = false
		} else {
			fmt.Println(ui.Success(fmt.Sprintf("ssh found: %s", path)))
		}

		// Check config
		fmt.Println()
		fmt.Println(ui.Section("Configuration"))

		cfg, err := config.Load()
		if err != nil {
			fmt.Println(ui.Error("Config: " + err.Error()))
			allOk = false
		} else {
			fmt.Println(ui.Success(fmt.Sprintf("Config loaded: %d job(s)", len(cfg.Jobs))))

			if err := cfg.Validate(); err != nil {
				fmt.Println(ui.Error("Validation: " + err.Error()))
				allOk = false
			} else {
				fmt.Println(ui.Success("Config validation passed"))
			}

			// Check connectivity for each job destination
			if len(cfg.Jobs) > 0 {
				fmt.Println()
				fmt.Println(ui.Section("Connectivity"))

				seen := make(map[string]bool)
				for _, job := range cfg.Jobs {
					key := fmt.Sprintf("%s:%d", job.Destination.Host, job.Destination.Port)
					if seen[key] {
						continue
					}
					seen[key] = true

					port := job.Destination.Port
					if port == 0 {
						port = 22
					}
					addr := fmt.Sprintf("%s:%d", job.Destination.Host, port)

					conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
					if err != nil {
						fmt.Println(ui.Error(fmt.Sprintf("%s — connection failed: %v", addr, err)))
						allOk = false
					} else {
						conn.Close()
						fmt.Println(ui.Success(fmt.Sprintf("%s — reachable", addr)))
					}
				}
			}
		}

		// Check data dir
		fmt.Println()
		fmt.Println(ui.Section("Storage"))

		dataDir := config.DataDir()
		if info, err := os.Stat(dataDir); err != nil {
			fmt.Println(ui.Warn(fmt.Sprintf("Data directory missing: %s", dataDir)))
			fmt.Println(ui.Info("  Will be created on first run"))
		} else if !info.IsDir() {
			fmt.Println(ui.Error(fmt.Sprintf("%s exists but is not a directory", dataDir)))
			allOk = false
		} else {
			fmt.Println(ui.Success(fmt.Sprintf("Data directory: %s", dataDir)))
		}

		// Summary
		fmt.Println()
		if allOk {
			fmt.Println(ui.Success("All checks passed — Keeper is ready"))
		} else {
			fmt.Println(ui.Error("Some checks failed — see above for details"))
		}

		return nil
	},
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}
