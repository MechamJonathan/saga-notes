package steps

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sampleXML = `<?xml version="1.0" encoding="UTF-8"?>
<HealthData>
  <Record type="HKQuantityTypeIdentifierStepCount" value="1200" startDate="2026-06-04 06:00:00 -0600" endDate="2026-06-04 07:00:00 -0600"/>
  <Record type="HKQuantityTypeIdentifierStepCount" value="3000" startDate="2026-06-04 08:00:00 -0600" endDate="2026-06-04 09:00:00 -0600"/>
  <Record type="HKQuantityTypeIdentifierHeartRate" value="72" startDate="2026-06-04 08:00:00 -0600" endDate="2026-06-04 08:01:00 -0600"/>
  <Record type="HKQuantityTypeIdentifierStepCount" value="500" startDate="2026-06-03 20:00:00 -0600" endDate="2026-06-03 20:30:00 -0600"/>
</HealthData>`

func TestAppleExportSumsByDay(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "export.xml")
	if err := os.WriteFile(xmlPath, []byte(sampleXML), 0o644); err != nil {
		t.Fatal(err)
	}

	r := appleExportReader{xmlPath: xmlPath, goal: 10000, cacheDir: dir}

	day := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	s, err := r.Today(day)
	if err != nil {
		t.Fatal(err)
	}
	if s.Count != 4200 { // 1200 + 3000, heart rate ignored
		t.Errorf("June 4 count = %d, want 4200", s.Count)
	}

	prev, _ := r.Today(day.AddDate(0, 0, -1))
	if prev.Count != 500 {
		t.Errorf("June 3 count = %d, want 500", prev.Count)
	}

	// A cache file should now exist and be reused without the source present.
	if _, err := os.Stat(filepath.Join(dir, "steps-cache.json")); err != nil {
		t.Errorf("expected steps-cache.json to be written: %v", err)
	}
}

func TestAppleExportNoPath(t *testing.T) {
	r := appleExportReader{goal: 10000}
	s, err := r.Today(time.Now())
	if err != nil {
		t.Fatalf("unexpected error with no path: %v", err)
	}
	if s.Count != 0 || s.Goal != 10000 {
		t.Errorf("got %+v, want zero count with goal 10000", s)
	}
}
