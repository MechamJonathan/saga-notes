package ui

import (
	"strings"
	"time"

	"saga-notes/internal/astro"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader draws a full-width teal bar: app name left, date/clock/moon right.
func renderHeader(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	left := "⚔  Saga Notes"
	right := strings.Join([]string{
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}, "  ·  ")
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	content := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().
		Background(teal).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Width(width).
		Render(content)
}
