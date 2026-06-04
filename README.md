# almanac

A terminal dashboard styled like a two-page journal. The left page shows an
**at-a-glance** view — a small calendar, the weather, and your daily goals — and
the right page is for **writing** free-form notes. A header carries the date, a
live clock, the moon phase, and your step count.

Built with Go and the [Charm](https://charm.sh) stack (BubbleTea, Lipgloss,
Bubbles). Single binary, local-first, no database.

```
almanac  ·  Thu, Jun 4  ·  06:14  ·  🌖 Waning Gibbous  ·  👟 7,432/10k
╭──────────────────────────────╮╭──────────────────────────────────────╮
│  📅 JUNE 2026                 ││  ✎ NOTES  Thu, Jun 4                   │
│  Su Mo Tu We Th Fr Sa         ││  ────────────────────────────────────  │
│      1  2  3 [4] 5  6         ││  Morning pages: ship the almanac.      │
│   7  8  9 10 11 12 13         ││                                        │
│  ...                          ││  ❝ The secret of getting ahead is      │
│  ☀ WEATHER                    ││     getting started.                   │
│  72°F  Partly Cloudy          ││     — Mark Twain                       │
│  H 81° · L 58°                ││                                        │
│  ✺ GOALS                      ││                                        │
│  ☑ Ship plan                  ││                                        │
│  ☐ Walk the dog               ││                                        │
╰──────────────────────────────╯╰────────────────────────────────────────╯
tab notes · ↑↓ move · space toggle · a add · e edit · [ ] day · w refresh · q
```

## Install & run

```sh
go build -o almanac .
./almanac
```

Below ~80 columns the two pages stack vertically.

## Keys

| Key       | Action                                              |
|-----------|-----------------------------------------------------|
| `tab`     | switch focus between Goals and Notes                |
| `↑` `↓`   | move the goal cursor / scroll the notes             |
| `space`   | toggle the selected goal done                       |
| `a`       | add a goal                                          |
| `e`       | edit the selected goal (Goals) / open `$EDITOR` (Notes) |
| `d`       | delete the selected goal                            |
| `i`       | write a note inline (Notes focused; `esc` saves)    |
| `s`       | set today's step count (manual source)              |
| `[` `]`   | previous / next day                                 |
| `w`       | refresh weather & steps                             |
| `q`       | quit                                                |

## Configuration

On first run, a default config is written to:

- macOS: `~/Library/Application Support/almanac/config.toml`
- Linux: `~/.config/almanac/config.toml`

Print the exact path with `almanac --config`.

```toml
accent = "#4ec9b0"

[weather]
api_key = ""          # your OpenWeatherMap API key
city    = "Salt Lake City"
lat     = 40.7608
lon     = -111.8910
units   = "imperial"  # "imperial" (°F) or "metric" (°C)

[steps]
source = "manual"     # "manual", "autoexport", or "appleexport"
path   = ""           # folder (autoexport) or export.xml (appleexport)
goal   = 10000
```

### Weather

Get a free API key at <https://openweathermap.org/api>, put it in
`weather.api_key`, and set your `lat`/`lon`. Until then the weather panel shows a
hint. The last successful fetch is cached, so going offline shows the previous
reading with a `(stale)` marker rather than an error.

### Steps

Three sources, selected by `steps.source`:

- **`manual`** (default) — press `s` to type today's count. Stored locally.
- **`autoexport`** — point `steps.path` at a folder of CSV/JSON files written by
  the [Health Auto Export](https://www.healthexportapp.com/) app. Files are
  scanned and step rows for the selected day are summed.
- **`appleexport`** — point `steps.path` at a native Apple Health `export.xml`.
  It is parsed once and cached (`steps-cache.json`); it is only re-parsed when
  the source file changes, since that file is typically very large.

## Data

Everything lives under the data dir (`~/.local/share/almanac` or
`$XDG_DATA_HOME/almanac`):

- `state.json` — goals, manual step counts, cached weather
- `notes/YYYY-MM-DD.md` — one Markdown file per day, editable outside the app
