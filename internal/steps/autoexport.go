package steps

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// autoExportReader reads step data from a folder of files written by the
// "Health Auto Export" app. It scans .csv and .json files in the folder and
// sums any rows/records that fall on the requested day.
type autoExportReader struct {
	dir  string
	goal int
}

func (r autoExportReader) Today(day time.Time) (Steps, error) {
	out := Steps{Date: day, Goal: r.goal}
	if r.dir == "" {
		return out, nil
	}
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return out, err
	}

	target := day.Format(dateKey)
	total := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(r.dir, e.Name())
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".csv":
			total += sumCSV(path, target)
		case ".json":
			total += sumJSON(path, target)
		}
	}
	out.Count = total
	return out, nil
}

// sumCSV reads a CSV with a header row, locating a date-ish column and a
// step-ish column, and sums step values whose date matches target.
func sumCSV(path, target string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	rd := csv.NewReader(f)
	rd.FieldsPerRecord = -1
	rows, err := rd.ReadAll()
	if err != nil || len(rows) < 2 {
		return 0
	}

	dateCol, stepCol := -1, -1
	for i, h := range rows[0] {
		lh := strings.ToLower(h)
		if dateCol == -1 && strings.Contains(lh, "date") {
			dateCol = i
		}
		if stepCol == -1 && strings.Contains(lh, "step") {
			stepCol = i
		}
	}
	if dateCol == -1 || stepCol == -1 {
		return 0
	}

	total := 0
	for _, row := range rows[1:] {
		if dateCol >= len(row) || stepCol >= len(row) {
			continue
		}
		if normalizeDate(row[dateCol]) != target {
			continue
		}
		total += parseStepValue(row[stepCol])
	}
	return total
}

// jsonRecord is a permissive shape covering common Health Auto Export JSON.
type jsonRecord struct {
	Date  string  `json:"date"`
	Qty   float64 `json:"qty"`
	Steps float64 `json:"steps"`
}

func sumJSON(path, target string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	// Try a top-level array first, then a {"data": [...]} wrapper.
	var recs []jsonRecord
	if err := json.Unmarshal(data, &recs); err != nil {
		var wrapper struct {
			Data []jsonRecord `json:"data"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return 0
		}
		recs = wrapper.Data
	}

	total := 0.0
	for _, rec := range recs {
		if normalizeDate(rec.Date) != target {
			continue
		}
		if rec.Qty > 0 {
			total += rec.Qty
		} else {
			total += rec.Steps
		}
	}
	return int(math.Round(total))
}

// normalizeDate extracts the YYYY-MM-DD prefix from a variety of timestamp
// formats (ISO, with time, with timezone, etc.).
func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// parseStepValue parses a step count that may be written as "7432", "7,432",
// or "7432.0".
func parseStepValue(s string) int {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(math.Round(f))
	}
	return 0
}
