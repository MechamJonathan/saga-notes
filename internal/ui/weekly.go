package ui

import (
	"fmt"
	"strings"
	"time"

	"saga-notes/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type weeklyModel struct {
	cursor  int // 0–6, Mon=0 … Sun=6
	anchor  time.Time
	days    [7]storage.DayEntry
	nonNegs []string
	styles  Styles
}

// weekStart returns midnight of the Monday that contains t.
func weekStart(t time.Time) time.Time {
	t = truncDay(t)
	wd := t.Weekday()
	if wd == time.Sunday {
		wd = 7
	}
	return t.AddDate(0, 0, -int(wd-time.Monday))
}

func loadWeekDays(anchor time.Time) [7]storage.DayEntry {
	var days [7]storage.DayEntry
	for i := range 7 {
		days[i], _ = storage.LoadDay(anchor.AddDate(0, 0, i))
	}
	return days
}

func newWeekly(styles Styles, nonNegs []string, today time.Time) weeklyModel {
	anchor := weekStart(today)
	wd := today.Weekday()
	cursor := int(wd) - 1 // Mon=1→0 … Sat=6→5
	if wd == time.Sunday {
		cursor = 6
	}
	return weeklyModel{
		cursor:  cursor,
		anchor:  anchor,
		days:    loadWeekDays(anchor),
		nonNegs: nonNegs,
		styles:  styles,
	}
}

// shiftWeek returns the model shifted by delta weeks (positive = forward).
func (m weeklyModel) shiftWeek(delta int) weeklyModel {
	m.anchor = m.anchor.AddDate(0, 0, delta*7)
	m.days = loadWeekDays(m.anchor)
	return m
}

// update handles key input. Returns the selected day when the user presses enter;
// zero time means no navigation.
func (m weeklyModel) update(msg tea.KeyMsg) (weeklyModel, time.Time) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 6 {
			m.cursor++
		}
	case "enter":
		return m, m.anchor.AddDate(0, 0, m.cursor)
	}
	return m, time.Time{}
}

// ratingDots renders a 5-dot string (● filled, ○ empty) for a 1–5 rating.
// Returns "–" for an unset rating (0).
func ratingDots(rating int) string {
	if rating == 0 {
		return "–"
	}
	var s strings.Builder
	for i := 1; i <= 5; i++ {
		if i <= rating {
			s.WriteString("●")
		} else {
			s.WriteString("○")
		}
	}
	return s.String()
}

// weeklyHabitsBar renders a ▓▓▓░  X/N summary for a single day's entry.
// Returns "–" when no habits are configured.
func weeklyHabitsBar(entry storage.DayEntry, total int) string {
	if total == 0 {
		return "–"
	}
	done := 0
	for i, d := range entry.NonNegs {
		if i < total && d {
			done++
		}
	}
	var bar strings.Builder
	for i := range total {
		if i < done {
			bar.WriteString("▓")
		} else {
			bar.WriteString("░")
		}
	}
	return fmt.Sprintf("%s  %d/%d", bar.String(), done, total)
}

// sparkBlocks maps a 0–5 rating to a single block character.
// 0 (unset) renders as a middle dot so gaps are visible.
var sparkBlocks = []string{"·", "▁", "▂", "▄", "▆", "█"}

// loadTrendDays returns n DayEntry values ending at today, oldest first.
func loadTrendDays(today time.Time, n int) []storage.DayEntry {
	days := make([]storage.DayEntry, n)
	for i := range n {
		day := today.AddDate(0, 0, -(n - 1 - i))
		days[i], _ = storage.LoadDay(day)
	}
	return days
}

// sparkline builds a single-row bar string from a selector applied to each entry.
func sparkline(data []storage.DayEntry, sel func(storage.DayEntry) int) string {
	var sb strings.Builder
	for _, e := range data {
		v := sel(e)
		if v < 0 || v > 5 {
			v = 0
		}
		sb.WriteString(sparkBlocks[v])
	}
	return sb.String()
}

func (m weeklyModel) view(width, _ int, now time.Time) string {
	today := truncDay(now)
	total := len(m.nonNegs)
	var b strings.Builder

	// Week range title
	end := m.anchor.AddDate(0, 0, 6)
	title := fmt.Sprintf(" WEEK OF %s – %s",
		strings.ToUpper(m.anchor.Format("Jan 2")),
		strings.ToUpper(end.Format("Jan 2, 2006")))
	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n\n")

	// Column header aligned to data columns:
	// 2 (cursor) + 10 (day label "Mon Jun  9") + 2 (gap) = 14 chars indent
	// mood col:  14 (5 chars via %-5s)
	// energy col: 14+5+3 = 22 (5 chars)
	// habits col: 22+5+3 = 30
	b.WriteString(strings.Repeat(" ", 14))
	b.WriteString(m.styles.Faint.Render("mood    energy  habits"))
	b.WriteString("\n")

	for i := range 7 {
		day := m.anchor.AddDate(0, 0, i)
		isToday := day.Equal(today)
		isFuture := day.After(today)

		cur := "  "
		if i == m.cursor {
			cur = m.styles.Selected.Render("› ")
		}

		// "Mon Jun  9" — 10 chars (_2 gives space-padded single-digit day)
		dayLabel := day.Format("Mon Jan _2")

		var rowText string
		if isFuture {
			rowText = dayLabel + "  –"
		} else {
			mood := ratingDots(m.days[i].Mood)
			energy := ratingDots(m.days[i].Energy)
			habits := weeklyHabitsBar(m.days[i], total)
			rowText = fmt.Sprintf("%s  %-5s   %-5s   %s", dayLabel, mood, energy, habits)
		}

		var line string
		switch {
		case isToday:
			line = cur + m.styles.Header.Render(rowText)
		case isFuture:
			line = cur + m.styles.Faint.Render(rowText)
		default:
			line = cur + rowText
		}

		b.WriteString(line + "\n")
	}

	// 28-day sparkline trend (oldest left → today right)
	const trendDays = 28
	trend := loadTrendDays(today, trendDays)
	moodLine := sparkline(trend, func(e storage.DayEntry) int { return e.Mood })
	energyLine := sparkline(trend, func(e storage.DayEntry) int { return e.Energy })

	b.WriteString("\n")
	b.WriteString(m.styles.Title.Render(" TREND") + m.styles.Faint.Render(fmt.Sprintf("  (%d days, right = today)", trendDays)))
	b.WriteString("\n")
	b.WriteString(m.styles.Faint.Render(" mood    "))
	b.WriteString(m.styles.Header.Render(moodLine))
	b.WriteString("\n")
	b.WriteString(m.styles.Faint.Render(" energy  "))
	b.WriteString(m.styles.Selected.Render(energyLine))
	b.WriteString("\n")

	return lipgloss.NewStyle().Width(width).Render(b.String())
}
