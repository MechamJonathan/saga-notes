package ui

import (
	"fmt"
	"strings"
	"time"

	"saga-notes/internal/astro"
	"saga-notes/internal/storage"
	"saga-notes/internal/weather"
)

// weatherState bundles everything the weather block needs to render.
type weatherState struct {
	cache    *storage.WeatherCache
	forecast []weather.ForecastDay
	unit     string  // "°F" / "°C"
	lat, lon float64 // for sunrise/sunset calculation
	loading  bool
	err      error
}

// renderWeather draws the compact left-panel weather block.
func renderWeather(s Styles, w weatherState, now time.Time) string {
	var b strings.Builder
	b.WriteString(s.Title.Render("☀ WEATHER"))
	b.WriteString("\n")

	switch {
	case w.cache == nil && w.loading:
		b.WriteString(s.Faint.Render("loading…"))
		return b.String()
	case w.cache == nil && w.err != nil:
		b.WriteString(s.Faint.Render(weatherErrHint(w.err)))
		return b.String()
	case w.cache == nil:
		b.WriteString(s.Faint.Render("unavailable"))
		return b.String()
	}

	c := w.cache
	if c.City != "" {
		b.WriteString(s.Faint.Render(c.City))
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("%s  %.0f%s  %s\n", c.Icon, c.TempNow, w.unit, c.Desc))
	b.WriteString(s.Faint.Render(fmt.Sprintf("H %.0f%s · L %.0f%s", c.TempHigh, w.unit, c.TempLow, w.unit)))

	if stale := staleLabel(c.FetchedAt); stale != "" {
		b.WriteString(s.Faint.Render("  " + stale))
	}

	if rise, set := astro.SunTimes(now, w.lat, w.lon); !rise.IsZero() {
		b.WriteString("\n")
		b.WriteString(s.Faint.Render(fmt.Sprintf("↑ %s  ↓ %s", rise.Format("3:04 PM"), set.Format("3:04 PM"))))
	}
	return b.String()
}

// renderForecast draws 4 compact forecast rows below the current-weather block.
// Returns an empty string when no forecast data is available.
func renderForecast(s Styles, days []weather.ForecastDay, units string) string {
	if len(days) == 0 {
		return ""
	}
	unit := "°C"
	if units == "imperial" {
		unit = "°F"
	}
	var lines []string
	for _, d := range days {
		pop := ""
		if d.Pop > 20 {
			pop = fmt.Sprintf("  %d%%", d.Pop)
		}
		row := fmt.Sprintf("%-3s  %s  H %.0f%s  L %.0f%s%s",
			d.Date.Format("Mon"), d.Icon, d.High, unit, d.Low, unit, pop)
		lines = append(lines, s.Faint.Render(row))
	}
	return strings.Join(lines, "\n")
}

// weatherErrHint turns a fetch error into a short user-facing hint.
func weatherErrHint(err error) string {
	if strings.Contains(err.Error(), "no OpenWeatherMap API key") {
		return "set weather.api_key in config.toml"
	}
	if strings.Contains(err.Error(), "lat/lon not set") {
		return "set weather.lat and weather.lon in config.toml"
	}
	return "offline"
}

// staleLabel marks cached data older than 30 minutes.
func staleLabel(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	age := time.Since(at)
	if age < 30*time.Minute {
		return ""
	}
	return fmt.Sprintf("(stale %dm)", int(age.Minutes()))
}
