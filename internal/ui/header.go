package ui

import (
	"strings"
	"time"

	"saga-notes/internal/astro"
)

// renderHeader draws a faint top line showing date, clock, and moon phase.
func renderHeader(s Styles, now time.Time, width int) string {
	moon := astro.MoonPhase(now)
	content := strings.Join([]string{
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}, "  ·  ")
	return s.Faint.Render(content)
}
