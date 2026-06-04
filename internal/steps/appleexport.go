package steps

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// appleExportReader parses a native Apple Health export.xml. Because that file
// is often hundreds of megabytes, it parses once and caches a per-day step sum,
// re-parsing only when the source file's modification time changes.
type appleExportReader struct {
	xmlPath  string
	goal     int
	cacheDir string
}

// stepsCache is the persisted parse result.
type stepsCache struct {
	SourceMod time.Time      `json:"source_mod"`
	ByDay     map[string]int `json:"by_day"` // YYYY-MM-DD -> steps
}

const dateKey = "2006-01-02"

func (r appleExportReader) Today(day time.Time) (Steps, error) {
	out := Steps{Date: day, Goal: r.goal}
	if r.xmlPath == "" {
		return out, nil
	}

	info, err := os.Stat(r.xmlPath)
	if err != nil {
		return out, err
	}

	cache, ok := r.loadCache()
	if !ok || !cache.SourceMod.Equal(info.ModTime()) {
		cache, err = r.parse(info.ModTime())
		if err != nil {
			return out, err
		}
		r.saveCache(cache)
	}

	out.Count = cache.ByDay[day.Format(dateKey)]
	return out, nil
}

func (r appleExportReader) cachePath() string {
	return filepath.Join(r.cacheDir, "steps-cache.json")
}

func (r appleExportReader) loadCache() (stepsCache, bool) {
	data, err := os.ReadFile(r.cachePath())
	if err != nil {
		return stepsCache{}, false
	}
	var c stepsCache
	if err := json.Unmarshal(data, &c); err != nil {
		return stepsCache{}, false
	}
	if c.ByDay == nil {
		c.ByDay = map[string]int{}
	}
	return c, true
}

func (r appleExportReader) saveCache(c stepsCache) {
	if err := os.MkdirAll(r.cacheDir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(r.cachePath(), data, 0o644)
}

// parse streams the export.xml and sums step-count records by start day.
func (r appleExportReader) parse(mod time.Time) (stepsCache, error) {
	cache := stepsCache{SourceMod: mod, ByDay: map[string]int{}}

	f, err := os.Open(r.xmlPath)
	if err != nil {
		return cache, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cache, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Record" {
			continue
		}
		sumStepRecord(se, cache.ByDay)
	}
	return cache, nil
}

// appleDateFormat matches the timestamps Apple writes, e.g.
// "2026-06-04 06:14:00 -0600".
const appleDateFormat = "2006-01-02 15:04:05 -0700"

// sumStepRecord adds a single <Record> element's value to the per-day map when
// it is a step-count record.
func sumStepRecord(se xml.StartElement, byDay map[string]int) {
	var typ, value, start string
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "type":
			typ = a.Value
		case "value":
			value = a.Value
		case "startDate":
			start = a.Value
		}
	}
	if typ != "HKQuantityTypeIdentifierStepCount" {
		return
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return
	}
	t, err := time.Parse(appleDateFormat, start)
	if err != nil {
		return
	}
	byDay[t.Format(dateKey)] += n
}
