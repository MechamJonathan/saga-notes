package ui

import (
	"strings"
	"testing"

	"saga-notes/internal/config"
	"saga-notes/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestModel builds a model and gives it a window size so View renders.
func newTestModel(t *testing.T) model {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	cfg := config.Default()
	state := storage.State{
		Goals: []storage.Goal{{Text: "Ship the plan"}, {Text: "Walk the dog", Done: true}},
	}
	m := New(cfg, state)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return next.(model)
}

func TestViewContainsSections(t *testing.T) {
	m := newTestModel(t)
	out := m.View()
	for _, want := range []string{
		"Saga Notes", "GOALS", "WEATHER", "NON-NEGOTIABLES", "NOTES", "Ship the plan",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestViewContainsNonNeg(t *testing.T) {
	m := newTestModel(t)
	out := m.View()
	for _, want := range []string{"SLEPT 7+ HOURS", "READ GOALS", "HYDRATED"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing non-neg %q", want)
		}
	}
}

func TestViewContainsDaySel(t *testing.T) {
	m := newTestModel(t)
	out := m.View()
	// One of the day labels must be highlighted in brackets.
	for _, d := range []string{"[S]", "[M]", "[T]", "[W]", "[T]", "[F]", "[S]"} {
		if strings.Contains(out, d) {
			return // found one bracket highlight — good
		}
	}
	t.Error("View() should show a bracketed day-of-week selector")
}

func TestViewMoodEnergy(t *testing.T) {
	m := newTestModel(t)
	out := m.View()
	for _, want := range []string{"MOOD", "ENERGY"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestViewStacksWhenNarrow(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 40})
	out := next.(model).View()
	if !strings.Contains(out, "GOALS") || !strings.Contains(out, "NON-NEGOTIABLES") {
		t.Error("narrow View() should still show both panels stacked")
	}
}

func TestTabSwitchesFocus(t *testing.T) {
	m := newTestModel(t)
	if m.focus != focusGoals {
		t.Fatalf("initial focus = %v, want goals", m.focus)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if next.(model).focus != focusNotes {
		t.Error("tab should switch focus to notes/daily panel")
	}
}

func TestAddGoalFlow(t *testing.T) {
	m := newTestModel(t)
	mi, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mi.(model)
	if !m.goals.editing() {
		t.Fatal("pressing 'a' should enter goal-add mode")
	}
	for _, r := range "Read" {
		mi, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mi.(model)
	}
	mi, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mi.(model)
	if len(m.goals.goals) != 3 || m.goals.goals[2].Text != "Read" {
		t.Errorf("goal not added: %+v", m.goals.goals)
	}
}

func TestToggleGoal(t *testing.T) {
	m := newTestModel(t)
	mi, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mi.(model)
	if !m.goals.goals[0].Done {
		t.Error("space should toggle the first goal done")
	}
}

func TestToggleNonNeg(t *testing.T) {
	m := newTestModel(t)
	// Switch to right panel.
	mi, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mi.(model)
	// Cursor starts at 0 (first non-neg); space toggles it.
	mi, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mi.(model)
	if !m.daily.entry.NonNegs[0] {
		t.Error("space should toggle first non-negotiable")
	}
	// Toggle back.
	mi, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mi.(model)
	if m.daily.entry.NonNegs[0] {
		t.Error("second space should un-toggle non-negotiable")
	}
}

func TestSetMood(t *testing.T) {
	m := newTestModel(t)
	mi, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mi.(model)
	// Navigate down past all non-negs to the mood row.
	for i := 0; i < len(m.daily.nonNegs); i++ {
		mi, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mi.(model)
	}
	if m.daily.cursor != len(m.daily.nonNegs) {
		t.Fatalf("cursor = %d, want mood row %d", m.daily.cursor, len(m.daily.nonNegs))
	}
	// Press '3' to set mood.
	mi, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = mi.(model)
	if m.daily.entry.Mood != 3 {
		t.Errorf("mood = %d, want 3", m.daily.entry.Mood)
	}
	// Press '3' again to clear it (toggle).
	mi, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = mi.(model)
	if m.daily.entry.Mood != 0 {
		t.Errorf("pressing same rating again should clear it, got %d", m.daily.entry.Mood)
	}
}
