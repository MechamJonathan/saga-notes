package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"almanac/internal/quotes"
	"almanac/internal/storage"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dailyMode int

const (
	dailyNormal   dailyMode = iota
	dailyEditNote           // notes textarea is active
)

// editorFinishedMsg is sent when an external $EDITOR session exits.
type editorFinishedMsg struct{ err error }

// noteSavedMsg signals a note write completed.
type noteSavedMsg struct {
	day  time.Time
	body string
}

// dailyModel is the right-panel structured daily journal page.
// Cursor positions:
//
//	[0, len(nonNegs)-1] → non-negotiable checkboxes
//	len(nonNegs)        → mood rating row
//	len(nonNegs)+1      → energy rating row
//	len(nonNegs)+2      → notes section
type dailyModel struct {
	day     time.Time
	entry   storage.DayEntry
	nonNegs []string // labels from config
	note    string   // loaded from .md file

	cursor int
	mode   dailyMode

	viewport viewport.Model
	textarea textarea.Model

	width, height int
	styles        Styles
}

func newDaily(styles Styles, nonNegs []string, day time.Time, entry storage.DayEntry, note string) dailyModel {
	entry = entry.EnsureNonNegs(len(nonNegs))

	ta := textarea.New()
	ta.Placeholder = "Write about your day…"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0

	m := dailyModel{
		day:      day,
		entry:    entry,
		nonNegs:  nonNegs,
		note:     note,
		mode:     dailyNormal,
		textarea: ta,
		viewport: viewport.New(0, 0),
		styles:   styles,
	}
	m.refreshViewport()
	return m
}

func (m dailyModel) maxCur() int { return len(m.nonNegs) + 2 }

func (m dailyModel) editing() bool { return m.mode == dailyEditNote }

func (m dailyModel) setDay(day time.Time, entry storage.DayEntry, note string) dailyModel {
	entry = entry.EnsureNonNegs(len(m.nonNegs))
	m.day = day
	m.entry = entry
	m.note = note
	m.mode = dailyNormal
	m.textarea.Blur()
	m.refreshViewport()
	return m
}

func (m *dailyModel) resize(width, height int) {
	m.width = width
	m.height = height
	vpH := max(2, height-m.overheadLines())
	m.viewport.Width = width
	m.viewport.Height = vpH
	m.textarea.SetWidth(width)
	m.textarea.SetHeight(vpH)
	m.refreshViewport()
}

// overheadLines is the number of fixed rows above + below the notes viewport.
//
//	day-sel(1) sep(1) blank(1) hdr(1) items(N) blank(1)
//	mood(1) energy(1) blank(1) NOTES-hdr(1) dot-above(1)
//	[viewport]
//	blank(1) dot-below(1) blank(1) quote(1) nl(1) author(1)
//	= 14 + N
func (m dailyModel) overheadLines() int { return 14 + len(m.nonNegs) }

func (m *dailyModel) refreshViewport() {
	content := m.note
	if strings.TrimSpace(content) == "" {
		content = m.styles.Faint.Render("no note yet — press i to write, e for $EDITOR")
	}
	m.viewport.SetContent(content)
}

func (m dailyModel) update(msg tea.KeyMsg) (dailyModel, tea.Cmd) {
	if m.mode == dailyEditNote {
		return m.updateTextarea(msg)
	}
	return m.updateNormal(msg)
}

func (m dailyModel) updateNormal(msg tea.KeyMsg) (dailyModel, tea.Cmd) {
	mc := m.maxCur()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < mc {
			m.cursor++
		}
	case " ":
		if m.cursor < len(m.nonNegs) {
			m.entry.NonNegs[m.cursor] = !m.entry.NonNegs[m.cursor]
			_ = storage.SaveDay(m.day, m.entry)
		}
	case "1", "2", "3", "4", "5":
		n, _ := strconv.Atoi(msg.String())
		switch m.cursor {
		case len(m.nonNegs): // mood
			if m.entry.Mood == n {
				m.entry.Mood = 0
			} else {
				m.entry.Mood = n
			}
			_ = storage.SaveDay(m.day, m.entry)
		case len(m.nonNegs) + 1: // energy
			if m.entry.Energy == n {
				m.entry.Energy = 0
			} else {
				m.entry.Energy = n
			}
			_ = storage.SaveDay(m.day, m.entry)
		}
	case "i", "enter":
		if m.cursor == mc {
			m.mode = dailyEditNote
			m.textarea.SetValue(m.note)
			m.textarea.CursorEnd()
			m.textarea.Focus()
			return m, textarea.Blink
		}
	case "e":
		if m.cursor == mc {
			return m, m.openEditor()
		}
	}
	// Pass scroll events to the viewport when on the notes section.
	if m.cursor == mc {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m dailyModel) updateTextarea(msg tea.KeyMsg) (dailyModel, tea.Cmd) {
	if msg.String() == "esc" {
		m.note = m.textarea.Value()
		m.mode = dailyNormal
		m.textarea.Blur()
		m.refreshViewport()
		return m, saveNoteCmd(m.day, m.note)
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m dailyModel) openEditor() tea.Cmd {
	path, err := storage.NotePath(m.day)
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{err} }
	}
	body := m.note
	if strings.TrimSpace(body) == "" {
		body = "\n"
	}
	if err := storage.SaveNote(m.day, body); err != nil {
		return func() tea.Msg { return editorFinishedMsg{err} }
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg { return editorFinishedMsg{err} })
}

