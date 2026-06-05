// Package ui implements the saga-notes terminal dashboard as a BubbleTea program.
package ui

import (
	"context"
	"time"

	"saga-notes/internal/config"
	"saga-notes/internal/storage"
	"saga-notes/internal/weather"

	tea "github.com/charmbracelet/bubbletea"
)

type focus int

const (
	focusGoals focus = iota
	focusNotes
)

// model is the root BubbleTea model.
type model struct {
	cfg    config.Config
	styles Styles
	state  storage.State

	width, height int

	now      time.Time
	selected time.Time // day shown in calendar/notes

	focus focus
	goals goalsModel
	daily dailyModel

	weather weatherState

	statusMsg string
}

// New builds the root model from loaded config and state.
func New(cfg config.Config, state storage.State) model {
	styles := NewStyles(cfg.Accent)
	now := time.Now()
	day := truncDay(now)

	note, _ := storage.LoadNote(day)
	entry, _ := storage.LoadDay(day)

	m := model{
		cfg:      cfg,
		styles:   styles,
		state:    state,
		now:      now,
		selected: day,
		goals:    newGoals(styles, state.Goals),
		daily:    newDaily(styles, cfg.Journal.NonNegotiables, day, entry, note),
	}
	m.weather = weatherState{cache: state.Weather, unit: cfg.TempUnit(), loading: true}
	return m
}

// --- messages ---

type tickMsg time.Time
type weatherMsg struct {
	w   weather.Weather
	err error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchWeatherCmd(m.cfg.Weather),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layoutDaily()
		return m, nil

	case tickMsg:
		m.now = time.Now()
		var cmd tea.Cmd
		if m.now.Second() == 0 && m.now.Minute()%10 == 0 {
			cmd = fetchWeatherCmd(m.cfg.Weather)
		}
		return m, tea.Batch(tickCmd(), cmd)

	case weatherMsg:
		m.weather.loading = false
		if msg.err != nil {
			m.weather.err = msg.err
			return m, nil
		}
		m.weather.err = nil
		cache := storage.WeatherCache{
			TempNow:   msg.w.TempNow,
			TempHigh:  msg.w.TempHigh,
			TempLow:   msg.w.TempLow,
			Desc:      msg.w.Desc,
			Icon:      msg.w.Icon,
			FetchedAt: time.Now(),
		}
		m.weather.cache = &cache
		m.state.Weather = &cache
		_ = storage.Save(m.state)
		return m, nil

	case noteSavedMsg:
		m.statusMsg = "note saved"
		return m, nil

	case editorFinishedMsg:
		body, _ := storage.LoadNote(m.selected)
		dayEntry, _ := storage.LoadDay(m.selected)
		m.daily = m.daily.setDay(m.selected, dayEntry, body)
		m.layoutDaily()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.goals.editing() {
		var cmd tea.Cmd
		var changed bool
		m.goals, changed, cmd = m.goals.update(msg)
		if changed {
			m.persistGoals()
		}
		return m, cmd
	}
	if m.daily.editing() {
		var cmd tea.Cmd
		m.daily, cmd = m.daily.update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		if m.focus == focusGoals {
			m.focus = focusNotes
		} else {
			m.focus = focusGoals
		}
		return m, nil
	case "[":
		return m.changeDay(-1)
	case "]":
		return m.changeDay(1)
	case "w":
		m.weather.loading = m.weather.cache == nil
		return m, fetchWeatherCmd(m.cfg.Weather)
	}

	if m.focus == focusGoals {
		var cmd tea.Cmd
		var changed bool
		m.goals, changed, cmd = m.goals.update(msg)
		if changed {
			m.persistGoals()
		}
		return m, cmd
	}
	var cmd tea.Cmd
	m.daily, cmd = m.daily.update(msg)
	return m, cmd
}

// changeDay moves the selected day and reloads day-scoped data.
func (m model) changeDay(delta int) (tea.Model, tea.Cmd) {
	m.selected = m.selected.AddDate(0, 0, delta)
	body, _ := storage.LoadNote(m.selected)
	dayEntry, _ := storage.LoadDay(m.selected)
	m.daily = m.daily.setDay(m.selected, dayEntry, body)
	m.layoutDaily()
	return m, nil
}

func (m *model) persistGoals() {
	m.state.Goals = m.goals.goals
	_ = storage.Save(m.state)
}

// --- commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func fetchWeatherCmd(cfg config.WeatherConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		w, err := weather.Fetch(ctx, cfg)
		return weatherMsg{w: w, err: err}
	}
}

func truncDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
