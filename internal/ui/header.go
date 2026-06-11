package ui

import (
	"time"

	"saga-notes/internal/astro"

	"github.com/charmbracelet/lipgloss"
)

// headerOuterH is the number of terminal rows the header panel occupies.
const headerOuterH = 3 // top border + 1 content row + bottom border

// renderHeaderPanel draws a full-width bordered panel with date, time, and moon phase.
// Placing the text inside a border means it lands on row 1, not row 0.
func renderHeaderPanel(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	text := now.Format("Mon, Jan 2") + "   " + now.Format("15:04") + "   " + moon.Glyph + "  " + moon.Name
	content := lipgloss.NewStyle().Foreground(s.Accent).Bold(true).Render(text)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Dim).
		Padding(0, 2).
		Width(width - borderW).
		Render(content)
}
