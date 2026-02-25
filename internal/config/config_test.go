package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LogLevel != "info" {
		t.Errorf("expected log_level 'info', got %q", cfg.LogLevel)
	}
	if len(cfg.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(cfg.Jobs))
	}
}

func TestConfigSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		LogDir:   tmpDir,
		LogLevel: "debug",
		Jobs: []Job{
			{
				Name: "test-job",
				Sources: []Source{
					{
						Path:    "/home/user/Projects",
						Exclude: []string{"node_modules/", ".git/"},
					},
				},
				Destination: Destination{
					Type:   "rsync",
					Host:   "backup.server.com",
					User:   "backupuser",
					Path:   "/backups/test",
					SSHKey: "~/.ssh/id_rsa",
					Port:   22,
				},
				Schedule:  "0 2 * * *",
				Bandwidth: "500k",
				Compress:  true,
			},
		},
	}

	// Marshal to YAML and write
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read back and unmarshal
	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got %q", loaded.LogLevel)
	}
	if len(loaded.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(loaded.Jobs))
	}

	job := loaded.Jobs[0]
	if job.Name != "test-job" {
		t.Errorf("expected job name 'test-job', got %q", job.Name)
	}
	if job.Destination.Host != "backup.server.com" {
		t.Errorf("expected host 'backup.server.com', got %q", job.Destination.Host)
	}
	if job.Bandwidth != "500k" {
		t.Errorf("expected bandwidth '500k', got %q", job.Bandwidth)
	}
	if !job.Compress {
		t.Error("expected compress to be true")
	}
	if len(job.Sources[0].Exclude) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(job.Sources[0].Exclude))
	}
}

func TestFindJob(t *testing.T) {
	cfg := Config{
		Jobs: []Job{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
		},
	}

	job, idx := cfg.FindJob("beta")
	if job == nil || idx != 1 {
		t.Errorf("expected to find 'beta' at index 1, got idx=%d", idx)
	}

	job, idx = cfg.FindJob("nonexistent")
	if job != nil || idx != -1 {
		t.Errorf("expected nil for nonexistent job, got idx=%d", idx)
	}
}

func TestAddRemoveJob(t *testing.T) {
	cfg := Config{}

	err := cfg.AddJob(Job{Name: "test"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}

	// Duplicate should fail
	err = cfg.AddJob(Job{Name: "test"})
	if err == nil {
		t.Fatal("expected error for duplicate job")
	}

	err = cfg.RemoveJob("test")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(cfg.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(cfg.Jobs))
	}

	// Remove nonexistent should fail
	err = cfg.RemoveJob("test")
	if err == nil {
		t.Fatal("expected error for removing nonexistent job")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Jobs: []Job{{
					Name:    "test",
					Sources: []Source{{Path: "/tmp"}},
					Destination: Destination{
						Type: "rsync",
						Host: "example.com",
						Path: "/backups",
					},
				}},
			},
			wantErr: false,
		},
		{
			name: "empty job name",
			cfg: Config{
				Jobs: []Job{{
					Sources: []Source{{Path: "/tmp"}},
					Destination: Destination{
						Type: "rsync",
						Host: "example.com",
						Path: "/backups",
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "no sources",
			cfg: Config{
				Jobs: []Job{{
					Name: "test",
					Destination: Destination{
						Type: "rsync",
						Host: "example.com",
						Path: "/backups",
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "missing dest host",
			cfg: Config{
				Jobs: []Job{{
					Name:    "test",
					Sources: []Source{{Path: "/tmp"}},
					Destination: Destination{
						Type: "rsync",
						Path: "/backups",
					},
				}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/.ssh/id_rsa", filepath.Join(home, ".ssh/id_rsa")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := ExpandPath(tt.input)
		if got != tt.want {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
