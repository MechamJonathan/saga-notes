package ui

import (
	"strings"
	"time"

	"saga-notes/internal/astro"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader draws the top bar: app name on the left, date/clock/moon on the right.
func renderHeader(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	left := s.Header.Render("⚔  Saga Notes")
	right := s.Faint.Render(strings.Join([]string{
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}, "  ·  "))
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}
