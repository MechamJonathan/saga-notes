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

// ForecastDay is a single day's aggregated forecast.
type ForecastDay struct {
	Date time.Time
	High float64
	Low  float64
	Icon string // emoji via iconGlyph
	Pop  int    // max probability of precipitation, %
}

// owmForecastResponse is the subset of the /data/2.5/forecast payload we use.
type owmForecastResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			TempMax float64 `json:"temp_max"`
			TempMin float64 `json:"temp_min"`
		} `json:"main"`
		Weather []struct {
			Icon string `json:"icon"`
		} `json:"weather"`
		Pop float64 `json:"pop"`
	} `json:"list"`
}

// FetchForecast retrieves a 4-day forecast (tomorrow through 4 days out).
func FetchForecast(ctx context.Context, cfg config.WeatherConfig) ([]ForecastDay, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}
	if cfg.Lat == 0 && cfg.Lon == 0 {
		return nil, ErrNoLocation
	}

	q := url.Values{}
	q.Set("appid", cfg.APIKey)
	q.Set("units", cfg.Units)
	q.Set("lat", strconv.FormatFloat(cfg.Lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(cfg.Lon, 'f', -1, 64))

	endpoint := "https://api.openweathermap.org/data/2.5/forecast?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather: OpenWeatherMap returned %s", resp.Status)
	}

	var body owmForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	today := forecastTruncDay(time.Now().Local())

	type dayAgg struct {
		high   float64
		low    float64
		icon   string
		maxPop float64
	}

	var order []time.Time
	agg := map[time.Time]*dayAgg{}

	for _, slot := range body.List {
		date := forecastTruncDay(time.Unix(slot.Dt, 0).Local())
		if date.Equal(today) {
			continue
		}
		d, ok := agg[date]
		if !ok {
			d = &dayAgg{high: slot.Main.TempMax, low: slot.Main.TempMin}
			agg[date] = d
			order = append(order, date)
		}
		if slot.Main.TempMax > d.high {
			d.high = slot.Main.TempMax
		}
		if slot.Main.TempMin < d.low {
			d.low = slot.Main.TempMin
		}
		if slot.Pop > d.maxPop {
			d.maxPop = slot.Pop
			if len(slot.Weather) > 0 {
				d.icon = slot.Weather[0].Icon
			}
		}
	}

	var result []ForecastDay
	for _, date := range order {
		if len(result) == 4 {
			break
		}
		d := agg[date]
		icon := "•"
		if d.icon != "" {
			icon = iconGlyph(d.icon)
		}
		result = append(result, ForecastDay{
			Date: date,
			High: d.high,
			Low:  d.low,
			Icon: icon,
			Pop:  int(d.maxPop * 100),
		})
	}
	return result, nil
}

func forecastTruncDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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
