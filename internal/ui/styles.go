package ui

import "github.com/charmbracelet/lipgloss"

// Retro Nordic palette — phosphor-bright aurora tones on a cool Nordic base.
const (
	nordTeal     = lipgloss.Color("#2de2d2") // phosphor teal   — accent, done items
	nordCyan     = lipgloss.Color("#5bc8fa") // arctic cyan     — header, selection
	nordBlue     = lipgloss.Color("#80aaff") // aurora blue     — section titles
	nordDeepBlue = lipgloss.Color("#4077d4") // vivid cobalt    — focused border, today
)

// Styles holds the lipgloss styles for the UI.
type Styles struct {
	Accent lipgloss.Color
	Dim    lipgloss.Color

	App         lipgloss.Style // outer frame
	PanelFocus  lipgloss.Style // a focused panel border
	PanelBlur   lipgloss.Style // an unfocused panel border
	Title       lipgloss.Style // section headings ("GOALS", "NOTES")
	Header      lipgloss.Style // top header bar text
	Footer      lipgloss.Style // bottom key-hint bar
	Faint       lipgloss.Style // secondary/dim text
	Selected    lipgloss.Style // selected list row
	Today       lipgloss.Style // today's calendar cell
	Done        lipgloss.Style // completed goal text
	ProgressOn  lipgloss.Style // filled progress segment
	ProgressOff lipgloss.Style // empty progress segment
}

// NewStyles builds the Nord-frost style set.
func NewStyles(_ string) Styles {
	dim := lipgloss.Color("240")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)

	return Styles{
		Accent: nordTeal,
		Dim:    dim,

		App:        lipgloss.NewStyle(),
		PanelFocus: panel.BorderForeground(nordDeepBlue),
		PanelBlur:  panel.BorderForeground(dim),

		Title:  lipgloss.NewStyle().Foreground(nordBlue).Bold(true),
		Header: lipgloss.NewStyle().Foreground(nordCyan).Bold(true),
		Footer: lipgloss.NewStyle().Foreground(dim),
		Faint:  lipgloss.NewStyle().Foreground(dim),

		Selected: lipgloss.NewStyle().Foreground(nordCyan).Bold(true),
		Today: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(nordDeepBlue).
			Bold(true),
		Done: lipgloss.NewStyle().Foreground(dim).Strikethrough(true),

		ProgressOn:  lipgloss.NewStyle().Foreground(nordTeal),
		ProgressOff: lipgloss.NewStyle().Foreground(dim),
	}
}

// progressBar renders a fixed-width filled/empty bar for a 0..1 fraction.
func (s Styles) progressBar(pct float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(pct*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	on, off := "", ""
	for i := 0; i < filled; i++ {
		on += "▓"
	}
	for i := 0; i < width-filled; i++ {
		off += "░"
	}
	return s.ProgressOn.Render(on) + s.ProgressOff.Render(off)
}
