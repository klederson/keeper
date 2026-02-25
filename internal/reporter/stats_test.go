package reporter

import (
	"math"
	"testing"
	"time"

	"github.com/klederson/keeper/internal/config"
)

func TestCalculateStats(t *testing.T) {
	now := time.Now()

	records := []config.RunRecord{
		{
			JobName:          "test",
			StartedAt:        now.Add(-3 * time.Hour),
			CompletedAt:      now.Add(-3*time.Hour + 2*time.Minute),
			Success:          true,
			BytesTransferred: 1000,
		},
		{
			JobName:          "test",
			StartedAt:        now.Add(-2 * time.Hour),
			CompletedAt:      now.Add(-2*time.Hour + 4*time.Minute),
			Success:          true,
			BytesTransferred: 2000,
		},
		{
			JobName:          "test",
			StartedAt:        now.Add(-1 * time.Hour),
			CompletedAt:      now.Add(-1*time.Hour + 6*time.Minute),
			Success:          false,
			BytesTransferred: 500,
			Errors:           []string{"connection refused"},
		},
	}

	stats := CalculateStats(records, time.Time{})

	if stats.TotalRuns != 3 {
		t.Errorf("TotalRuns = %d, want 3", stats.TotalRuns)
	}
	if stats.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", stats.SuccessCount)
	}
	if stats.FailCount != 1 {
		t.Errorf("FailCount = %d, want 1", stats.FailCount)
	}

	expectedRate := 66.66
	if math.Abs(stats.SuccessRate-expectedRate) > 1.0 {
		t.Errorf("SuccessRate = %.2f, want ~%.2f", stats.SuccessRate, expectedRate)
	}

	if stats.TotalBytes != 3500 {
		t.Errorf("TotalBytes = %d, want 3500", stats.TotalBytes)
	}

	expectedAvg := 4 * time.Minute
	if math.Abs(float64(stats.AvgDuration-expectedAvg)) > float64(time.Second) {
		t.Errorf("AvgDuration = %v, want ~%v", stats.AvgDuration, expectedAvg)
	}
}

func TestCalculateStatsEmpty(t *testing.T) {
	stats := CalculateStats(nil, time.Time{})

	if stats.TotalRuns != 0 {
		t.Errorf("TotalRuns = %d, want 0", stats.TotalRuns)
	}
	if stats.SuccessRate != 0 {
		t.Errorf("SuccessRate = %.2f, want 0", stats.SuccessRate)
	}
}

func TestCalculateStatsSinceFilter(t *testing.T) {
	now := time.Now()

	records := []config.RunRecord{
		{
			JobName:     "test",
			StartedAt:   now.Add(-48 * time.Hour),
			CompletedAt: now.Add(-48*time.Hour + time.Minute),
			Success:     true,
		},
		{
			JobName:     "test",
			StartedAt:   now.Add(-1 * time.Hour),
			CompletedAt: now.Add(-1*time.Hour + time.Minute),
			Success:     true,
		},
	}

	stats := CalculateStats(records, now.Add(-24*time.Hour))

	if stats.TotalRuns != 1 {
		t.Errorf("TotalRuns = %d, want 1 (filtered by since)", stats.TotalRuns)
	}
}

func TestCalculateStatsDryRunExcluded(t *testing.T) {
	now := time.Now()

	records := []config.RunRecord{
		{
			JobName:     "test",
			StartedAt:   now.Add(-1 * time.Hour),
			CompletedAt: now.Add(-1*time.Hour + time.Minute),
			Success:     true,
			DryRun:      true,
		},
		{
			JobName:     "test",
			StartedAt:   now.Add(-30 * time.Minute),
			CompletedAt: now.Add(-30*time.Minute + time.Minute),
			Success:     true,
		},
	}

	stats := CalculateStats(records, time.Time{})

	if stats.TotalRuns != 1 {
		t.Errorf("TotalRuns = %d, want 1 (dry runs excluded)", stats.TotalRuns)
	}
}
