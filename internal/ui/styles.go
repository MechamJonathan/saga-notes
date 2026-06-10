package ui

import "github.com/charmbracelet/lipgloss"

// Single-hue palette — everything derives from one phosphor teal.
const (
	teal    = lipgloss.Color("#2de2d2") // primary accent
	tealDim = lipgloss.Color("#1a8a84") // darker teal — focused border, today bg
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

// NewStyles builds the single-teal style set.
func NewStyles(_ string) Styles {
	dim := lipgloss.Color("240")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)

	return Styles{
		Accent: teal,
		Dim:    dim,

		App:        lipgloss.NewStyle(),
		PanelFocus: panel.BorderForeground(teal),
		PanelBlur:  panel.BorderForeground(dim),

		Title:  lipgloss.NewStyle().Foreground(teal).Bold(true),
		Header: lipgloss.NewStyle().Foreground(teal).Bold(true),
		Footer: lipgloss.NewStyle().Foreground(dim),
		Faint:  lipgloss.NewStyle().Foreground(dim),

		Selected: lipgloss.NewStyle().Foreground(teal).Bold(true),
		Today: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(tealDim).
			Bold(true),
		Done: lipgloss.NewStyle().Foreground(dim).Strikethrough(true),

		ProgressOn:  lipgloss.NewStyle().Foreground(teal),
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
