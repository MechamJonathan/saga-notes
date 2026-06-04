package ui

import (
	"fmt"
	"strings"
	"time"
)

// renderCalendar draws a compact month grid for the month containing `month`,
// highlighting `selected` if it falls within that month.
func renderCalendar(s Styles, month, selected time.Time) string {
	var b strings.Builder

	heading := strings.ToUpper(month.Format("January 2006"))
	b.WriteString(s.Title.Render("📅 " + heading))
	b.WriteString("\n")
	b.WriteString(s.Faint.Render("Su Mo Tu We Th Fr Sa"))
	b.WriteString("\n")

	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	// Leading blanks for the weekday the 1st lands on (Sunday = 0).
	lead := int(first.Weekday())
	b.WriteString(strings.Repeat("   ", lead))

	daysInMonth := first.AddDate(0, 1, -1).Day()
	col := lead
	for day := 1; day <= daysInMonth; day++ {
		cell := fmt.Sprintf("%2d", day)
		if day == selected.Day() && month.Month() == selected.Month() && month.Year() == selected.Year() {
			cell = s.Today.Render(cell)
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
