package backend

import (
	"testing"

	"github.com/klederson/keeper/internal/config"
)

func TestBuildArgs(t *testing.T) {
	r := NewRsync()

	tests := []struct {
		name     string
		job      *config.Job
		source   *config.Source
		dryRun   bool
		wantArgs []string
	}{
		{
			name: "basic with compression",
			job: &config.Job{
				Compress: true,
				Destination: config.Destination{
					Port: 22,
				},
			},
			source: &config.Source{
				Path: "/home/user/Projects",
			},
			dryRun: false,
			wantArgs: []string{"-av", "--stats", "--human-readable", "-z", "-e", "ssh"},
		},
		{
			name: "dry run",
			job: &config.Job{
				Destination: config.Destination{
					Port: 22,
				},
			},
			source: &config.Source{},
			dryRun: true,
			wantArgs: []string{"-av", "--stats", "--human-readable", "--dry-run", "-e", "ssh"},
		},
		{
			name: "with delete and bandwidth",
			job: &config.Job{
				Delete:    true,
				Bandwidth: "500k",
				Destination: config.Destination{
					Port: 22,
				},
			},
			source: &config.Source{},
			dryRun: false,
			wantArgs: []string{"-av", "--stats", "--human-readable", "--delete", "--bwlimit=500k", "-e", "ssh"},
		},
		{
			name: "with SSH key and port",
			job: &config.Job{
				Destination: config.Destination{
					SSHKey: "/home/user/.ssh/backup_key",
					Port:   2222,
				},
			},
			source: &config.Source{},
			dryRun: false,
			wantArgs: []string{"-av", "--stats", "--human-readable", "-e", "ssh -i /home/user/.ssh/backup_key -p 2222"},
		},
		{
			name: "with excludes",
			job: &config.Job{
				Destination: config.Destination{
					Port: 22,
				},
			},
			source: &config.Source{
				Exclude: []string{"node_modules/", ".git/"},
			},
			dryRun: false,
			wantArgs: []string{"-av", "--stats", "--human-readable", "-e", "ssh", "--exclude=node_modules/", "--exclude=.git/"},
		},
		{
			name: "with includes",
			job: &config.Job{
				Destination: config.Destination{
					Port: 22,
				},
			},
			source: &config.Source{
				Include: []string{"**/*.go", "**/*.py"},
			},
			dryRun: false,
			wantArgs: []string{"-av", "--stats", "--human-readable", "-e", "ssh", "--include=**/*.go", "--include=**/*.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.buildArgs(tt.job, tt.source, tt.dryRun)
			if len(got) != len(tt.wantArgs) {
				t.Errorf("len mismatch: got %d, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantArgs), got, tt.wantArgs)
				return
			}
			for i := range got {
				if got[i] != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q\nfull got:  %v\nfull want: %v", i, got[i], tt.wantArgs[i], got, tt.wantArgs)
				}
			}
		})
	}
}

func TestBuildDest(t *testing.T) {
	r := NewRsync()

	tests := []struct {
		job  *config.Job
		want string
	}{
		{
			job: &config.Job{Destination: config.Destination{
				User: "backup",
				Host: "server.com",
				Path: "/backups/test",
			}},
			want: "backup@server.com:/backups/test",
		},
		{
			job: &config.Job{Destination: config.Destination{
				Host: "server.com",
				Path: "/backups/test",
			}},
			want: "server.com:/backups/test",
		},
	}

	for _, tt := range tests {
		got := r.buildDest(tt.job)
		if got != tt.want {
			t.Errorf("buildDest() = %q, want %q", got, tt.want)
		}
	}
}

func TestParseStatsLine(t *testing.T) {
	r := NewRsync()

	tests := []struct {
		line         string
		wantFiles    int
		wantXfer     int
		wantTotal    int64
		wantXferSize int64
	}{
		{
			line:      "Number of files: 1,234",
			wantFiles: 1234,
		},
		{
			line:     "Number of regular files transferred: 56",
			wantXfer: 56,
		},
		{
			line:      "Total file size: 1,234,567",
			wantTotal: 1234567,
		},
		{
			line:         "Total transferred file size: 456,789 ",
			wantXferSize: 456789,
		},
	}

	for _, tt := range tests {
		result := &Result{}
		r.parseStatsLine(tt.line, result)

		if tt.wantFiles > 0 && result.FilesTotal != tt.wantFiles {
			t.Errorf("line %q: FilesTotal = %d, want %d", tt.line, result.FilesTotal, tt.wantFiles)
		}
		if tt.wantXfer > 0 && result.FilesTransferred != tt.wantXfer {
			t.Errorf("line %q: FilesTransferred = %d, want %d", tt.line, result.FilesTransferred, tt.wantXfer)
		}
		if tt.wantTotal > 0 && result.BytesTotal != tt.wantTotal {
			t.Errorf("line %q: BytesTotal = %d, want %d", tt.line, result.BytesTotal, tt.wantTotal)
		}
		if tt.wantXferSize > 0 && result.BytesTransferred != tt.wantXferSize {
			t.Errorf("line %q: BytesTransferred = %d, want %d", tt.line, result.BytesTransferred, tt.wantXferSize)
		}
	}
}

func TestParseIntComma(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1234", 1234},
		{"1,234", 1234},
		{"1,234,567", 1234567},
		{"0", 0},
	}

	for _, tt := range tests {
		got := parseIntComma(tt.input)
		if got != tt.want {
			t.Errorf("parseIntComma(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseSizeBytes(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1024", 1024},
		{"1K", 1024},
		{"1M", 1048576},
		{"1G", 1073741824},
		{"1,234", 1234},
	}

	for _, tt := range tests {
		got := parseSizeBytes(tt.input)
		if got != tt.want {
			t.Errorf("parseSizeBytes(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
