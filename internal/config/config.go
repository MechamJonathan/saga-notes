// Package config loads and persists the user's almanac configuration from
// ~/.config/almanac/config.toml, creating a sensible default on first run.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the on-disk user configuration.
type Config struct {
	Accent  string        `toml:"accent"` // hex accent color, e.g. "#4ec9b0"
	Weather WeatherConfig `toml:"weather"`
	Steps   StepsConfig   `toml:"steps"`
	Journal JournalConfig `toml:"journal"`
}

// JournalConfig controls the right-panel daily journal page.
type JournalConfig struct {
	// NonNegotiables is the list of daily habit labels shown as checkboxes.
	NonNegotiables []string `toml:"non_negotiables"`
}

// WeatherConfig controls the OpenWeatherMap integration.
type WeatherConfig struct {
	APIKey string  `toml:"api_key"` // OpenWeatherMap API key
	City   string  `toml:"city"`    // display label, e.g. "Salt Lake City"
	Lat    float64 `toml:"lat"`
	Lon    float64 `toml:"lon"`
	Units  string  `toml:"units"` // "imperial" (°F) or "metric" (°C)
}

// StepsConfig selects and configures the step-count source.
type StepsConfig struct {
	Source string `toml:"source"` // "manual", "autoexport", or "appleexport"
	Path   string `toml:"path"`   // file or folder for the export-based sources
	Goal   int    `toml:"goal"`   // daily step goal
}

// Default returns the configuration used on first run.
func Default() Config {
	return Config{
		Accent: "#4ec9b0",
		Weather: WeatherConfig{
			Units: "imperial",
		},
		Steps: StepsConfig{
			Source: "manual",
			Goal:   10000,
		},
		Journal: JournalConfig{
			NonNegotiables: []string{
				"SLEPT 7+ HOURS",
				"READ GOALS",
				"FOLLOWED MEAL PLAN",
				"HYDRATED",
				"STEP GOAL",
			},
		},
	}
}

// Path returns the absolute path to the config file.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "almanac", "config.toml"), nil
}

// Load reads the config file, writing and returning defaults if none exists.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// First run: persist defaults so the user has a file to edit.
		if werr := Save(cfg); werr != nil {
			return cfg, werr
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.applyDefaults()
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// applyDefaults fills in any empty fields a partial config file may have left.
func (c *Config) applyDefaults() {
	d := Default()
	if c.Accent == "" {
		c.Accent = d.Accent
	}
	if c.Weather.Units == "" {
		c.Weather.Units = d.Weather.Units
	}
	if c.Steps.Source == "" {
		c.Steps.Source = d.Steps.Source
	}
	if c.Steps.Goal == 0 {
		c.Steps.Goal = d.Steps.Goal
	}
	if len(c.Journal.NonNegotiables) == 0 {
		c.Journal.NonNegotiables = d.Journal.NonNegotiables
	}
}

// TempUnit returns the display suffix for the configured units.
func (c Config) TempUnit() string {
	if c.Weather.Units == "metric" {
		return "°C"
	}
	return "°F"
}
