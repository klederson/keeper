package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/klederson/keeper/internal/config"
)

func prompt(label, defaultVal, placeholder string) string {
	promptStr := AccentStyle.Render("▸ ")
	title := SubtitleStyle.Render(label)

	hint := ""
	if defaultVal != "" {
		hint = MutedStyle.Render(fmt.Sprintf(" [%s]", defaultVal))
	} else if placeholder != "" {
		hint = MutedStyle.Render(fmt.Sprintf(" (%s)", placeholder))
	}

	fmt.Printf("%s%s%s ", promptStr, title, hint)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

func promptSelect(label string, options []string, defaultIdx int) int {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render(label))

	for i, opt := range options {
		marker := MutedStyle.Render("  ○")
		if i == defaultIdx {
			marker = AccentStyle.Render("  ●")
		}
		fmt.Printf("%s %s\n", marker, TextStyle.Render(opt))
	}

	hint := MutedStyle.Render(fmt.Sprintf("  Enter choice [1-%d]", len(options)))
	if defaultIdx >= 0 {
		hint += MutedStyle.Render(fmt.Sprintf(" (default: %d)", defaultIdx+1))
	}
	fmt.Printf("%s: ", hint)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultIdx
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
		return idx - 1
	}

	return defaultIdx
}

func promptConfirm(label string, defaultYes bool) bool {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	fmt.Printf("%s %s %s: ",
		AccentStyle.Render("▸"),
		SubtitleStyle.Render(label),
		MutedStyle.Render(hint),
	)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}
	return input == "y" || input == "yes"
}

func RunInitForm() (*config.Config, error) {
	fmt.Println(Section("Initialize Keeper"))
	fmt.Println(MutedStyle.Render("  Configure your default backup destination\n"))

	destHost := prompt("Destination host", "", "backup.server.com")
	destUser := prompt("Destination user", "", "backupuser")
	destPath := prompt("Destination base path", "", "/backups")
	sshKey := prompt("SSH key path", "~/.ssh/id_rsa", "")

	logLevels := []string{"info (recommended)", "debug (verbose)", "warn (quiet)", "error (minimal)"}
	logLevelIdx := promptSelect("Log level", logLevels, 0)
	logLevelValues := []string{"info", "debug", "warn", "error"}
	logLevel := logLevelValues[logLevelIdx]

	cfg := config.DefaultConfig()
	cfg.LogLevel = logLevel

	fmt.Println()
	fmt.Println(Success("Configuration initialized"))
	fmt.Println(Label("  Host", destHost))
	fmt.Println(Label("  User", destUser))
	fmt.Println(Label("  Path", destPath))
	fmt.Println(Label("  SSH Key", sshKey))
	fmt.Println()
	fmt.Println(Info("Run 'keeper add' to create your first backup job"))

	return &cfg, nil
}

func RunAddJobForm(existing *config.Job) (*config.Job, error) {
	job := config.Job{
		Destination: config.Destination{
			Type: "rsync",
			Port: 22,
		},
		Compress: true,
	}

	if existing != nil {
		job = *existing
	}

	title := "Add Backup Job"
	if existing != nil {
		title = "Edit Backup Job: " + existing.Name
	}

	fmt.Println(Section(title))
	fmt.Println(MutedStyle.Render("  Configure your backup job\n"))

	// Basic info
	name := prompt("Job name", job.Name, "my-project")

	sourcePath := ""
	if len(job.Sources) > 0 {
		sourcePath = job.Sources[0].Path
	}
	sourcePath = prompt("Source path", sourcePath, "/home/user/Projects")

	existingExcludes := ""
	if len(job.Sources) > 0 {
		existingExcludes = strings.Join(job.Sources[0].Exclude, ", ")
	}
	excludes := prompt("Exclude patterns (comma-separated)", existingExcludes, "node_modules/, .git/, vendor/")

	fmt.Println()
	fmt.Println(SubtitleStyle.Render("  Destination"))

	destHost := prompt("Host", job.Destination.Host, "backup.server.com")
	destUser := prompt("User", job.Destination.User, "backupuser")
	destPath := prompt("Path", job.Destination.Path, "/backups/my-project")
	sshKey := prompt("SSH key path", job.Destination.SSHKey, "~/.ssh/id_rsa")

	fmt.Println()
	fmt.Println(SubtitleStyle.Render("  Schedule"))

	scheduleOptions := []string{
		"Every day at 2:00 AM (0 2 * * *)",
		"Every 6 hours (0 */6 * * *)",
		"Every hour (0 * * * *)",
		"Weekly - Sunday 3:00 AM (0 3 * * 0)",
		"Custom cron expression",
	}
	scheduleValues := []string{
		"0 2 * * *",
		"0 */6 * * *",
		"0 * * * *",
		"0 3 * * 0",
		"custom",
	}

	schedIdx := promptSelect("Backup schedule", scheduleOptions, 0)
	schedule := scheduleValues[schedIdx]

	if schedule == "custom" {
		schedule = prompt("Cron expression", job.Schedule, "0 2 * * *")
	}

	compress := promptConfirm("Enable compression? (rsync -z)", job.Compress)
	bandwidth := prompt("Bandwidth limit (0 = unlimited)", job.Bandwidth, "0")

	// Parse excludes
	var excludeList []string
	if excludes != "" {
		for _, e := range strings.Split(excludes, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				excludeList = append(excludeList, e)
			}
		}
	}

	if sshKey == "" {
		sshKey = "~/.ssh/id_rsa"
	}
	if bandwidth == "" {
		bandwidth = "0"
	}

	result := &config.Job{
		Name: name,
		Sources: []config.Source{
			{
				Path:    sourcePath,
				Exclude: excludeList,
			},
		},
		Destination: config.Destination{
			Type:   "rsync",
			Host:   destHost,
			User:   destUser,
			Path:   destPath,
			SSHKey: sshKey,
			Port:   22,
		},
		Schedule:  schedule,
		Bandwidth: bandwidth,
		Compress:  compress,
	}

	return result, nil
}

func ConfirmRemove(jobName string) (bool, error) {
	return promptConfirm(fmt.Sprintf("Remove job %q? This cannot be undone", jobName), false), nil
}
