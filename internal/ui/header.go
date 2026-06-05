package ui

import (
	"strings"
	"time"

	"saga-notes/internal/astro"
)

// renderHeader draws the top bar: app name · date · clock · moon.
func renderHeader(s Styles, now time.Time) string {
	moon := astro.MoonPhase(now)
	parts := []string{
		s.Header.Render("Saga Notes"),
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}
	return s.Faint.Render(strings.Join(parts, "  ·  "))
}
