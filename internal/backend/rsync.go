package backend

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/klederson/keeper/internal/config"
)

type RsyncBackend struct{}

func NewRsync() *RsyncBackend {
	return &RsyncBackend{}
}

func (r *RsyncBackend) Name() string {
	return "rsync"
}

func (r *RsyncBackend) Validate(job *config.Job) error {
	if _, err := exec.LookPath("rsync"); err != nil {
		return fmt.Errorf("rsync not found in PATH")
	}
	if job.Destination.Host == "" {
		return fmt.Errorf("destination host is required")
	}
	if job.Destination.Path == "" {
		return fmt.Errorf("destination path is required")
	}
	return nil
}

func (r *RsyncBackend) Run(ctx context.Context, job *config.Job, dryRun bool) (*Result, error) {
	result := &Result{
		StartedAt: time.Now(),
	}

	for _, source := range job.Sources {
		args := r.buildArgs(job, &source, dryRun)
		dest := r.buildDest(job)
		srcPath := config.ExpandPath(source.Path)
		if !strings.HasSuffix(srcPath, "/") {
			srcPath += "/"
		}

		fullArgs := append(args, srcPath, dest)

		slog.Info("executing rsync",
			"job", job.Name,
			"source", srcPath,
			"dest", dest,
			"dry_run", dryRun,
			"args", strings.Join(fullArgs, " "),
		)

		cmd := exec.CommandContext(ctx, "rsync", fullArgs...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("pipe error: %v", err))
			continue
		}

		cmd.Stderr = cmd.Stdout

		if err := cmd.Start(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("start error: %v", err))
			continue
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			slog.Debug("rsync output", "line", line)
			r.parseStatsLine(line, result)
		}

		if err := cmd.Wait(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("rsync error: %v", err))
		}
	}

	result.CompletedAt = time.Now()
	result.Success = len(result.Errors) == 0

	return result, nil
}

func (r *RsyncBackend) buildArgs(job *config.Job, source *config.Source, dryRun bool) []string {
	args := []string{"-av", "--stats", "--human-readable"}

	if job.Compress {
		args = append(args, "-z")
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	if job.Delete {
		args = append(args, "--delete")
	}

	if job.Bandwidth != "" && job.Bandwidth != "0" {
		args = append(args, "--bwlimit="+job.Bandwidth)
	}

	// SSH options
	sshCmd := "ssh"
	if job.Destination.SSHKey != "" {
		key := config.ExpandPath(job.Destination.SSHKey)
		sshCmd += " -i " + key
	}
	if job.Destination.Port != 0 && job.Destination.Port != 22 {
		sshCmd += fmt.Sprintf(" -p %d", job.Destination.Port)
	}
	args = append(args, "-e", sshCmd)

	// Include patterns
	for _, inc := range source.Include {
		args = append(args, "--include="+inc)
	}

	// Exclude patterns
	for _, exc := range source.Exclude {
		args = append(args, "--exclude="+exc)
	}

	return args
}

func (r *RsyncBackend) buildDest(job *config.Job) string {
	user := job.Destination.User
	host := job.Destination.Host
	path := job.Destination.Path

	if user != "" {
		return fmt.Sprintf("%s@%s:%s", user, host, path)
	}
	return fmt.Sprintf("%s:%s", host, path)
}

var (
	filesPattern    = regexp.MustCompile(`Number of files: (\d[\d,]*)`)
	xferPattern     = regexp.MustCompile(`Number of (?:regular )?files transferred: (\d[\d,]*)`)
	totalSizePattern = regexp.MustCompile(`Total file size: ([\d,\.]+[KMG]?)`)
	xferSizePattern = regexp.MustCompile(`Total transferred file size: ([\d,\.]+)\s`)
)

func (r *RsyncBackend) parseStatsLine(line string, result *Result) {
	if m := filesPattern.FindStringSubmatch(line); len(m) > 1 {
		result.FilesTotal = parseIntComma(m[1])
	}
	if m := xferPattern.FindStringSubmatch(line); len(m) > 1 {
		result.FilesTransferred = parseIntComma(m[1])
	}
	if m := totalSizePattern.FindStringSubmatch(line); len(m) > 1 {
		result.BytesTotal = parseSizeBytes(m[1])
	}
	if m := xferSizePattern.FindStringSubmatch(line); len(m) > 1 {
		result.BytesTransferred = parseSizeBytes(m[1])
	}
}

func parseIntComma(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.Atoi(s)
	return n
}

func parseSizeBytes(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	multiplier := int64(1)

	if strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	f, _ := strconv.ParseFloat(s, 64)
	return int64(f) * multiplier
}
