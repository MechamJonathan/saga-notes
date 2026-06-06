// Package weather fetches current conditions from OpenWeatherMap.
package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"saga-notes/internal/config"
)

// Weather is a normalized snapshot of current conditions.
type Weather struct {
	City     string
	TempNow  float64
	TempHigh float64
	TempLow  float64
	Desc     string
	Icon     string // emoji glyph
	Pop      int    // probability of precipitation, %
}

// ErrNoAPIKey indicates the user has not configured an OpenWeatherMap key.
var ErrNoAPIKey = errors.New("weather: no OpenWeatherMap API key configured")

// ErrNoLocation indicates the user has not configured lat/lon coordinates.
var ErrNoLocation = errors.New("weather: lat/lon not set in config.toml")

// owmResponse is the subset of the /data/2.5/weather payload we use.
type owmResponse struct {
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp    float64 `json:"temp"`
		TempMin float64 `json:"temp_min"`
		TempMax float64 `json:"temp_max"`
	} `json:"main"`
	Name string `json:"name"`
}

// Fetch retrieves current conditions for the configured location.
func Fetch(ctx context.Context, cfg config.WeatherConfig) (Weather, error) {
	if cfg.APIKey == "" {
		return Weather{}, ErrNoAPIKey
	}
	if cfg.Lat == 0 && cfg.Lon == 0 {
		return Weather{}, ErrNoLocation
	}

	q := url.Values{}
	q.Set("appid", cfg.APIKey)
	q.Set("units", cfg.Units)
	q.Set("lat", strconv.FormatFloat(cfg.Lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(cfg.Lon, 'f', -1, 64))

	endpoint := "https://api.openweathermap.org/data/2.5/weather?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Weather{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Weather{}, fmt.Errorf("weather: OpenWeatherMap returned %s", resp.Status)
	}

	var body owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Weather{}, err
	}

	w := Weather{
		City:     cfg.City,
		TempNow:  body.Main.Temp,
		TempHigh: body.Main.TempMax,
		TempLow:  body.Main.TempMin,
	}
	if w.City == "" {
		w.City = body.Name
	}
	if len(body.Weather) > 0 {
		w.Desc = body.Weather[0].Description
		w.Icon = iconGlyph(body.Weather[0].Icon)
	}
	return w, nil
}

// iconGlyph maps an OpenWeatherMap icon code to an emoji.
func iconGlyph(code string) string {
	if len(code) < 2 {
		return "•"
	}
	switch code[:2] {
	case "01":
		return "☀"
	case "02":
		return "🌤"
	case "03":
		return "⛅"
	case "04":
		return "☁"
	case "09":
		return "🌧"
	case "10":
		return "🌦"
	case "11":
		return "⛈"
	case "13":
		return "❄"
	case "50":
		return "🌫"
	default:
		return "•"
	}
}
