package reporter

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/klederson/keeper/internal/config"
)

type Store struct {
	path string
}

func NewStore() *Store {
	return &Store{
		path: filepath.Join(config.DataDir(), "history.jsonl"),
	}
}

func (s *Store) Append(record config.RunRecord) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		slog.Error("creating history dir", "error", err)
		return
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("opening history file", "error", err)
		return
	}
	defer f.Close()

	data, err := json.Marshal(record)
	if err != nil {
		slog.Error("marshaling record", "error", err)
		return
	}

	f.Write(data)
	f.Write([]byte("\n"))
}

func (s *Store) LoadAll() []config.RunRecord {
	f, err := os.Open(s.path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var records []config.RunRecord
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var r config.RunRecord
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		records = append(records, r)
	}

	return records
}

func (s *Store) GetJobRecords(jobName string, limit int) []config.RunRecord {
	all := s.LoadAll()

	var filtered []config.RunRecord
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].JobName == jobName {
			filtered = append(filtered, all[i])
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}

	return filtered
}

func (s *Store) GetRecentRecords(limit int) []config.RunRecord {
	all := s.LoadAll()

	start := len(all) - limit
	if start < 0 {
		start = 0
	}

	// Return most recent first
	result := make([]config.RunRecord, 0, limit)
	for i := len(all) - 1; i >= start; i-- {
		result = append(result, all[i])
	}

	return result
}
