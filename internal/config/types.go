package config

import "time"

type Config struct {
	LogDir   string `yaml:"log_dir" mapstructure:"log_dir"`
	LogLevel string `yaml:"log_level" mapstructure:"log_level"`
	Jobs     []Job  `yaml:"jobs" mapstructure:"jobs"`
}

type Job struct {
	Name        string      `yaml:"name" mapstructure:"name"`
	Sources     []Source    `yaml:"sources" mapstructure:"sources"`
	Destination Destination `yaml:"destination" mapstructure:"destination"`
	Schedule    string      `yaml:"schedule" mapstructure:"schedule"`
	Bandwidth   string      `yaml:"bandwidth" mapstructure:"bandwidth"`
	Delete      bool        `yaml:"delete" mapstructure:"delete"`
	Compress    bool        `yaml:"compress" mapstructure:"compress"`
}

type Source struct {
	Path    string   `yaml:"path" mapstructure:"path"`
	Include []string `yaml:"include,omitempty" mapstructure:"include"`
	Exclude []string `yaml:"exclude,omitempty" mapstructure:"exclude"`
}

type Destination struct {
	Type   string `yaml:"type" mapstructure:"type"`
	Host   string `yaml:"host" mapstructure:"host"`
	User   string `yaml:"user" mapstructure:"user"`
	Path   string `yaml:"path" mapstructure:"path"`
	SSHKey string `yaml:"ssh_key" mapstructure:"ssh_key"`
	Port   int    `yaml:"port" mapstructure:"port"`
}

type RunRecord struct {
	JobName          string    `json:"job_name"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	Success          bool      `json:"success"`
	FilesTotal       int       `json:"files_total"`
	FilesTransferred int       `json:"files_transferred"`
	BytesTotal       int64     `json:"bytes_total"`
	BytesTransferred int64     `json:"bytes_transferred"`
	Errors           []string  `json:"errors,omitempty"`
	DryRun           bool      `json:"dry_run"`
}
