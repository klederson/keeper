package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigDir  = ".config/keeper"
	DefaultConfigFile = "config.yaml"
	DefaultDataDir    = ".local/share/keeper"
	DefaultLogDir     = ".local/share/keeper/logs"
)

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultConfigDir)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), DefaultConfigFile)
}

func DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultDataDir)
}

func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

func DefaultConfig() Config {
	return Config{
		LogDir:   "~/" + DefaultLogDir,
		LogLevel: "info",
		Jobs:     []Job{},
	}
}

func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found at %s — run 'keeper init' first", path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.LogDir == "" {
		cfg.LogDir = "~/" + DefaultLogDir
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := ConfigPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func (c *Config) FindJob(name string) (*Job, int) {
	for i := range c.Jobs {
		if c.Jobs[i].Name == name {
			return &c.Jobs[i], i
		}
	}
	return nil, -1
}

func (c *Config) AddJob(job Job) error {
	if existing, _ := c.FindJob(job.Name); existing != nil {
		return fmt.Errorf("job %q already exists", job.Name)
	}
	c.Jobs = append(c.Jobs, job)
	return nil
}

func (c *Config) RemoveJob(name string) error {
	_, idx := c.FindJob(name)
	if idx < 0 {
		return fmt.Errorf("job %q not found", name)
	}
	c.Jobs = append(c.Jobs[:idx], c.Jobs[idx+1:]...)
	return nil
}

func (c *Config) Validate() error {
	for _, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job name cannot be empty")
		}
		if len(job.Sources) == 0 {
			return fmt.Errorf("job %q: at least one source required", job.Name)
		}
		for _, src := range job.Sources {
			if src.Path == "" {
				return fmt.Errorf("job %q: source path cannot be empty", job.Name)
			}
		}
		if job.Destination.Type == "" {
			return fmt.Errorf("job %q: destination type required", job.Name)
		}
		if job.Destination.Host == "" {
			return fmt.Errorf("job %q: destination host required", job.Name)
		}
		if job.Destination.Path == "" {
			return fmt.Errorf("job %q: destination path required", job.Name)
		}
	}
	return nil
}

func EnsureDataDir() error {
	return os.MkdirAll(DataDir(), 0755)
}
