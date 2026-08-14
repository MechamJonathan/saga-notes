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
	cursor     int // 0–6, Mon=0 … Sun=6
	anchor     time.Time
	days       [7]storage.DayEntry
	nonNegs    []string
	styles     Styles
	heatMetric int // 0=combined 1=habits 2=mood 3=energy
	monthDays  []storage.DayEntry
	monthStart time.Time
	notes      [7]string

	completedGoals []storage.CompletedGoal
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

func loadWeekNotes(anchor time.Time) [7]string {
	var notes [7]string
	for i := range 7 {
		notes[i], _ = storage.LoadNote(anchor.AddDate(0, 0, i))
	}
	return notes
}

func newWeekly(styles Styles, nonNegs []string, today time.Time) weeklyModel {
	anchor := weekStart(today)
	wd := today.Weekday()
	cursor := int(wd) - 1 // Mon=1→0 … Sat=6→5
	if wd == time.Sunday {
		cursor = 6
	}
	monthDays, monthStart := loadMonthDays(anchor)
	return weeklyModel{
		cursor:     cursor,
		anchor:     anchor,
		days:       loadWeekDays(anchor),
		nonNegs:    nonNegs,
		styles:     styles,
		monthDays:  monthDays,
		monthStart: monthStart,
		notes:      loadWeekNotes(anchor),
	}
}

// shiftWeek returns the model shifted by delta weeks (positive = forward).
func (m weeklyModel) shiftWeek(delta int) weeklyModel {
	m.anchor = m.anchor.AddDate(0, 0, delta*7)
	m.days = loadWeekDays(m.anchor)
	m.monthDays, m.monthStart = loadMonthDays(m.anchor)
	m.notes = loadWeekNotes(m.anchor)
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
	case "m":
		m.heatMetric = (m.heatMetric + 1) % 4
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

var heatMetricLabels = []string{"combined", "habits", "mood", "energy"}

func heatCellStyle(score float64, s Styles) lipgloss.Style {
	switch {
	case score < 0:
		return s.HeatNone
	case score < 0.34:
		return s.HeatLow
	case score < 0.67:
		return s.HeatMid
	default:
		return s.HeatHigh
	}
}

// heatScore returns 0–1 for the given metric, or -1 if no data is available.
func heatScore(e storage.DayEntry, metric, numHabits int) float64 {
	switch metric {
	case 1: // habits
		if numHabits == 0 {
			return -1
		}
		done := 0
		for i, v := range e.NonNegs {
			if i < numHabits && v {
				done++
			}
		}
		return float64(done) / float64(numHabits)
	case 2: // mood
		if e.Mood == 0 {
			return -1
		}
		return float64(e.Mood) / 5.0
	case 3: // energy
		if e.Energy == 0 {
			return -1
		}
		return float64(e.Energy) / 5.0
	default: // combined
		var scores []float64
		if numHabits > 0 {
			done := 0
			for i, v := range e.NonNegs {
				if i < numHabits && v {
					done++
				}
			}
			scores = append(scores, float64(done)/float64(numHabits))
		}
		if e.Mood > 0 {
			scores = append(scores, float64(e.Mood)/5.0)
		}
		if e.Energy > 0 {
			scores = append(scores, float64(e.Energy)/5.0)
		}
		if len(scores) == 0 {
			return -1
		}
		sum := 0.0
		for _, s := range scores {
			sum += s
		}
		return sum / float64(len(scores))
	}
}

// loadMonthDays loads every day in the month containing anchor.
// Returns the entries slice and the first day of that month.
func loadMonthDays(anchor time.Time) ([]storage.DayEntry, time.Time) {
	start := truncDay(time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location()))
	n := start.AddDate(0, 1, -1).Day()
	days := make([]storage.DayEntry, n)
	for i := range n {
		days[i], _ = storage.LoadDay(start.AddDate(0, 0, i))
	}
	return days, start
}

