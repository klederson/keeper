package backend

import "fmt"

func New(backendType string) (BackupBackend, error) {
	switch backendType {
	case "rsync":
		return NewRsync(), nil
	default:
		return nil, fmt.Errorf("unknown backend type: %q", backendType)
	}
}
