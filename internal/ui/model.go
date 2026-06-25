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
	focusWeek
)

// model is the root BubbleTea model.
type model struct {
	cfg    config.Config
	styles Styles
	state  storage.State

	width, height int

	now      time.Time
	selected time.Time // day shown in calendar/notes

	focus  focus
	goals  goalsModel
	daily  dailyModel
	weekly weeklyModel

	weather weatherState

	statusMsg string
	showHelp  bool
}

// New builds the root model from loaded config and state.
func New(cfg config.Config, state storage.State) model {
	styles := NewStyles(cfg.Accent)
	now := time.Now()
	day := truncDay(now)

	note, _ := storage.LoadNote(day)
	entry, _ := storage.LoadDay(day)

	nonNegs := state.NonNegotiables
	if len(nonNegs) == 0 {
		nonNegs = cfg.Journal.NonNegotiables
	}
	streaks := storage.ComputeNonNegStreaks(len(nonNegs), truncDay(now))

	m := model{
		cfg:      cfg,
		styles:   styles,
		state:    state,
		now:      now,
		selected: day,
		goals:    newGoals(styles, state.Goals),
		daily:    newDaily(styles, nonNegs, streaks, day, entry, note),
		weekly:   newWeekly(styles, nonNegs, now),
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
type statusClearMsg struct{}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return statusClearMsg{} })
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
			City:      msg.w.City,
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

	case nonNegsSavedMsg:
		m.state.NonNegotiables = msg.labels
		_ = storage.Save(m.state)
		m.layoutDaily()
		m.refreshStreaks()
		m.weekly.nonNegs = msg.labels
		m.statusMsg = "habits saved"
		return m, statusClearCmd()

	case noteSavedMsg:
		m.statusMsg = "note saved"
		return m, statusClearCmd()

	case editorFinishedMsg:
		body, _ := storage.LoadNote(m.selected)
		dayEntry, _ := storage.LoadDay(m.selected)
		m.daily = m.daily.setDay(m.selected, dayEntry, body)
		m.layoutDaily()
		m.statusMsg = "note saved"
		return m, statusClearCmd()

	case statusClearMsg:
		m.statusMsg = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay intercepts all keys except its own dismiss keys.
	if m.showHelp {
		if msg.String() == "?" || msg.String() == "esc" {
			m.showHelp = false
		}
		return m, nil
	}

	if msg.String() == "?" {
		m.showHelp = true
		return m, nil
	}

	if m.goals.editing() {
		var cmd tea.Cmd
		var changed bool
		var status string
		m.goals, changed, status, cmd = m.goals.update(msg)
		if changed {
			m.persistGoals()
		}
		if status != "" {
			m.statusMsg = status
			cmd = tea.Batch(cmd, statusClearCmd())
		}
		return m, cmd
	}
	if m.daily.editing() {
		var cmd tea.Cmd
		m.daily, cmd = m.daily.update(msg)
		m.layoutDaily()
		m.refreshStreaks()
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		m.focus = (m.focus + 1) % 3
		return m, nil
	case "[":
		if m.focus == focusWeek {
			m.weekly = m.weekly.shiftWeek(-1)
			return m, nil
		}
		return m.changeDay(-1)
	case "]":
		if m.focus == focusWeek {
			m.weekly = m.weekly.shiftWeek(1)
			return m, nil
		}
		return m.changeDay(1)
	case "t":
		return m.jumpToday()
	case "w":
		m.weather.loading = m.weather.cache == nil
		return m, fetchWeatherCmd(m.cfg.Weather)
	}

	if m.focus == focusWeek {
		var day time.Time
		m.weekly, day = m.weekly.update(msg)
		if !day.IsZero() {
			m.selected = day
			m.focus = focusNotes
			body, _ := storage.LoadNote(m.selected)
			entry, _ := storage.LoadDay(m.selected)
			m.daily = m.daily.setDay(m.selected, entry, body)
			m.layoutDaily()
		}
		return m, nil
	}
	if m.focus == focusGoals {
		var cmd tea.Cmd
		var changed bool
		var status string
		m.goals, changed, status, cmd = m.goals.update(msg)
		if changed {
			m.persistGoals()
		}
		if status != "" {
			m.statusMsg = status
			cmd = tea.Batch(cmd, statusClearCmd())
		}
		return m, cmd
	}
	var cmd tea.Cmd
	m.daily, cmd = m.daily.update(msg)
	m.layoutDaily()
	m.refreshStreaks()
	return m, cmd
}

// jumpToday resets the selected day to today. In weekly mode it also resets
// the displayed week to the current week.
func (m model) jumpToday() (tea.Model, tea.Cmd) {
	m.selected = truncDay(m.now)
	body, _ := storage.LoadNote(m.selected)
	dayEntry, _ := storage.LoadDay(m.selected)
	m.daily = m.daily.setDay(m.selected, dayEntry, body)
	m.layoutDaily()
	if m.focus == focusWeek {
		m.weekly = newWeekly(m.styles, m.daily.nonNegs, m.now)
	}
	return m, nil
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

func (m *model) persistNonNegs() {
	m.state.NonNegotiables = m.daily.nonNegs
	_ = storage.Save(m.state)
}

func (m *model) refreshStreaks() {
	m.daily.streaks = storage.ComputeNonNegStreaks(len(m.daily.nonNegs), truncDay(m.now))
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
