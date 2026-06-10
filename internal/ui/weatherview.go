package ui

import (
	"fmt"
	"strings"
	"time"

	"saga-notes/internal/storage"
)

// weatherState bundles everything the weather block needs to render.
type weatherState struct {
	cache   *storage.WeatherCache
	unit    string // "°F" / "°C"
	loading bool
	err     error
}

// renderWeather draws the compact left-panel weather block.
func renderWeather(s Styles, w weatherState) string {
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
	return b.String()
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
