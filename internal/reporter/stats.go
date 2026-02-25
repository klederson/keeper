package reporter

import (
	"time"

	"github.com/klederson/keeper/internal/config"
)

type Stats struct {
	TotalRuns        int
	SuccessCount     int
	FailCount        int
	SuccessRate      float64
	TotalBytes       int64
	AvgDuration      time.Duration
	LastRun          *config.RunRecord
}

func CalculateStats(records []config.RunRecord, since time.Time) Stats {
	var stats Stats
	var totalDuration time.Duration

	for i := range records {
		r := &records[i]
		if r.DryRun {
			continue
		}
		if !since.IsZero() && r.StartedAt.Before(since) {
			continue
		}

		stats.TotalRuns++
		if r.Success {
			stats.SuccessCount++
		} else {
			stats.FailCount++
		}

		stats.TotalBytes += r.BytesTransferred
		totalDuration += r.CompletedAt.Sub(r.StartedAt)

		if stats.LastRun == nil || r.CompletedAt.After(stats.LastRun.CompletedAt) {
			stats.LastRun = r
		}
	}

	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalRuns) * 100
		stats.AvgDuration = totalDuration / time.Duration(stats.TotalRuns)
	}

	return stats
}

func CalculateJobStats(store *Store, jobName string) Stats {
	records := store.GetJobRecords(jobName, 0)
	return CalculateStats(records, time.Time{})
}