// renderHeatMap builds a GitHub-style monthly calendar grid shaded by daily score.
func (m weeklyModel) renderHeatMap(today time.Time) string {
	var b strings.Builder
	numHabits := len(m.nonNegs)
	n := len(m.monthDays)

	header := fmt.Sprintf(" %s %d  ·  m: %s",
		strings.ToUpper(m.monthStart.Month().String()),
		m.monthStart.Year(),
		heatMetricLabels[m.heatMetric],
	)
	b.WriteString(m.styles.Title.Render(header))
	b.WriteString("\n")
	b.WriteString(m.styles.Faint.Render(" Su Mo Tu We Th Fr Sa"))
	b.WriteString("\n")

	col := int(m.monthStart.Weekday()) // Sunday=0 offset for the first of the month
	var rowCells []string
	for range col {
		rowCells = append(rowCells, "  ")
	}

	for day := 1; day <= n; day++ {
		entry := m.monthDays[day-1]
		date := m.monthStart.AddDate(0, 0, day-1)
		isToday := date.Equal(today)
		isFuture := date.After(today)

		cellText := fmt.Sprintf("%2d", day)
		var styled string
		switch {
		case isToday:
			styled = m.styles.Today.Render(cellText)
		case isFuture:
			styled = m.styles.HeatNone.Render(cellText)
		default:
			styled = heatCellStyle(heatScore(entry, m.heatMetric, numHabits), m.styles).Render(cellText)
		}
		rowCells = append(rowCells, styled)
		col++

		if col == 7 || day == n {
			for col < 7 {
				rowCells = append(rowCells, "  ")
				col++
			}
			b.WriteString(" ")
			b.WriteString(strings.Join(rowCells, " "))
			b.WriteString("\n")
			rowCells = nil
			col = 0
		}
	}

	return b.String()
}

func (m weeklyModel) renderWeeklyStats(today time.Time) string {
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

	b.WriteString(m.renderCompletedGoals())

	return b.String()
}

// renderCompletedGoals renders a dated list of completed goals, newest first.
func (m weeklyModel) renderCompletedGoals() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.styles.Title.Render(" COMPLETED GOALS"))
	b.WriteString("\n")
	if len(m.completedGoals) == 0 {
		b.WriteString(m.styles.Faint.Render("  no completed goals yet"))
		b.WriteString("\n")
		return b.String()
	}
	for i := len(m.completedGoals) - 1; i >= 0; i-- {
		g := m.completedGoals[i]
		date := m.styles.Faint.Render(g.CompletedAt.Format("Jan _2"))
		b.WriteString(fmt.Sprintf("  %s  %s\n", date, g.Text))
	}
	return b.String()
}

func (m weeklyModel) renderNotePreview(width int) string {
	const maxLines = 8

	var b strings.Builder
	day := m.anchor.AddDate(0, 0, m.cursor)
	b.WriteString(m.styles.Title.Render(strings.ToUpper(day.Format("Mon, Jan 2"))))
	b.WriteString("\n")

	note := strings.TrimSpace(m.notes[m.cursor])
	if note == "" {
		b.WriteString(m.styles.Faint.Render("  (no note)"))
		b.WriteString("\n")
		return b.String()
	}

	wrapped := lipgloss.NewStyle().Width(width).Render(note)
	lines := strings.Split(wrapped, "\n")
	truncated := len(lines) > maxLines
	if truncated {
		lines = lines[:maxLines]
	}
	b.WriteString(strings.Join(lines, "\n"))
	if truncated {
		b.WriteString("\n")
		b.WriteString(m.styles.Faint.Render("  ..."))
	}
	b.WriteString("\n")
	return b.String()
}

func (m weeklyModel) view(width, _ int, now time.Time) string {
	today := truncDay(now)
	left := lipgloss.NewStyle().PaddingRight(6).Render(m.renderWeeklyStats(today))
	right := lipgloss.JoinVertical(lipgloss.Left,
		m.renderNotePreview(width/3),
		m.renderHeatMap(today),
	)
	joined := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.Place(width, lipgloss.Height(joined), lipgloss.Center, lipgloss.Top, joined)
}
