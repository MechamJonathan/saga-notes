package ui

import "github.com/charmbracelet/lipgloss"

// palette holds the two accent shades that define a theme.
type palette struct {
	accent    lipgloss.Color // bright accent — titles, borders, selected items
	accentDim lipgloss.Color // darker accent — today cell background, heat map low tier
}

// themePalettes maps theme names to their color palettes.
var themePalettes = map[string]palette{
	"teal":   {accent: "#2de2d2", accentDim: "#1a8a84"},
	"amber":  {accent: "#fbbf24", accentDim: "#92400e"},
	"nordic": {accent: "#88c0d0", accentDim: "#4c566a"},
}

// themeOrder defines the cycle order for the in-app T keybinding.
var themeOrder = []string{"teal", "amber", "nordic"}

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
	Normal      lipgloss.Style // primary body text (white)
	Selected    lipgloss.Style // selected list row
	Today       lipgloss.Style // today's calendar cell
	Done        lipgloss.Style // completed goal text
	ProgressOn  lipgloss.Style // filled progress segment
	ProgressOff lipgloss.Style // empty progress segment

	// Heat map tiers for the monthly calendar in the weekly view.
	HeatNone lipgloss.Style // no data
	HeatLow  lipgloss.Style // 0–33%
	HeatMid  lipgloss.Style // 34–66%
	HeatHigh lipgloss.Style // 67–100%
}

// NewStyles builds the style set for the named theme ("teal", "amber", …).
// Unknown names fall back to teal.
func NewStyles(theme string) Styles {
	p, ok := themePalettes[theme]
	if !ok {
		p = themePalettes["teal"]
	}
	accent := p.accent
	accentDim := p.accentDim
	dim := lipgloss.Color("240")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)

	return Styles{
		Accent: accent,
		Dim:    dim,

		App:        lipgloss.NewStyle(),
		PanelFocus: panel.BorderForeground(accent),
		PanelBlur:  panel.BorderForeground(dim),

		Title:  lipgloss.NewStyle().Foreground(accent).Bold(true),
		Header: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Footer: lipgloss.NewStyle().Foreground(dim),
		Faint:  lipgloss.NewStyle().Foreground(dim),
		Normal: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),

		Selected: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Today: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(accentDim).
			Bold(true),
		Done: lipgloss.NewStyle().Foreground(dim).Strikethrough(true),

		ProgressOn:  lipgloss.NewStyle().Foreground(accent),
		ProgressOff: lipgloss.NewStyle().Foreground(dim),

		HeatNone: lipgloss.NewStyle().Foreground(dim),
		HeatLow:  lipgloss.NewStyle().Foreground(accentDim),
		HeatMid:  lipgloss.NewStyle().Foreground(accent),
		HeatHigh: lipgloss.NewStyle().Foreground(accent).Bold(true),
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
