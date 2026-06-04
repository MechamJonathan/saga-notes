package ui

import (
	"fmt"
	"strings"
	"time"

	"almanac/internal/astro"
	"almanac/internal/steps"
)

// renderHeader draws the top bar: date · clock · moon · compact steps.
func renderHeader(s Styles, now time.Time, st steps.Steps, stepsKnown bool) string {
	moon := astro.MoonPhase(now)
	parts := []string{
		s.Header.Render("almanac"),
		now.Format("Mon, Jan 2"),
		now.Format("15:04"),
		moon.Glyph + " " + moon.Name,
	}
	if stepsKnown {
		parts = append(parts, fmt.Sprintf("👟 %s/%s", humanCount(st.Count), humanGoal(st.Goal)))
	}
	return s.Faint.Render(strings.Join(parts, "  ·  "))
}

// humanCount renders a step count with thousands separators (e.g. 7,432).
func humanCount(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}

// humanGoal renders a goal compactly (10000 -> 10k).
func humanGoal(n int) string {
	if n >= 1000 && n%1000 == 0 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return humanCount(n)
}
