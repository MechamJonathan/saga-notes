// Package export writes saga journal data to portable text formats.
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"saga-notes/internal/storage"
)

// ToMarkdown writes all journal data to a Markdown file in dir and returns
// the output path. The file is named saga-export-YYYY-MM-DD.md using today's date.
func ToMarkdown(state storage.State, nonNegs []string, dir string) (string, error) {
	days, err := allDays()
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString("# Saga Journal Export\n\n")
	b.WriteString(fmt.Sprintf("_Exported %s_\n\n", time.Now().Format("2006-01-02 15:04")))

	// Goals section
	if len(state.Goals) > 0 || len(state.CompletedGoals) > 0 {
		b.WriteString("---\n\n## Goals\n\n")
		for _, g := range state.Goals {
			check := " "
			if g.Done {
				check = "x"
			}
			b.WriteString(fmt.Sprintf("- [%s] %s\n", check, g.Text))
		}
		if len(state.CompletedGoals) > 0 {
			b.WriteString("\n### Completed\n\n")
			for _, g := range state.CompletedGoals {
				b.WriteString(fmt.Sprintf("- [x] %s  _(completed %s)_\n", g.Text, g.CompletedAt.Format("2006-01-02")))
			}
		}
		b.WriteString("\n")
	}

	// Daily entries
	for _, day := range days {
		entry, _ := storage.LoadDay(day)
		note, _ := storage.LoadNote(day)

		hasData := entry.Mood != 0 || entry.Energy != 0 || strings.TrimSpace(note) != ""
		for _, v := range entry.NonNegs {
			if v {
				hasData = true
				break
			}
		}
		if !hasData {
			continue
		}

		b.WriteString("---\n\n")
		b.WriteString(fmt.Sprintf("## %s\n\n", day.Format("2006-01-02  Monday")))

		if entry.Mood != 0 || entry.Energy != 0 {
			mood := "—"
			energy := "—"
			if entry.Mood != 0 {
				mood = fmt.Sprintf("%d/5", entry.Mood)
			}
			if entry.Energy != 0 {
				energy = fmt.Sprintf("%d/5", entry.Energy)
			}
			b.WriteString(fmt.Sprintf("**Mood:** %s  **Energy:** %s\n\n", mood, energy))
		}

		entry = entry.EnsureNonNegs(len(nonNegs))
		if len(nonNegs) > 0 {
			b.WriteString("### Daily Habits\n\n")
			for i, label := range nonNegs {
				check := " "
				if i < len(entry.NonNegs) && entry.NonNegs[i] {
					check = "x"
				}
				b.WriteString(fmt.Sprintf("- [%s] %s\n", check, label))
			}
			b.WriteString("\n")
		}

		if note := strings.TrimSpace(note); note != "" {
			b.WriteString("### Notes\n\n")
			b.WriteString(note)
			b.WriteString("\n\n")
		}
	}

	name := fmt.Sprintf("saga-export-%s.md", time.Now().Format("2006-01-02"))
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// allDays returns the union of days with notes and days with structured
// entries, sorted chronologically.
func allDays() ([]time.Time, error) {
	notes, err := storage.ListNotes()
	if err != nil {
		return nil, err
	}
	entries, err := storage.ListDays()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]time.Time)
	for _, d := range notes {
		seen[d.Format(storage.DateKey)] = d
	}
	for _, d := range entries {
		seen[d.Format(storage.DateKey)] = d
	}

	days := make([]time.Time, 0, len(seen))
	for _, d := range seen {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	return days, nil
}
