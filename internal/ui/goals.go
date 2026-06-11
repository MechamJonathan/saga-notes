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

// goalsModel is the interactive daily-goals list (largest left-panel section).
type goalsModel struct {
	goals  []storage.Goal
	cursor int
	mode   goalMode
	input  textinput.Model
	styles Styles
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
		if m.cursor < len(m.goals) {
			m.goals[m.cursor].Done = !m.goals[m.cursor].Done
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
			m.mode = goalEditing
			m.input.SetValue(m.goals[m.cursor].Text)
			m.input.CursorEnd()
			m.input.Focus()
			return m, false, "", textinput.Blink
		}
	case "d":
		if len(active) > 0 {
			i := m.cursor
			m.goals = append(m.goals[:i], m.goals[i+1:]...)
			m.clampCursor()
			return m, true, "goal removed", nil
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

func (m goalsModel) updateInput(msg tea.KeyMsg) (goalsModel, bool, string, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		mode := m.mode
		m.mode = goalNormal
		m.input.Blur()
		if text == "" {
			return m, false, "", nil
		}
		var status string
		if mode == goalAdding {
			m.goals = append(m.goals, storage.Goal{Text: text})
			m.cursor = len(m.goals) - 1
			status = "goal added"
		} else if len(m.goals) > 0 {
			m.goals[m.cursor].Text = text
			status = "goal updated"
		}
		return m, true, status, nil
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
	b.WriteString(m.styles.Title.Render("✺ ACTIVE GOALS"))
	b.WriteString("\n")

	if len(active) == 0 && m.mode != goalAdding {
		b.WriteString(m.styles.Faint.Render("  no active goals — press a to add"))
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
			label = m.styles.Faint.Render("☐ " + g.Text)
		}
		b.WriteString(cur + label + "\n")
	}

	if m.mode == goalAdding {
		b.WriteString("  " + m.styles.Faint.Render("☐") + " " + m.input.View() + "\n")
	}

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
