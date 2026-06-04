// Package quotes serves a deterministic quote-of-the-day from an embedded list.
package quotes

import (
	_ "embed"
	"strings"
	"time"
)

//go:embed quotes.txt
var raw string

// Quote is a single quotation with its attribution.
type Quote struct {
	Text   string
	Author string
}

// all is the parsed quote list, populated once at init.
var all []Quote

func init() {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: "quote text | Author"
		text, author, _ := strings.Cut(line, "|")
		all = append(all, Quote{
			Text:   strings.TrimSpace(text),
			Author: strings.TrimSpace(author),
		})
	}
}

// OfDay returns the quote for the given day. The selection is deterministic:
// the same day always yields the same quote, and it rotates daily.
func OfDay(day time.Time) Quote {
	if len(all) == 0 {
		return Quote{}
	}
	// Days since the Unix epoch gives a stable, monotonically-increasing index.
	idx := int(day.Truncate(24*time.Hour).Unix()/86400) % len(all)
	if idx < 0 {
		idx += len(all)
	}
	return all[idx]
}
