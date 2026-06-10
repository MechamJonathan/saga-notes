package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// stackBreakpoint is the width below which panels stack vertically.
const stackBreakpoint = 80

// Panel chrome. Note how lipgloss sizing works: a style's .Width(n) sets the
// width *including* horizontal padding but *excluding* the border, so the
// rendered block is n+border wide and the usable text area is n-padding wide.
const (
	borderW  = 2 // left + right border columns
	paddingW = 4 // left + right padding columns
	borderH  = 2 // top + bottom border rows
)

// panelOuterWidths returns the total rendered width of each panel and whether
// the layout is stacked.
func (m model) panelOuterWidths() (left, right int, stacked bool) {
	if m.width < stackBreakpoint {
		return m.width, m.width, true
	}
	left = int(float64(m.width) * 0.38)
	return left, m.width - left, false
}

// panelOuterHeights returns the total rendered height of the top/left and
// bottom/right panels. The region excludes the header and footer rows.
func (m model) panelOuterHeights() (top, bottom int, stacked bool) {
	region := m.height - 1 // minus footer
	if m.width < stackBreakpoint {
		t := region / 2
		return t, region - t, true
	}
	return region, region, false
}

// styleWidth converts a desired outer width to the value passed to .Width().
func styleWidth(outer int) int { return outer - borderW }

// contentWidth is the usable text width inside a panel of the given outer width.
func contentWidth(outer int) int { return outer - borderW - paddingW }

// styleHeight converts a desired outer height to the value passed to .Height().
func styleHeight(outer int) int { return outer - borderH }

// layoutDaily resizes the daily-panel viewport/textarea to fit the current window.
func (m *model) layoutDaily() {
	if m.width == 0 || m.height == 0 {
		return
	}
	_, rightOuterW, stacked := m.panelOuterWidths()
	topH, botH, _ := m.panelOuterHeights()
	outerH := topH
	if stacked {
		outerH = botH
	}
	m.daily.resize(contentWidth(rightOuterW), max(1, styleHeight(outerH)))
}

// View renders the full dashboard.
func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading…"
	}

	footer := m.footerLine()

	leftOuterW, rightOuterW, stacked := m.panelOuterWidths()
	topH, botH, _ := m.panelOuterHeights()

	leftContent := m.leftPanel(contentWidth(leftOuterW))
	rightContent := m.daily.view(contentWidth(rightOuterW), m.focus == focusNotes, m.now)

	left := m.panelStyle(focusGoals).
		Width(styleWidth(leftOuterW)).Height(max(1, styleHeight(topH))).
		Render(leftContent)
	right := m.panelStyle(focusNotes).
		Width(styleWidth(rightOuterW)).Height(max(1, styleHeight(botH))).
		Render(rightContent)

	if stacked {
		return left + "\n" + right + "\n" + footer
	}
	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return panels + "\n" + footer
}

// leftPanel composes the calendar, weather, and goals sections.
func (m model) leftPanel(innerW int) string {
	cal := renderCalendar(m.styles, m.selected, m.now, m.selected)
	wx := renderWeather(m.styles, m.weather)
	goals := m.goals.view(innerW, m.focus == focusGoals)
	return strings.Join([]string{cal, wx, goals}, "\n\n")
}

// panelStyle returns the focused or blurred border for the given panel.
func (m model) panelStyle(f focus) lipgloss.Style {
	if m.focus == f {
		return m.styles.PanelFocus
	}
	return m.styles.PanelBlur
}

// footerLine renders the key hints.
func (m model) footerLine() string {
	return renderFooter(m.styles, m)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
