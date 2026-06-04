package quotes

import (
	"testing"
	"time"
)

func TestOfDayDeterministic(t *testing.T) {
	day := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	a := OfDay(day)
	b := OfDay(day.Add(3 * time.Hour)) // same calendar day
	if a != b {
		t.Errorf("OfDay not stable within a day: %v vs %v", a, b)
	}
	if a.Text == "" {
		t.Error("OfDay returned empty text")
	}
}

func TestOfDayRotates(t *testing.T) {
	day := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	seen := map[string]bool{}
	distinct := 0
	for i := 0; i < len(all); i++ {
		q := OfDay(day.AddDate(0, 0, i))
		if !seen[q.Text] {
			seen[q.Text] = true
			distinct++
		}
	}
	if distinct < 2 {
		t.Errorf("expected quotes to rotate, only saw %d distinct over %d days", distinct, len(all))
	}
}
