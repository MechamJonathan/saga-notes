package ui

import (
	"fmt"
	"strings"
	"time"
)

// renderCalendar draws a compact month grid for the month containing `month`.
// today is highlighted with the Today style; selected (if different) with Selected.
func renderCalendar(s Styles, month, today, selected time.Time) string {
	var b strings.Builder

	heading := strings.ToUpper(month.Format("January 2006"))
	b.WriteString(s.Title.Render("📅 " + heading))
	b.WriteString("\n")
	b.WriteString(s.Faint.Render("Su Mo Tu We Th Fr Sa"))
	b.WriteString("\n")

	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	lead := int(first.Weekday())
	b.WriteString(strings.Repeat("   ", lead))

	sameDay := func(a, b time.Time) bool {
		return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
	}

	daysInMonth := first.AddDate(0, 1, -1).Day()
	col := lead
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(month.Year(), month.Month(), d, 0, 0, 0, 0, month.Location())
		cell := fmt.Sprintf("%2d", d)
		switch {
		case sameDay(date, selected):
			cell = s.Today.Render(cell)
		case sameDay(date, today):
			cell = s.Selected.Render(cell)
		}
		b.WriteString(cell)
		col++
		if col%7 == 0 {
			b.WriteString("\n")
		} else {
			b.WriteString(" ")
		}
	}
	return b.String()
}
