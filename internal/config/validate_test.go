package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Default()
}

func TestValidateDefault(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Errorf("Default() should be valid, got: %v", err)
	}
}

func TestValidateAccentHex(t *testing.T) {
	cases := []struct {
		accent  string
		wantErr bool
	}{
		{"#4ec9b0", false},  // lowercase 6-digit
		{"#4EC9B0", false},  // uppercase 6-digit
		{"#abc", false},     // 3-digit shorthand
		{"#ABC", false},     // 3-digit uppercase
		{"", true},          // empty (applyDefaults fills it, but test Validate directly)
		{"4ec9b0", true},    // missing #
		{"#gggggg", true},   // invalid hex digits
		{"#4ec9b", true},    // 5 digits
		{"purple", true},    // named colour — not accepted
		{"rgb(0,0,0)", true},
	}
	for _, tc := range cases {
		cfg := validConfig()
		cfg.Accent = tc.accent
		err := cfg.Validate()
		if tc.wantErr && err == nil {
			t.Errorf("accent %q: expected error, got nil", tc.accent)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("accent %q: unexpected error: %v", tc.accent, err)
		}
	}
}

func TestValidateUnits(t *testing.T) {
	cfg := validConfig()
	cfg.Weather.Units = "imperial"
	if err := cfg.Validate(); err != nil {
		t.Errorf("units=imperial should be valid: %v", err)
	}

	cfg.Weather.Units = "metric"
	if err := cfg.Validate(); err != nil {
		t.Errorf("units=metric should be valid: %v", err)
	}

	cfg.Weather.Units = "celsius"
	if err := cfg.Validate(); err == nil {
		t.Error("units=celsius should be invalid")
	}

	cfg.Weather.Units = ""
	if err := cfg.Validate(); err == nil {
		t.Error("empty units should be invalid (applyDefaults fills it; validate catches it if skipped)")
	}
}

func TestValidateWeatherAPIKeyWithoutCoords(t *testing.T) {
	cfg := validConfig()
	cfg.Weather.APIKey = "somekey"
	cfg.Weather.Lat = 0
	cfg.Weather.Lon = 0
	err := cfg.Validate()
	if err == nil {
		t.Error("api_key set but lat/lon=0 should produce an error")
	}
	if !strings.Contains(err.Error(), "lat and lon") {
		t.Errorf("error should mention lat/lon, got: %v", err)
	}
}

func TestValidateWeatherCoordsWithoutAPIKey(t *testing.T) {
	cfg := validConfig()
	cfg.Weather.APIKey = ""
	cfg.Weather.Lat = 40.7
	cfg.Weather.Lon = -111.9
	err := cfg.Validate()
	if err == nil {
		t.Error("lat/lon set but empty api_key should produce an error")
	}
	if !strings.Contains(err.Error(), "api_key") {
		t.Errorf("error should mention api_key, got: %v", err)
	}
}

func TestValidateWeatherFullyConfigured(t *testing.T) {
	cfg := validConfig()
	cfg.Weather.APIKey = "somekey"
	cfg.Weather.Lat = 40.7608
	cfg.Weather.Lon = -111.8910
	cfg.Weather.City = "Salt Lake City"
	if err := cfg.Validate(); err != nil {
		t.Errorf("fully configured weather should be valid: %v", err)
	}
}

func TestValidateWeatherNotConfigured(t *testing.T) {
	cfg := validConfig()
	// All zero/empty — valid: weather just won't load
	cfg.Weather.APIKey = ""
	cfg.Weather.Lat = 0
	cfg.Weather.Lon = 0
	if err := cfg.Validate(); err != nil {
		t.Errorf("unconfigured weather should be valid: %v", err)
	}
}

func TestValidateMultipleIssues(t *testing.T) {
	cfg := validConfig()
	cfg.Accent = "bad"
	cfg.Weather.Units = "kelvin"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for multiple issues")
	}
	if !strings.Contains(err.Error(), "2 issue(s)") {
		t.Errorf("error should report count, got: %v", err)
	}
}
