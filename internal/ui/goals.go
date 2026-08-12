package ui

import (
	"strings"

	"saga-notes/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type goalMode int

const (
	goalNormal goalMode = iota
	goalAdding
	goalEditing
)

type deletedGoal struct {
	goal  storage.Goal
	index int
}

// goalsModel is the interactive daily-goals list (largest left-panel section).
type goalsModel struct {
	goals         []storage.Goal
	cursor        int
	mode          goalMode
	input         textinput.Model
	styles        Styles
	confirmDelete bool
	deleted       *deletedGoal // non-nil when undo is available
}

func newGoals(styles Styles, goals []storage.Goal) goalsModel {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.CharLimit = 120
	return goalsModel{goals: goals, input: ti, styles: styles}
}

// editing reports whether the component is currently capturing text input.
func (m goalsModel) editing() bool { return m.mode != goalNormal }

// activeIndices returns the m.goals indices of non-done goals, in order.
func (m goalsModel) activeIndices() []int {
	var idx []int
	for i, g := range m.goals {
		if !g.Done {
			idx = append(idx, i)
		}
	}
	return idx
}

// clampCursor ensures cursor points to an active (non-done) goal. If none
// exist it stays at 0. Always picks the closest active goal at or before
// the current cursor position.
func (m *goalsModel) clampCursor() {
	active := m.activeIndices()
	if len(active) == 0 {
		m.cursor = 0
		return
	}
	best := active[0]
	for _, idx := range active {
		if idx <= m.cursor {
			best = idx
		}
	}
	m.cursor = best
}

// hasCompleted reports whether any goals are marked done.
func (m goalsModel) hasCompleted() bool {
	for _, g := range m.goals {
		if g.Done {
			return true
		}
	}
	return false
}

// update handles a key message when goals is focused. It returns the updated
// model, whether persistent state changed, a status label, and any command.
func (m goalsModel) update(msg tea.KeyMsg) (goalsModel, bool, string, tea.Cmd) {
	if m.mode != goalNormal {
		return m.updateInput(msg)
	}

	// Any key other than 'd' cancels a pending delete confirmation.
	if m.confirmDelete && msg.String() != "d" {
		m.confirmDelete = false
	}

	active := m.activeIndices()

	switch msg.String() {
	case "up", "k":
		for i, idx := range active {
			if idx == m.cursor && i > 0 {
				m.cursor = active[i-1]
				break
			}
		}
	case "down", "j":
		for i, idx := range active {
			if idx == m.cursor && i < len(active)-1 {
				m.cursor = active[i+1]
				break
			}
		}
	case " ":
		if len(active) > 0 {
			m.goals[m.cursor].Done = true
			m.clampCursor()
			return m, true, "", nil
		}
	case "a":
		m.mode = goalAdding
		m.input.SetValue("")
		m.input.Focus()
		return m, false, "", textinput.Blink
	case "e":
		if len(active) > 0 {
			return m, false, "", m.enterEditMode()
		}
	case "d":
		if len(active) == 0 {
			break
		}
		if !m.confirmDelete {
			m.confirmDelete = true
			return m, false, "press d again to delete  ·  esc to cancel", nil
		}
		m.confirmDelete = false
		i := m.cursor
		removed := m.goals[i]
		m.goals = append(m.goals[:i], m.goals[i+1:]...)
		m.clampCursor()
		m.deleted = &deletedGoal{goal: removed, index: i}
		return m, true, "goal removed  ·  u to undo", nil
	case "u":
		if m.deleted != nil {
			idx := m.deleted.index
			if idx > len(m.goals) {
				idx = len(m.goals)
			}
			tail := append([]storage.Goal{m.deleted.goal}, m.goals[idx:]...)
			m.goals = append(m.goals[:idx], tail...)
			m.cursor = idx
			m.deleted = nil
			return m, true, "goal restored", nil
		}
	case "c":
		if m.hasCompleted() {
			var kept []storage.Goal
			for _, g := range m.goals {
				if !g.Done {
					kept = append(kept, g)
				}
			}
			m.goals = kept
			m.clampCursor()
			return m, true, "completed cleared", nil
		}
	}
	return m, false, "", nil
}

// enterEditMode starts editing the currently selected goal. Callers must ensure
// a non-done goal exists at m.cursor before calling.
func (m *goalsModel) enterEditMode() tea.Cmd {
	if len(m.activeIndices()) == 0 {
		return nil
	}
	m.mode = goalEditing
	m.input.SetValue(m.goals[m.cursor].Text)
	m.input.CursorEnd()
	m.input.Focus()
	return textinput.Blink
}

// commitEdit finishes the current add/edit operation, saving the typed text.
// Mirrors pressing enter; returns whether persisted state changed and a
// status label (both zero-valued if the input was empty, a no-op).
func (m *goalsModel) commitEdit() (bool, string) {
	text := strings.TrimSpace(m.input.Value())
	mode := m.mode
	m.mode = goalNormal
	m.input.Blur()
	if text == "" {
		return false, ""
	}
	if mode == goalAdding {
		m.goals = append(m.goals, storage.Goal{Text: text})
		m.cursor = len(m.goals) - 1
		return true, "goal added"
	}
	if len(m.goals) > 0 {
		m.goals[m.cursor].Text = text
		return true, "goal updated"
	}
	return false, ""
}

func (m goalsModel) updateInput(msg tea.KeyMsg) (goalsModel, bool, string, tea.Cmd) {
	switch msg.String() {
	case "enter":
		changed, status := m.commitEdit()
		return m, changed, status, nil
	case "esc":
		m.mode = goalNormal
		m.input.Blur()
		return m, false, "", nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, false, "", cmd
}

func (m goalsModel) view(width int, focused bool) string {
	var b strings.Builder

	// Partition goal indices into active and completed.
	var active, completed []int
	for i, g := range m.goals {
		if g.Done {
			completed = append(completed, i)
		} else {
			active = append(active, i)
		}
	}

	// Active Goals section.
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(m.styles.ContrastFg).
		Background(m.styles.Accent).
		Width(width).
		PaddingLeft(1).
		Render("✺ ACTIVE GOALS"))
	b.WriteString("\n")

	if len(active) == 0 && m.mode != goalAdding {
		b.WriteString(m.styles.Faint.Render("  no active goals"))
		b.WriteString("\n")
	}

	for _, i := range active {
		g := m.goals[i]
		if m.mode == goalEditing && i == m.cursor {
			b.WriteString("  " + m.styles.Faint.Render("☐") + " " + m.input.View() + "\n")
			continue
		}
		cur := "  "
		if focused && i == m.cursor && m.mode == goalNormal {
			cur = m.styles.Selected.Render("› ")
		}
		var label string
		if focused && i == m.cursor {
			label = m.styles.Selected.Render("☐ " + g.Text)
		} else {
			label = m.styles.Normal.Render("☐ " + g.Text)
		}
		b.WriteString(cur + label + "\n")
	}

	if m.mode == goalAdding {
		b.WriteString("  " + m.styles.Faint.Render("☐") + " " + m.input.View() + "\n")
	}

	b.WriteString(m.styles.Faint.Render("  press a to add"))
	b.WriteString("\n")

	// Recently Completed section — display only, not selectable.
	if len(completed) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Faint.Render("  RECENTLY COMPLETED"))
		b.WriteString("\n")
		for _, i := range completed {
			b.WriteString("  " + m.styles.Done.Render("☑ "+m.goals[i].Text) + "\n")
		}
	}

	return lipgloss.NewStyle().Width(width).Render(b.String())
}
