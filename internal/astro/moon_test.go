package astro

import (
	"testing"
	"time"
)

// daysDur converts a fractional number of days to a time.Duration.
func daysDur(days float64) time.Duration {
	return time.Duration(days * 24 * float64(time.Hour))
}

func TestMoonPhaseKnownDates(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"reference new moon", knownNewMoon, "New Moon"},
		{"full moon ~half cycle later", knownNewMoon.Add(daysDur(synodicMonth / 2)), "Full Moon"},
		{"first quarter ~quarter cycle later", knownNewMoon.Add(daysDur(synodicMonth / 4)), "First Quarter"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MoonPhase(tc.date)
			if got.Name != tc.want {
				t.Errorf("MoonPhase(%v).Name = %q, want %q", tc.date, got.Name, tc.want)
			}
			if got.Glyph == "" {
				t.Errorf("MoonPhase(%v).Glyph is empty", tc.date)
			}
		})
	}
}

func TestMoonPhaseFractionInRange(t *testing.T) {
	d := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 60; i++ {
		p := MoonPhase(d.AddDate(0, 0, i))
		if p.Fraction < 0 || p.Fraction >= 1 {
			t.Fatalf("fraction out of range on day %d: %v", i, p.Fraction)
		}
	}
}
