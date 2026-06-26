// Command saga is a terminal dashboard styled like a two-page journal: a
// left "info" page (calendar, weather, goals) and a right "writing" page
// (notes), with a header showing the date, clock, and moon phase.
package main

import (
	"flag"
	"fmt"
	"os"

	"saga-notes/internal/config"
	"saga-notes/internal/storage"
	"saga-notes/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	showConfig := flag.Bool("config", false, "print the config file path and exit")
	flag.Usage = usage
	flag.Parse()

	cfgPath, _ := config.Path()
	if *showConfig {
		fmt.Println(cfgPath)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "saga: loading config: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "saga: %v\n\nEdit your config: %s\n", err, cfgPath)
		os.Exit(1)
	}

	state, err := storage.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "saga: loading state: %v\n", err)
		os.Exit(1)
	}

	// Request the terminal window to maximize before entering alt screen.
	fmt.Print("\033[9;1t")

	p := tea.NewProgram(ui.New(cfg, state), tea.WithAltScreen())
	_, runErr := p.Run()
	fmt.Print("\033[9;0t") // restore window to pre-launch size
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "saga: %v\n", runErr)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `saga — a terminal journal dashboard

Usage:
  saga [flags]

Flags:
  --config   print the config file path and exit
  -h, --help show this help

Keys:
  tab        switch between Goals and Notes
  ↑/↓        move (goals) or scroll (notes)
  space      toggle a goal
  a/e/d      add / edit / delete a goal
  a/e/d      add / edit / delete a non-negotiable (notes focused, on a non-neg)
  i / e      write a note inline / open $EDITOR (notes focused, on notes)
  [ / ]      previous / next day
  t          jump to today
  w          refresh weather
  q          quit
`)
}
