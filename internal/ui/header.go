package ui

import (
	"time"

	"saga-notes/internal/astro"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader draws a full-width teal bar with date, time, and moon phase.
func renderHeader(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	text := " " + now.Format("Mon, Jan 2") + "   " + now.Format("15:04") + "   " + moon.Name
	return lipgloss.NewStyle().
		Background(s.Accent).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Width(width).
		Render(text)
}
