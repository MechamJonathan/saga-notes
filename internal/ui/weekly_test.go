package ui

import (
	"strings"
	"testing"
	"time"

	"saga-notes/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// --- weekStart ---

func TestWeekStart(t *testing.T) {
	cases := []struct {
		date    string
		wantMon string
	}{
		{"2026-06-15", "2026-06-15"}, // Monday → itself
		{"2026-06-17", "2026-06-15"}, // Wednesday → prior Monday
		{"2026-06-20", "2026-06-15"}, // Saturday → prior Monday
		{"2026-06-21", "2026-06-15"}, // Sunday   → prior Monday (not next)
		{"2026-06-22", "2026-06-22"}, // Monday again → itself
	}
	for _, tc := range cases {
		t.Run(tc.date, func(t *testing.T) {
			d, _ := time.Parse("2006-01-02", tc.date)
			got := weekStart(d).Format("2006-01-02")
			if got != tc.wantMon {
				t.Errorf("weekStart(%s) = %s, want %s", tc.date, got, tc.wantMon)
			}
		})
	}
}

// --- ratingDots ---

func TestRatingDots(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "–"},
		{1, "●○○○○"},
		{2, "●●○○○"},
		{3, "●●●○○"},
		{4, "●●●●○"},
		{5, "●●●●●"},
	}
	for _, tc := range cases {
		if got := ratingDots(tc.in); got != tc.want {
			t.Errorf("ratingDots(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// --- weeklyHabitsBar ---

func TestWeeklyHabitsBarNoHabits(t *testing.T) {
	if got := weeklyHabitsBar(storage.DayEntry{}, 0); got != "–" {
		t.Errorf("no habits: got %q, want –", got)
	}
}

func TestWeeklyHabitsBarAllDone(t *testing.T) {
	e := storage.DayEntry{NonNegs: []bool{true, true, true}}
	got := weeklyHabitsBar(e, 3)
	if !strings.Contains(got, "3/3") {
		t.Errorf("all done: got %q, want string containing 3/3", got)
	}
	if strings.Contains(got, "░") {
		t.Errorf("all done: got %q, should contain no empty blocks", got)
	}
}

func TestWeeklyHabitsBarPartial(t *testing.T) {
	e := storage.DayEntry{NonNegs: []bool{true, false, true}}
	got := weeklyHabitsBar(e, 3)
	if !strings.Contains(got, "2/3") {
		t.Errorf("partial: got %q, want string containing 2/3", got)
	}
	if !strings.Contains(got, "░") {
		t.Errorf("partial: got %q, should contain at least one empty block", got)
	}
}

func TestWeeklyHabitsBarNoneDone(t *testing.T) {
	e := storage.DayEntry{NonNegs: []bool{false, false}}
	got := weeklyHabitsBar(e, 2)
	if !strings.Contains(got, "0/2") {
		t.Errorf("none done: got %q, want string containing 0/2", got)
	}
}

// --- sparkline ---

func TestSparkline(t *testing.T) {
	data := []storage.DayEntry{
		{Mood: 0}, // unset → ·
		{Mood: 1}, // lowest → ▁
		{Mood: 5}, // max   → █
	}
	got := sparkline(data, func(e storage.DayEntry) int { return e.Mood })
	want := "·▁█"
	if got != want {
		t.Errorf("sparkline = %q, want %q", got, want)
	}
}

func TestSparklineAllUnset(t *testing.T) {
	data := make([]storage.DayEntry, 5)
	got := sparkline(data, func(e storage.DayEntry) int { return e.Mood })
	if got != "·····" {
		t.Errorf("all unset: got %q, want %q", got, "·····")
	}
}

// --- weeklyModel.update ---

func TestWeeklyUpdateCursorDown(t *testing.T) {
	m := weeklyModel{cursor: 3, styles: NewStyles("#2de2d2")}
	m2, day := m.update(tea.KeyMsg{Type: tea.KeyDown})
	if m2.cursor != 4 {
		t.Errorf("down: cursor = %d, want 4", m2.cursor)
	}
	if !day.IsZero() {
		t.Error("down should not return a navigation day")
	}
}

func TestWeeklyUpdateCursorUp(t *testing.T) {
	m := weeklyModel{cursor: 3, styles: NewStyles("#2de2d2")}
	m2, _ := m.update(tea.KeyMsg{Type: tea.KeyUp})
	if m2.cursor != 2 {
		t.Errorf("up: cursor = %d, want 2", m2.cursor)
	}
}

func TestWeeklyUpdateCursorBoundaries(t *testing.T) {
	top := weeklyModel{cursor: 0, styles: NewStyles("#2de2d2")}
	m2, _ := top.update(tea.KeyMsg{Type: tea.KeyUp})
	if m2.cursor != 0 {
		t.Error("up at Mon should stay at 0")
	}

	bot := weeklyModel{cursor: 6, styles: NewStyles("#2de2d2")}
	m3, _ := bot.update(tea.KeyMsg{Type: tea.KeyDown})
	if m3.cursor != 6 {
		t.Error("down at Sun should stay at 6")
	}
}

func TestWeeklyUpdateEnterReturnsDay(t *testing.T) {
	monday, _ := time.Parse("2006-01-02", "2026-06-15")
	m := weeklyModel{cursor: 2, anchor: monday, styles: NewStyles("#2de2d2")}
	_, day := m.update(tea.KeyMsg{Type: tea.KeyEnter})
	want := "2026-06-17" // Monday + 2 days = Wednesday
	if got := day.Format("2006-01-02"); got != want {
		t.Errorf("enter returned %s, want %s", got, want)
	}
}

func TestWeeklyShiftWeek(t *testing.T) {
	monday, _ := time.Parse("2006-01-02", "2026-06-15")
	m := weeklyModel{anchor: monday, styles: NewStyles("#2de2d2")}
	fwd := m.shiftWeek(1)
	if got := fwd.anchor.Format("2006-01-02"); got != "2026-06-22" {
		t.Errorf("shift +1 anchor = %s, want 2026-06-22", got)
	}
	back := m.shiftWeek(-1)
	if got := back.anchor.Format("2006-01-02"); got != "2026-06-08" {
		t.Errorf("shift -1 anchor = %s, want 2026-06-08", got)
	}
}

// --- integration: tab cycling and weekly view rendering ---

func TestTabCyclesThreeFoci(t *testing.T) {
	m := newTestModel(t)
	if m.focus != focusGoals {
		t.Fatalf("initial focus = %v, want focusGoals", m.focus)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	if m.focus != focusNotes {
		t.Errorf("after 1st tab: focus = %v, want focusNotes", m.focus)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	if m.focus != focusWeek {
		t.Errorf("after 2nd tab: focus = %v, want focusWeek", m.focus)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	if m.focus != focusGoals {
		t.Errorf("after 3rd tab: focus = %v, want focusGoals (full cycle)", m.focus)
	}
}

func TestWeeklyViewContainsSections(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	next, _ = next.(model).Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	out := m.View()
	for _, want := range []string{"WEEK OF", "mood", "energy", "TREND"} {
		if !strings.Contains(out, want) {
			t.Errorf("weekly View() missing %q", want)
		}
	}
}

func TestWeeklyFooterHints(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	next, _ = next.(model).Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	out := m.View()
	for _, want := range []string{"enter open day", "[ ] week"} {
		if !strings.Contains(out, want) {
			t.Errorf("weekly footer missing %q", want)
		}
	}
}

func TestWeeklyBracketShiftsWeek(t *testing.T) {
	m := newTestModel(t)
	// Navigate to weekly focus
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	next, _ = next.(model).Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(model)
	anchor := m.weekly.anchor

	// ] should advance the week
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = next.(model)
	if !m.weekly.anchor.After(anchor) {
		t.Error("] in weekly mode should advance the anchor week")
	}

	// [ should retreat
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = next.(model)
	if m.weekly.anchor != anchor {
		t.Errorf("[ after ] should restore original anchor, got %s want %s",
			m.weekly.anchor.Format("2006-01-02"), anchor.Format("2006-01-02"))
	}
}
