package ui

import (
	"fmt"
	"strings"
	"time"

	"saga-notes/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type noteEntry struct {
	day     time.Time
	content string
}

type searchResult struct {
	day     time.Time
	excerpt string
}

type searchModel struct {
	input   textinput.Model
	notes   []noteEntry    // all notes, loaded when search opens
	results []searchResult // filtered by current query
	cursor  int
	styles  Styles
}

func newSearch(styles Styles) searchModel {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "search notes…"
	ti.CharLimit = 100
	return searchModel{input: ti, styles: styles}
}

// open loads all notes from disk and resets state. Called when "/" is pressed.
func (m searchModel) open() (searchModel, tea.Cmd) {
	days, _ := storage.ListNotes()
	m.notes = make([]noteEntry, 0, len(days))
	for _, d := range days {
		content, _ := storage.LoadNote(d)
		if strings.TrimSpace(content) != "" {
			m.notes = append(m.notes, noteEntry{day: d, content: content})
		}
	}
	m.input.SetValue("")
	m.cursor = 0
	m.results = m.filter("")
	return m, m.input.Focus()
}

func (m searchModel) filter(query string) []searchResult {
	if query == "" {
		out := make([]searchResult, len(m.notes))
		for i, n := range m.notes {
			out[len(m.notes)-1-i] = searchResult{day: n.day, excerpt: firstNonBlankLine(n.content)}
		}
		return out
	}
	q := strings.ToLower(query)
	var out []searchResult
	for i := len(m.notes) - 1; i >= 0; i-- {
		n := m.notes[i]
		if strings.Contains(strings.ToLower(n.content), q) {
			out = append(out, searchResult{day: n.day, excerpt: matchExcerpt(n.content, query)})
		}
	}
	return out
}

// update handles keys while search is open.
// Returns (model, selectedDay, cmd); selectedDay non-zero means "navigate here and close".
func (m searchModel) update(msg tea.KeyMsg) (searchModel, time.Time, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.results) > 0 {
			return m, m.results[m.cursor].day, nil
		}
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.results = m.filter(m.input.Value())
		m.cursor = 0
		return m, time.Time{}, cmd
	}
	return m, time.Time{}, nil
}

const maxSearchResults = 8

func (m searchModel) view(width, height int) string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("SEARCH NOTES") + "\n")
	b.WriteString(m.input.View() + "\n\n")

	if len(m.results) == 0 {
		b.WriteString(m.styles.Faint.Render("  no results"))
	} else {
		limit := min(maxSearchResults, len(m.results))
		for i, r := range m.results[:limit] {
			date := r.day.Format("Jan _2, 2006")
			excerpt := r.excerpt
			maxExcerpt := max(10, width-35)
			if len([]rune(excerpt)) > maxExcerpt {
				excerpt = string([]rune(excerpt)[:maxExcerpt]) + "…"
			}
			if i == m.cursor {
				b.WriteString(m.styles.Selected.Render(fmt.Sprintf("  %s  %s", date, excerpt)))
			} else {
				b.WriteString(fmt.Sprintf("  %s  %s",
					m.styles.Faint.Render(date),
					m.styles.Normal.Render(excerpt),
				))
			}
			b.WriteString("\n")
		}
		if len(m.results) > maxSearchResults {
			b.WriteString(m.styles.Faint.Render(fmt.Sprintf("  … %d more", len(m.results)-maxSearchResults)))
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Accent).
		Padding(1, 3).
		Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func firstNonBlankLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

func matchExcerpt(content, query string) string {
	lower := strings.ToLower(content)
	q := strings.ToLower(query)
	idx := strings.Index(lower, q)
	if idx < 0 {
		return firstNonBlankLine(content)
	}
	start := max(0, idx-20)
	end := min(len(content), idx+len(query)+40)
	excerpt := strings.ReplaceAll(content[start:end], "\n", " ")
	if start > 0 {
		excerpt = "…" + excerpt
	}
	if end < len(content) {
		excerpt = excerpt + "…"
	}
	return excerpt
}
