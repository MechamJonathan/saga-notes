// Package astro provides small, dependency-free astronomical calculations.
package astro

import (
	"math"
	"time"
)

// synodicMonth is the average length of one lunar phase cycle, in days.
const synodicMonth = 29.53058867

// knownNewMoon is a reference new moon: 2000-01-06 18:14 UTC.
var knownNewMoon = time.Date(2000, time.January, 6, 18, 14, 0, 0, time.UTC)

// Phase describes the moon's appearance on a given day.
type Phase struct {
	// Fraction is the position in the cycle, 0..1 (0 = new, 0.5 = full).
	Fraction float64
	// Name is the human-readable phase name.
	Name string
	// Glyph is a single-rune emoji for the phase.
	Glyph string
}

// MoonPhase returns the moon phase for the given time.
func MoonPhase(t time.Time) Phase {
	days := t.UTC().Sub(knownNewMoon).Hours() / 24
	frac := math.Mod(days/synodicMonth, 1)
	if frac < 0 {
		frac++
	}

	// Eight phases, each spanning 1/8 of the cycle, centered on its midpoint.
	idx := int(math.Floor(frac*8+0.5)) % 8
	return Phase{Fraction: frac, Name: phaseNames[idx], Glyph: phaseGlyphs[idx]}
}

var phaseNames = [8]string{
	"New Moon",
	"Waxing Crescent",
	"First Quarter",
	"Waxing Gibbous",
	"Full Moon",
	"Waning Gibbous",
	"Last Quarter",
	"Waning Crescent",
}

var phaseGlyphs = [8]string{
	"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘",
}
