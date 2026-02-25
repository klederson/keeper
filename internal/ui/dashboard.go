package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/klederson/keeper/internal/config"
	"github.com/klederson/keeper/internal/reporter"
)

type DashboardModel struct {
	cfg       *config.Config
	store     *reporter.Store
	stats     reporter.Stats
	records   []config.RunRecord
	cursor    int
	width     int
	height    int
	quitting  bool
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewDashboard(cfg *config.Config, store *reporter.Store) DashboardModel {
	allRecords := store.LoadAll()
	stats := reporter.CalculateStats(allRecords, time.Now().AddDate(0, 0, -30))
	recent := store.GetRecentRecords(10)

	return DashboardModel{
		cfg:     cfg,
		store:   store,
		stats:   stats,
		records: recent,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tickCmd()
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.cfg.Jobs)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.refreshData()
		return m, tickCmd()
	}

	return m, nil
}

func (m *DashboardModel) refreshData() {
	allRecords := m.store.LoadAll()
	m.stats = reporter.CalculateStats(allRecords, time.Now().AddDate(0, 0, -30))
	m.records = m.store.GetRecentRecords(10)
}

func (m DashboardModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	width := m.width
	if width == 0 {
		width = 80
	}

	var b strings.Builder

	// Title
	title := TitleStyle.Render(" Keeper Dashboard ")
	titleBar := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(title)
	b.WriteString(titleBar + "\n\n")

	// Jobs table
	b.WriteString(renderJobsTable(m.cfg, m.store, m.cursor))
	b.WriteString("\n")

	// Stats panel
	b.WriteString(renderStatsPanel(m.stats))
	b.WriteString("\n")

	// Recent activity
	b.WriteString(renderRecentActivity(m.records))
	b.WriteString("\n")

	// Footer
	footer := MutedStyle.Render("  [↑/↓] navigate  [q] quit")
	b.WriteString(footer + "\n")

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func renderJobsTable(cfg *config.Config, store *reporter.Store, cursor int) string {
	var b strings.Builder
	b.WriteString(headerLine("Jobs") + "\n")

	if len(cfg.Jobs) == 0 {
		b.WriteString(MutedStyle.Render("  No jobs configured\n"))
		return b.String()
	}

	// Header
	hdr := fmt.Sprintf("  %-16s %-12s %-12s %-14s",
		SubtitleStyle.Render("Name"),
		SubtitleStyle.Render("Schedule"),
		SubtitleStyle.Render("Last Run"),
		SubtitleStyle.Render("Status"),
	)
	b.WriteString(hdr + "\n")
	b.WriteString(MutedStyle.Render("  "+strings.Repeat("─", 56)) + "\n")

	for i, job := range cfg.Jobs {
		lastRun := MutedStyle.Render("never")
		status := MutedStyle.Render("—")

		records := store.GetJobRecords(job.Name, 1)
		if len(records) > 0 {
			r := records[0]
			lastRun = formatTimeAgo(r.CompletedAt)
			if r.Success {
				status = AccentStyle.Render("✓ success")
			} else {
				status = ErrorStyle.Render("✗ failed")
			}
		}

		prefix := "  "
		nameStyle := TextStyle
		if i == cursor {
			prefix = AccentStyle.Render("▸ ")
			nameStyle = lipgloss.NewStyle().Foreground(Theme.Accent).Bold(true)
		}

		line := fmt.Sprintf("%s%-16s %-12s %-12s %-14s",
			prefix,
			nameStyle.Render(job.Name),
			MutedStyle.Render(job.Schedule),
			lastRun,
			status,
		)
		b.WriteString(line + "\n")
	}

	return b.String()
}

func renderStatsPanel(stats reporter.Stats) string {
	var b strings.Builder
	b.WriteString(headerLine("Stats (30 days)") + "\n")

	left := fmt.Sprintf("  Success rate: %s  │  Total: %s transferred",
		AccentStyle.Render(fmt.Sprintf("%.1f%%", stats.SuccessRate)),
		TextStyle.Render(formatBytesCompact(stats.TotalBytes)),
	)
	right := fmt.Sprintf("  Avg duration: %s │  Jobs run: %s",
		TextStyle.Render(formatDurationCompact(stats.AvgDuration)),
		TextStyle.Render(fmt.Sprintf("%d", stats.TotalRuns)),
	)

	b.WriteString(left + "\n")
	b.WriteString(right + "\n")

	return b.String()
}

func renderRecentActivity(records []config.RunRecord) string {
	var b strings.Builder
	b.WriteString(headerLine("Recent Activity") + "\n")

	if len(records) == 0 {
		b.WriteString(MutedStyle.Render("  No activity yet\n"))
		return b.String()
	}

	for _, r := range records {
		icon := AccentStyle.Render("✓")
		if !r.Success {
			icon = ErrorStyle.Render("✗")
		}

		timeStr := MutedStyle.Render(fmt.Sprintf("%-8s", formatTimeAgo(r.CompletedAt)))
		name := SubtitleStyle.Render(fmt.Sprintf("[%s]", r.JobName))

		detail := ""
		if r.Success {
			dur := r.CompletedAt.Sub(r.StartedAt).Round(time.Second)
			detail = TextStyle.Render(fmt.Sprintf("%s in %s", formatBytesCompact(r.BytesTransferred), dur))
		} else if len(r.Errors) > 0 {
			msg := r.Errors[0]
			if len(msg) > 40 {
				msg = msg[:40] + "..."
			}
			detail = ErrorStyle.Render(msg)
		}

		b.WriteString(fmt.Sprintf("  %s %s %s %s\n", timeStr, name, icon, detail))
	}

	return b.String()
}

func headerLine(title string) string {
	return lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true).
		Padding(0, 1).
		Render(title)
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}

func formatBytesCompact(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDurationCompact(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %02ds", m, s)
}
