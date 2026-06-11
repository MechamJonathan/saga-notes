package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFooter draws contextual key hints and, when set, a right-aligned status message.
func renderFooter(s Styles, m model) string {
	var hints []string

	switch {
	case m.daily.mode == dailyEditNonNeg:
		hints = []string{"enter save", "esc cancel"}
	case m.daily.editing():
		hints = []string{"esc save & exit"}
	case m.goals.editing():
		hints = []string{"enter save", "esc cancel"}
	case m.focus == focusNotes:
		mc := m.daily.maxCur()
		nn := len(m.daily.nonNegs)
		switch {
		case m.daily.cursor == mc:
			hints = []string{"tab goals", "i write", "e $EDITOR", "↑↓ scroll", "[ ] day", "t today", "w refresh", "q quit"}
		case m.daily.cursor == nn || m.daily.cursor == nn+1:
			hints = []string{"tab goals", "1-5 rating", "↑↓ move", "[ ] day", "t today", "q quit"}
		default:
			hints = []string{"tab goals", "space toggle", "a add", "e edit", "d del", "↑↓ move", "[ ] day", "t today", "q quit"}
		}
	default:
		hints = []string{"tab journal", "↑↓ move", "space toggle", "a add", "e edit", "d del", "[ ] day", "t today", "w refresh", "q quit"}
		if m.goals.hasCompleted() {
			hints = append(hints, "c clear done")
		}
	}

	left := s.Footer.Render(strings.Join(hints, " · "))
	if m.statusMsg == "" {
		return left
	}

	right := s.Header.Render(m.statusMsg)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}
