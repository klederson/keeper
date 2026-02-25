package backend

import (
	"bufio"
	"bytes"
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
		return fmt.Errorf("rsync not found in PATH — install it with your package manager")
	}
	if job.Destination.Host == "" {
		return fmt.Errorf("destination host is required")
	}
	if job.Destination.Path == "" {
		return fmt.Errorf("destination path is required")
	}
	for _, src := range job.Sources {
		if src.Path == "" {
			return fmt.Errorf("source path cannot be empty")
		}
	}
	return nil
}

func (r *RsyncBackend) Run(ctx context.Context, job *config.Job, dryRun bool, onProgress func(ProgressEvent)) (*Result, error) {
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
		)

		cmd := exec.CommandContext(ctx, "rsync", fullArgs...)

		// Separate stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("pipe error: %v", err))
			continue
		}

		var stderrBuf bytes.Buffer
		cmd.Stderr = &stderrBuf

		if err := cmd.Start(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to start rsync: %v", err))
			continue
		}

		// Read stdout line by line for progress + stats
		filesCount := 0
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			slog.Debug("rsync", "out", line)

			// Parse stats from the summary block
			r.parseStatsLine(line, result)

			// Track file transfers for progress
			if isFileLine(line) {
				filesCount++
				if onProgress != nil {
					onProgress(ProgressEvent{
						CurrentFile: line,
						FilesCount:  filesCount,
						Phase:       "transferring",
					})
				}
			}
		}

		exitErr := cmd.Wait()
		stderrOutput := strings.TrimSpace(stderrBuf.String())

		if exitErr != nil {
			exitCode := cmdExitCode(exitErr)
			explanation := rsyncExitCodeMessage(exitCode)

			// Collect the most useful error details
			errParts := []string{fmt.Sprintf("rsync exited with code %d: %s", exitCode, explanation)}

			// Add stderr lines (often contains the real error)
			if stderrOutput != "" {
				for _, line := range strings.Split(stderrOutput, "\n") {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "rsync error:") {
						errParts = append(errParts, line)
					}
				}
			}

			result.Errors = append(result.Errors, errParts...)
		}

		if onProgress != nil {
			onProgress(ProgressEvent{Phase: "done"})
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

// Stats parsing — handles rsync --stats output
var (
	// "Number of files: 1,234 (reg: 1,100, dir: 134)"
	filesPattern = regexp.MustCompile(`Number of files: ([\d,]+)`)
	// "Number of regular files transferred: 56"
	xferPattern = regexp.MustCompile(`Number of (?:regular )?files transferred: ([\d,]+)`)
	// "Total file size: 1.23G bytes" or "Total file size: 1,234,567 bytes"
	totalSizePattern = regexp.MustCompile(`Total file size: ([\d,\.]+(?:\.\d+)?[KMG]?) bytes`)
	// "Total transferred file size: 456,789 bytes"
	xferSizePattern = regexp.MustCompile(`Total transferred file size: ([\d,\.]+(?:\.\d+)?[KMG]?) bytes`)
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

// isFileLine returns true if the line looks like a file being transferred
// (not a stats line, not blank, not a header)
func isFileLine(line string) bool {
	if line == "" {
		return false
	}
	// rsync stats/summary lines
	if strings.HasPrefix(line, "Number of") ||
		strings.HasPrefix(line, "Total file") ||
		strings.HasPrefix(line, "Total transferred") ||
		strings.HasPrefix(line, "Literal data") ||
		strings.HasPrefix(line, "Matched data") ||
		strings.HasPrefix(line, "File list") ||
		strings.HasPrefix(line, "sent ") ||
		strings.HasPrefix(line, "total size") ||
		strings.HasPrefix(line, "sending incremental") ||
		strings.HasPrefix(line, "building file list") ||
		strings.HasPrefix(line, "created directory") ||
		strings.HasPrefix(line, "Speedup is") ||
		strings.HasPrefix(line, "rsync error") {
		return false
	}
	return true
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

func cmdExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// rsyncExitCodeMessage translates rsync exit codes to human-readable messages.
func rsyncExitCodeMessage(code int) string {
	messages := map[int]string{
		1:  "syntax or usage error",
		2:  "protocol incompatibility",
		3:  "errors selecting input/output files/dirs",
		4:  "requested action not supported",
		5:  "error starting client-server protocol",
		6:  "daemon unable to append to log file",
		10: "error in socket I/O",
		11: "error in file I/O",
		12: "error in rsync protocol data stream",
		13: "errors with program diagnostics",
		14: "error in IPC code",
		20: "received SIGUSR1 or SIGINT",
		21: "some error returned by waitpid()",
		22: "error allocating core memory buffers",
		23: "partial transfer — some files/attrs were not transferred (check permissions and paths)",
		24: "partial transfer — vanished source files (files changed during transfer)",
		25: "the --max-delete limit stopped deletions",
		30: "timeout in data send/receive",
		35: "timeout waiting for daemon connection",
		255: "SSH connection failed — check host, port, and SSH key",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}
