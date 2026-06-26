package config

import (
	"fmt"
	"regexp"
	"strings"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)

// Validate checks for invalid or inconsistent values and returns a single error
// listing every problem found. Returns nil when the config is valid.
//
// Call after Load() so that applyDefaults has already filled in empty fields —
// this catches values that are present but wrong, not merely absent.
func (c Config) Validate() error {
	var issues []string

	if !hexColorRe.MatchString(c.Accent) {
		issues = append(issues, fmt.Sprintf(
			"accent %q is not a valid hex colour — use #RRGGBB or #RGB (e.g. #4ec9b0)",
			c.Accent,
		))
	}

	if c.Weather.Units != "imperial" && c.Weather.Units != "metric" {
		issues = append(issues, fmt.Sprintf(
			"weather.units %q must be \"imperial\" or \"metric\"",
			c.Weather.Units,
		))
	}

	if c.Weather.APIKey != "" && c.Weather.Lat == 0 && c.Weather.Lon == 0 {
		issues = append(issues, "weather.api_key is set but lat and lon are both 0 — add coordinates to enable weather")
	}

	if c.Weather.APIKey == "" && (c.Weather.Lat != 0 || c.Weather.Lon != 0) {
		issues = append(issues, "weather.lat/lon are set but api_key is missing — add an OpenWeatherMap API key")
	}

	if len(issues) == 0 {
		return nil
	}
	return fmt.Errorf("config.toml has %d issue(s):\n  • %s",
		len(issues), strings.Join(issues, "\n  • "))
}
