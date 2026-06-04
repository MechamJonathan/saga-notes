// Package steps reads daily step counts from a configurable source: manual
// entry, a Health Auto Export folder, or a native Apple Health export.xml.
package steps

import (
	"time"

	"almanac/internal/config"
)

// Steps is a single day's step count and goal.
type Steps struct {
	Date  time.Time
	Count int
	Goal  int
}

// Pct returns progress toward the goal as a 0..1 fraction.
func (s Steps) Pct() float64 {
	if s.Goal <= 0 {
		return 0
	}
	p := float64(s.Count) / float64(s.Goal)
	if p > 1 {
		return 1
	}
	return p
}

// Reader returns the step count for a given day.
type Reader interface {
	Today(day time.Time) (Steps, error)
}

// New builds a Reader from configuration.
//
// manualLookup supplies the count for the "manual" source (typically reading
// from local state). cacheDir is where the appleexport source writes its parse
// cache.
func New(cfg config.StepsConfig, manualLookup func(time.Time) int, cacheDir string) Reader {
	switch cfg.Source {
	case "autoexport":
		return autoExportReader{dir: cfg.Path, goal: cfg.Goal}
	case "appleexport":
		return appleExportReader{xmlPath: cfg.Path, goal: cfg.Goal, cacheDir: cacheDir}
	default: // "manual"
		return manualReader{lookup: manualLookup, goal: cfg.Goal}
	}
}

type manualReader struct {
	lookup func(time.Time) int
	goal   int
}

func (r manualReader) Today(day time.Time) (Steps, error) {
	count := 0
	if r.lookup != nil {
		count = r.lookup(day)
	}
	return Steps{Date: day, Count: count, Goal: r.goal}, nil
}
