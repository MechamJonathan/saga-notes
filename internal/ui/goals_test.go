package ui

import (
	"testing"

	"saga-notes/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

func testGoals(goals ...storage.Goal) goalsModel {
	return newGoals(NewStyles("#2de2d2"), goals)
}

func TestGoalNavDownSkipsCompleted(t *testing.T) {
	// layout: active(0), done(1), active(2)
	// down from 0 should land on 2, skipping the done goal at 1
	m := testGoals(
		storage.Goal{Text: "a"},
		storage.Goal{Text: "b", Done: true},
		storage.Goal{Text: "c"},
	)
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyDown})
	if m2.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (skipped done goal at index 1)", m2.cursor)
	}
}

func TestGoalNavUpSkipsCompleted(t *testing.T) {
	m := testGoals(
		storage.Goal{Text: "a"},
		storage.Goal{Text: "b", Done: true},
		storage.Goal{Text: "c"},
	)
	m.cursor = 2
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyUp})
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (skipped done goal at index 1)", m2.cursor)
	}
}

func TestGoalNavAtBoundaries(t *testing.T) {
	m := testGoals(storage.Goal{Text: "only"})
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyUp})
	if m2.cursor != 0 {
		t.Error("up at first goal should stay at 0")
	}
	m3, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyDown})
	if m3.cursor != 0 {
		t.Error("down at last goal should stay at 0")
	}
}

func TestGoalClearCompleted(t *testing.T) {
	m := testGoals(
		storage.Goal{Text: "active"},
		storage.Goal{Text: "done1", Done: true},
		storage.Goal{Text: "done2", Done: true},
	)
	m2, changed, status, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if !changed {
		t.Error("c should mark state as changed")
	}
	if status != "completed cleared" {
		t.Errorf("status = %q, want %q", status, "completed cleared")
	}
	if len(m2.goals) != 1 || m2.goals[0].Text != "active" {
		t.Errorf("goals after clear = %+v, want only the active goal", m2.goals)
	}
}

func TestGoalClearNoOp(t *testing.T) {
	m := testGoals(storage.Goal{Text: "active"})
	_, changed, _, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if changed {
		t.Error("c with no completed goals should not change state")
	}
}

func TestGoalToggleClampsToActive(t *testing.T) {
	// Toggle the only active goal done — cursor should remain at 0 (no active goals left)
	m := testGoals(storage.Goal{Text: "only"})
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeySpace})
	if !m2.goals[0].Done {
		t.Error("space should mark goal as done")
	}
	if m2.cursor != 0 {
		t.Errorf("cursor = %d after toggle, want 0", m2.cursor)
	}
}

func TestGoalToggleCompletedCannotBeSelected(t *testing.T) {
	// After toggling, cursor must always point to an active goal
	m := testGoals(
		storage.Goal{Text: "a"},
		storage.Goal{Text: "b"},
		storage.Goal{Text: "c"},
	)
	m.cursor = 1
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeySpace})
	// goal[1] is now done; cursor should move to goal[0] (closest active at or before 1)
	if m2.goals[1].Done == false {
		t.Error("space should mark goal[1] as done")
	}
	if m2.goals[m2.cursor].Done {
		t.Errorf("cursor landed on a done goal at index %d", m2.cursor)
	}
}

func TestGoalDelete(t *testing.T) {
	m := testGoals(
		storage.Goal{Text: "first"},
		storage.Goal{Text: "second"},
	)
	m2, changed, status, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !changed {
		t.Error("d should mark state as changed")
	}
	if status != "goal removed" {
		t.Errorf("status = %q, want %q", status, "goal removed")
	}
	if len(m2.goals) != 1 || m2.goals[0].Text != "second" {
		t.Errorf("after delete: %+v", m2.goals)
	}
}

func TestGoalDeleteDoesNotActOnCompleted(t *testing.T) {
	// All goals done — d should be a no-op
	m := testGoals(storage.Goal{Text: "done", Done: true})
	_, changed, _, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if changed {
		t.Error("d should not act when there are no active goals")
	}
}

func TestGoalAddFlow(t *testing.T) {
	m := testGoals()
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m2.editing() {
		t.Fatal("a should enter adding mode")
	}
	for _, r := range "New goal" {
		m2, _, _, _ = m2.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m3, changed, status, _ := m2.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !changed {
		t.Error("enter should mark state changed")
	}
	if status != "goal added" {
		t.Errorf("status = %q, want %q", status, "goal added")
	}
	if len(m3.goals) != 1 || m3.goals[0].Text != "New goal" {
		t.Errorf("goals = %+v, want [{Text:New goal}]", m3.goals)
	}
}

func TestGoalEditFlow(t *testing.T) {
	m := testGoals(storage.Goal{Text: "original"})
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !m2.editing() {
		t.Fatal("e should enter editing mode")
	}
	// Clear input and type new text
	m2.input.SetValue("updated")
	m3, changed, status, _ := m2.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !changed {
		t.Error("enter should mark state changed")
	}
	if status != "goal updated" {
		t.Errorf("status = %q, want %q", status, "goal updated")
	}
	if m3.goals[0].Text != "updated" {
		t.Errorf("goal text = %q, want %q", m3.goals[0].Text, "updated")
	}
}

func TestGoalEscCancelsEdit(t *testing.T) {
	m := testGoals(storage.Goal{Text: "original"})
	m2, _, _, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m3, changed, _, _ := m2.update(tea.KeyMsg{Type: tea.KeyEsc})
	if changed {
		t.Error("esc should not mark state changed")
	}
	if m3.editing() {
		t.Error("esc should exit editing mode")
	}
	if m3.goals[0].Text != "original" {
		t.Error("esc should not modify goal text")
	}
}
