package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpSection struct {
	title string
	rows  [][2]string
}

var helpSections = []helpSection{
	{
		title: "GLOBAL",
		rows: [][2]string{
			{"tab", "cycle focus  goals → notes → week"},
			{"[ / ]", "previous / next day"},
			{"t", "jump to today"},
			{"w", "refresh weather"},
			{"?", "toggle this help"},
			{"q", "quit"},
		},
	},
	{
		title: "GOALS  (left panel)",
		rows: [][2]string{
			{"↑↓ / j k", "navigate active goals"},
			{"space", "toggle done"},
			{"a", "add new goal"},
			{"e", "edit selected goal"},
			{"d", "delete selected goal"},
			{"c", "clear all completed goals"},
		},
	},
	{
		title: "DAILY JOURNAL  (right panel)",
		rows: [][2]string{
			{"↑↓", "navigate sections"},
			{"space", "toggle habit done"},
			{"a / e / d", "add / edit / delete habit label"},
			{"1–5", "set mood or energy rating"},
			{"i / enter", "write note inline"},
			{"e", "open note in $EDITOR  (on notes row)"},
		},
	},
	{
		title: "WEEKLY VIEW  (tab twice)",
		rows: [][2]string{
			{"↑↓ / j k", "navigate days"},
			{"enter", "open selected day in journal"},
			{"[ / ]", "previous / next week"},
			{"t", "jump to current week"},
		},
	},
}

func renderHelp(s Styles, width, height int) string {
	keyW := 12 // fixed width for the key column

	var b strings.Builder
	for i, sec := range helpSections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(s.Title.Render(sec.title))
		b.WriteString("\n")
		for _, row := range sec.rows {
			key := lipgloss.NewStyle().
				Foreground(s.Accent).Bold(true).
				Width(keyW).
				Render(row[0])
			b.WriteString("  " + key + "  " + s.Faint.Render(row[1]) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(s.Faint.Render("  press ? or esc to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Accent).
		Padding(1, 3).
		Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
