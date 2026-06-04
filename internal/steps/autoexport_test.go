package steps

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAutoExportCSV(t *testing.T) {
	dir := t.TempDir()
	csv := "Date,Step Count\n2026-06-04,\"7,432\"\n2026-06-04,68\n2026-06-03,5000\n"
	if err := os.WriteFile(filepath.Join(dir, "steps.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}

	r := autoExportReader{dir: dir, goal: 10000}
	s, err := r.Today(time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Count != 7500 { // 7432 + 68
		t.Errorf("CSV count = %d, want 7500", s.Count)
	}
}

func TestAutoExportJSON(t *testing.T) {
	dir := t.TempDir()
	js := `[{"date":"2026-06-04T00:00:00-06:00","qty":7432},{"date":"2026-06-03","qty":5000}]`
	if err := os.WriteFile(filepath.Join(dir, "steps.json"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}

	r := autoExportReader{dir: dir, goal: 10000}
	s, _ := r.Today(time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC))
	if s.Count != 7432 {
		t.Errorf("JSON count = %d, want 7432", s.Count)
	}
}

func TestParseStepValue(t *testing.T) {
	cases := map[string]int{"7432": 7432, "7,432": 7432, "7432.0": 7432, "": 0, "x": 0}
	for in, want := range cases {
		if got := parseStepValue(in); got != want {
			t.Errorf("parseStepValue(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestStepsPct(t *testing.T) {
	if p := (Steps{Count: 5000, Goal: 10000}).Pct(); p != 0.5 {
		t.Errorf("Pct = %v, want 0.5", p)
	}
	if p := (Steps{Count: 20000, Goal: 10000}).Pct(); p != 1 {
		t.Errorf("Pct clamps to 1, got %v", p)
	}
	if p := (Steps{Count: 100, Goal: 0}).Pct(); p != 0 {
		t.Errorf("Pct with zero goal = %v, want 0", p)
	}
}
