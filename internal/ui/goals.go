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

// update handles a key message when goals is focused. It returns the updated
// model, whether persistent state changed, and any command to run.
func (m goalsModel) update(msg tea.KeyMsg) (goalsModel, bool, tea.Cmd) {
	if m.mode != goalNormal {
		return m.updateInput(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.goals)-1 {
			m.cursor++
		}
	case " ":
		if len(m.goals) > 0 {
			m.goals[m.cursor].Done = !m.goals[m.cursor].Done
			return m, true, nil
		}
	case "a":
		m.mode = goalAdding
		m.input.SetValue("")
		m.input.Focus()
		return m, false, textinput.Blink
	case "e":
		if len(m.goals) > 0 {
			m.mode = goalEditing
			m.input.SetValue(m.goals[m.cursor].Text)
			m.input.CursorEnd()
			m.input.Focus()
			return m, false, textinput.Blink
		}
	case "d":
		if len(m.goals) > 0 {
			m.goals = append(m.goals[:m.cursor], m.goals[m.cursor+1:]...)
			if m.cursor >= len(m.goals) && m.cursor > 0 {
				m.cursor--
			}
			return m, true, nil
		}
	}
	return m, false, nil
}

func (m goalsModel) updateInput(msg tea.KeyMsg) (goalsModel, bool, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		mode := m.mode
		m.mode = goalNormal
		m.input.Blur()
		if text == "" {
			return m, false, nil
		}
		if mode == goalAdding {
			m.goals = append(m.goals, storage.Goal{Text: text})
			m.cursor = len(m.goals) - 1
		} else if len(m.goals) > 0 {
			m.goals[m.cursor].Text = text
		}
		return m, true, nil
	case "esc":
		m.mode = goalNormal
		m.input.Blur()
		return m, false, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, false, cmd
}

func (m goalsModel) view(width int, focused bool) string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("✺ GOALS"))
	b.WriteString("\n")

	if len(m.goals) == 0 && m.mode != goalAdding {
		b.WriteString(m.styles.Faint.Render("no goals yet — press a to add"))
	}

	for i, g := range m.goals {
		check := "☐"
		if g.Done {
			check = "☑"
		}
		text := g.Text
		if m.mode == goalEditing && i == m.cursor {
			b.WriteString("  " + check + " " + m.input.View() + "\n")
			continue
		}
		label := check + " " + text
		switch {
		case g.Done:
			label = m.styles.Faint.Render(check) + " " + m.styles.Done.Render(text)
		case focused && i == m.cursor:
			label = m.styles.Selected.Render(label)
		}
		cursor := "  "
		if focused && i == m.cursor && m.mode == goalNormal {
			cursor = m.styles.Selected.Render("› ")
		}
		b.WriteString(cursor + label + "\n")
	}

	if m.mode == goalAdding {
		b.WriteString("  ☐ " + m.input.View() + "\n")
	}

	return lipgloss.NewStyle().Width(width).Render(b.String())
}