func saveNoteCmd(day time.Time, body string) tea.Cmd {
	return func() tea.Msg {
		_ = storage.SaveNote(day, strings.TrimRight(body, "\n")+"\n")
		return noteSavedMsg{day: day, body: body}
	}
}

// --- rendering ---------------------------------------------------------------

func (m dailyModel) view(width int, focused bool) string {
	var b strings.Builder

	// Day-of-week selector + date label
	daySel := m.renderDaySel()
	dateLabel := m.styles.Faint.Render(m.day.Format("Mon, Jan 2"))
	gap := max(1, width-lipgloss.Width(daySel)-lipgloss.Width(dateLabel))
	b.WriteString(daySel + strings.Repeat(" ", gap) + dateLabel)
	b.WriteString("\n")
	b.WriteString(m.styles.Faint.Render(strings.Repeat("─", max(1, width))))
	b.WriteString("\n")
	b.WriteString("\n")

	// Non-negotiables
	b.WriteString(m.sectionHdr("NON-NEGOTIABLES", width))
	b.WriteString("\n")
	for i, label := range m.nonNegs {
		done := i < len(m.entry.NonNegs) && m.entry.NonNegs[i]
		b.WriteString(m.renderNonNeg(i, label, done, focused))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Mood + Energy ratings
	b.WriteString(m.renderRating("MOOD  ", m.entry.Mood, len(m.nonNegs), focused))
	b.WriteString("\n")
	b.WriteString(m.renderRating("ENERGY", m.entry.Energy, len(m.nonNegs)+1, focused))
	b.WriteString("\n")
	b.WriteString("\n")

	// Notes section with dot-grid aesthetic
	b.WriteString(m.sectionHdr("NOTES", width))
	b.WriteString("\n")
	dots := m.styles.Faint.Render(dotLine(width))
	b.WriteString(dots)
	b.WriteString("\n")
	if m.mode == dailyEditNote {
		b.WriteString(m.textarea.View())
	} else {
		b.WriteString(m.viewport.View())
	}
	b.WriteString("\n")
	b.WriteString(dots)
	b.WriteString("\n")

	// Daily quote at the bottom
	q := quotes.OfDay(m.day)
	if q.Text != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Quote.Render("❝ " + q.Text))
		b.WriteString("\n")
		b.WriteString(m.styles.Quote.Render("   — " + q.Author))
	}

	return lipgloss.NewStyle().Width(width).Render(b.String())
}

// sectionHdr renders a full-width bold header bar in the journal's style:
// black text on accent-coloured background.
func (m dailyModel) sectionHdr(title string, width int) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(m.styles.Accent).
		Width(width).
		PaddingLeft(1).
		Render(title)
}

func (m dailyModel) renderDaySel() string {
	labels := []string{"S", "M", "T", "W", "T", "F", "S"}
	wd := int(m.day.Weekday())
	var parts []string
	for i, d := range labels {
		if i == wd {
			parts = append(parts, m.styles.Today.Render("["+d+"]"))
		} else {
			parts = append(parts, m.styles.Faint.Render(" "+d+" "))
		}
	}
	return strings.Join(parts, "")
}

func (m dailyModel) renderNonNeg(i int, label string, done bool, focused bool) string {
	cur := "  "
	if focused && m.cursor == i && m.mode == dailyNormal {
		cur = m.styles.Selected.Render("› ")
	}

	check := m.styles.Faint.Render("○")
	if done {
		check = lipgloss.NewStyle().Foreground(m.styles.Accent).Render("◉")
	}

	var text string
	switch {
	case focused && m.cursor == i:
		text = m.styles.Selected.Render(label)
	case done:
		text = label // full brightness when accomplished
	default:
		text = m.styles.Faint.Render(label)
	}
	return cur + check + "  " + text
}

func (m dailyModel) renderRating(label string, val, cursorPos int, focused bool) string {
	cur := "  "
	if focused && m.cursor == cursorPos && m.mode == dailyNormal {
		cur = m.styles.Selected.Render("› ")
	}

	var lbl string
	if focused && m.cursor == cursorPos {
		lbl = m.styles.Selected.Render(label)
	} else {
		lbl = m.styles.Faint.Render(label)
	}

	var nums []string
	for i := 1; i <= 5; i++ {
		n := fmt.Sprintf("%d", i)
		if val == i {
			nums = append(nums, lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true).Render("●"+n))
		} else {
			nums = append(nums, m.styles.Faint.Render("○"+n))
		}
	}
	return cur + lbl + "  " + strings.Join(nums, "  ")
}

// dotLine builds a row of evenly-spaced dots spanning width columns.
func dotLine(width int) string {
	var b strings.Builder
	for b.Len()+1 <= width {
		if b.Len() > 0 {
			b.WriteString("  ")
		}
		b.WriteString("·")
	}
	return b.String()
}
