package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func Logo() string {
	logo := `
 ┃ ╻┏━╸┏━╸┏━┓┏━╸┏━┓
 ┣━┫┣╸ ┣╸ ┣━┛┣╸ ┣┳┛
 ╹ ╹┗━╸┗━╸╹  ┗━╸╹┗╸`
	return TitleStyle.Render(logo)
}

func Banner() string {
	return Logo() + "\n" + MutedStyle.Render("  Backup daemon & CLI tool") + "\n"
}

func Success(msg string) string {
	return AccentStyle.Render("✓") + " " + TextStyle.Render(msg)
}

func Error(msg string) string {
	return ErrorStyle.Render("✗") + " " + ErrorStyle.Render(msg)
}

func Warn(msg string) string {
	return WarningStyle.Render("!") + " " + WarningStyle.Render(msg)
}

func Info(msg string) string {
	return lipgloss.NewStyle().Foreground(Theme.Primary).Render("•") + " " + TextStyle.Render(msg)
}

func Label(label, value string) string {
	return SubtitleStyle.Render(label+":") + " " + TextStyle.Render(value)
}

func Section(title string) string {
	return "\n" + HeaderStyle.Render(title) + "\n"
}

type TableColumn struct {
	Title string
	Width int
}

func Table(columns []TableColumn, rows [][]string) string {
	if len(rows) == 0 {
		return MutedStyle.Render("  No data to display")
	}

	var b strings.Builder

	// Header
	headerCells := make([]string, len(columns))
	for i, col := range columns {
		style := lipgloss.NewStyle().
			Width(col.Width).
			Foreground(Theme.Primary).
			Bold(true).
			Padding(0, 1)
		headerCells[i] = style.Render(col.Title)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Separator
	sepCells := make([]string, len(columns))
	for i, col := range columns {
		style := lipgloss.NewStyle().
			Width(col.Width).
			Foreground(Theme.Panel).
			Padding(0, 1)
		sepCells[i] = style.Render(strings.Repeat("─", col.Width-2))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sepCells...))
	b.WriteString("\n")

	// Rows
	for _, row := range rows {
		cells := make([]string, len(columns))
		for i, col := range columns {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			style := lipgloss.NewStyle().
				Width(col.Width).
				Foreground(Theme.Text).
				Padding(0, 1)
			cells[i] = style.Render(val)
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		b.WriteString("\n")
	}

	return b.String()
}

func KeyValue(pairs [][2]string) string {
	maxLabel := 0
	for _, p := range pairs {
		if len(p[0]) > maxLabel {
			maxLabel = len(p[0])
		}
	}

	var b strings.Builder
	for _, p := range pairs {
		label := SubtitleStyle.Render(fmt.Sprintf("  %-*s", maxLabel+1, p[0]+":"))
		value := TextStyle.Render(p[1])
		b.WriteString(label + " " + value + "\n")
	}
	return b.String()
}

func StatusIcon(success bool) string {
	if success {
		return AccentStyle.Render("✓")
	}
	return ErrorStyle.Render("✗")
}
