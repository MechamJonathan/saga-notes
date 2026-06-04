// Command almanac is a terminal dashboard styled like a two-page journal: a
// left "info" page (calendar, weather, goals) and a right "writing" page
// (notes), with a header showing the date, clock, moon phase, and step count.
package main

import (
	"flag"
	"fmt"
	"os"

	"almanac/internal/config"
	"almanac/internal/storage"
	"almanac/internal/ui"

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
		fmt.Fprintf(os.Stderr, "almanac: loading config: %v\n", err)
		os.Exit(1)
	}

	state, err := storage.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "almanac: loading state: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(cfg, state), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "almanac: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `almanac — a terminal journal dashboard

Usage:
  almanac [flags]

Flags:
  --config   print the config file path and exit
  -h, --help show this help

Keys:
  tab        switch between Goals and Notes
  ↑/↓        move (goals) or scroll (notes)
  space      toggle a goal
  a/e/d      add / edit / delete a goal
  i / e      write a note inline / open $EDITOR (notes focused)
  s          set today's step count (manual source)
  [ / ]      previous / next day
  w          refresh weather & steps
  q          quit
`)
}
