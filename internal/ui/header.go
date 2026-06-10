package ui

import (
	"strings"
	"time"

	"saga-notes/internal/astro"

	lipgloss "github.com/charmbracelet/lipgloss"
)

// renderHeader draws a full-width teal bar: app name left, date/clock/moon right.
func renderHeader(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	info := strings.Join([]string{
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}, "  ·  ")
	title := "Saga Notes"
	gap := max(1, width-len(title)-len(info)-2) // -2 for leading space on each side
	content := " " + title + strings.Repeat(" ", gap) + info + " "
	return lipgloss.NewStyle().
		Background(s.Accent).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Render(content)
}
