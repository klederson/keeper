package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var Theme = struct {
	BG        color.Color
	Panel     color.Color
	Primary   color.Color
	Accent    color.Color
	Error     color.Color
	Warning   color.Color
	Text      color.Color
	Muted     color.Color
	Secondary color.Color
}{
	BG:        lipgloss.Color("#1b1e28"),
	Panel:     lipgloss.Color("#303340"),
	Primary:   lipgloss.Color("#ADD7FF"),
	Accent:    lipgloss.Color("#5DE4c7"),
	Error:     lipgloss.Color("#d0679d"),
	Warning:   lipgloss.Color("#fffac2"),
	Text:      lipgloss.Color("#e4f0fb"),
	Muted:     lipgloss.Color("#767c9d"),
	Secondary: lipgloss.Color("#a6accd"),
}

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Theme.Secondary)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Theme.Accent)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Theme.Error)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Theme.Warning)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Theme.Muted)

	TextStyle = lipgloss.NewStyle().
			Foreground(Theme.Text)

	SuccessBadge = lipgloss.NewStyle().
			Foreground(Theme.Accent).
			Bold(true)

	ErrorBadge = lipgloss.NewStyle().
			Foreground(Theme.Error).
			Bold(true)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Panel).
			Padding(0, 1)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Theme.Primary).
				Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Theme.Panel)
)
