// Package storage persists saga-notes local-first state: goals, cached live
// data (in state.json), and per-day notes as Markdown files.
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// DateKey is the canonical date format used for per-day data.
const DateKey = "2006-01-02"

// Goal is a single daily goal item.
type Goal struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// DayEntry holds per-day structured journal fields: mood, energy, and
// non-negotiable completion states.
type DayEntry struct {
	Mood    int    `json:"mood"`     // 0 = unset, 1–5
	Energy  int    `json:"energy"`   // 0 = unset, 1–5
	NonNegs []bool `json:"non_negs"` // parallel to config.Journal.NonNegotiables
}

// EnsureNonNegs pads or trims NonNegs to match n configured non-negotiables.
func (e DayEntry) EnsureNonNegs(n int) DayEntry {
	for len(e.NonNegs) < n {
		e.NonNegs = append(e.NonNegs, false)
	}
	if len(e.NonNegs) > n {
		e.NonNegs = e.NonNegs[:n]
	}
	return e
}

// WeatherCache stores the last successful weather fetch.
type WeatherCache struct {
	City      string    `json:"city"`
	TempNow   float64   `json:"temp_now"`
	TempHigh  float64   `json:"temp_high"`
	TempLow   float64   `json:"temp_low"`
	Desc      string    `json:"desc"`
	Icon      string    `json:"icon"`
	Pop       int       `json:"pop"` // probability of precipitation, %
	FetchedAt time.Time `json:"fetched_at"`
}

// State is the JSON-serialized application state.
type State struct {
	Goals          []Goal        `json:"goals"`
	Weather        *WeatherCache `json:"weather,omitempty"`
	NonNegotiables []string      `json:"non_negotiables,omitempty"`
}

// Dir returns the saga-notes data directory (~/.local/share/saga-notes on Linux,
// ~/Library/Application Support/saga-notes on macOS).
func Dir() (string, error) {
	base, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Prefer XDG_DATA_HOME when set; otherwise use a stable per-OS location.
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "saga-notes"), nil
	}
	return filepath.Join(base, ".local", "share", "saga-notes"), nil
}

func statePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

// Load reads state.json, returning an empty (initialized) State if none exists.
func Load() (State, error) {
	s := State{}
	path, err := statePath()
	if err != nil {
		return s, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, err
	}
	return s, nil
}

// Save writes state.json atomically, creating the data dir as needed.
func Save(s State) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// dayEntryPath returns the path for a per-day entry JSON file.
func dayEntryPath(day time.Time) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "days", day.Format(DateKey)+".json"), nil
}

// LoadDay reads the structured day entry for a given day, returning an empty
// entry (without error) if none exists yet.
func LoadDay(day time.Time) (DayEntry, error) {
	path, err := dayEntryPath(day)
	if err != nil {
		return DayEntry{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return DayEntry{}, nil
	}
	if err != nil {
		return DayEntry{}, err
	}
	var e DayEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return DayEntry{}, err
	}
	return e, nil
}

// SaveDay writes the structured day entry atomically.
func SaveDay(day time.Time, e DayEntry) error {
	path, err := dayEntryPath(day)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// notesDir returns the directory holding per-day note files.
func notesDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes"), nil
}

// NotePath returns the Markdown file path for the given day.
func NotePath(day time.Time) (string, error) {
	dir, err := notesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, day.Format(DateKey)+".md"), nil
}

// LoadNote reads the note for a day, returning "" if the file does not exist.
func LoadNote(day time.Time) (string, error) {
	path, err := NotePath(day)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ComputeNonNegStreaks returns the current consecutive-day streak for each of
// the n non-negotiables, relative to today. If today's entry isn't done yet,
// the count reflects yesterday-and-back so the user sees what they stand to lose.
func ComputeNonNegStreaks(n int, today time.Time) []int {
	streaks := make([]int, n)
	for i := range streaks {
		day := today
		skippedToday := false
		for {
			e, _ := LoadDay(day)
			done := i < len(e.NonNegs) && e.NonNegs[i]
			if !done {
				if !skippedToday && day.Equal(today) {
					skippedToday = true
					day = day.AddDate(0, 0, -1)
					continue
				}
				break
			}
			streaks[i]++
			day = day.AddDate(0, 0, -1)
			if streaks[i] >= 365 {
				break
			}
		}
	}
	return streaks
}

// SaveNote writes (or removes, if empty) the note for a day.
func SaveNote(day time.Time, body string) error {
	path, err := NotePath(day)
	if err != nil {
		return err
	}
	if body == "" {
		// Don't leave empty files lying around.
		if rerr := os.Remove(path); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			return rerr
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
