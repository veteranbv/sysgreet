package banner

import "strings"

// Level grades how much attention an item deserves; renderers map it to
// color.
type Level int

const (
	LevelNormal Level = iota
	LevelWarn
	LevelCrit
)

// String names the level for JSON output.
func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "warn"
	case LevelCrit:
		return "crit"
	default:
		return "normal"
	}
}

// Item is one labeled fact in a section, e.g. Label "Mem", Value "23%",
// Detail "3.7/16.0 GiB" with a meter at 0.23.
type Item struct {
	Label  string
	Value  string
	Detail string
	// Meter, when set, is the fill fraction of a usage bar. Values above 1
	// mean over capacity (load above core count) and render as a full bar.
	Meter *float64
	Level Level
}

// Text renders the item as a single plain line, used for compact output
// and the JSON "lines" field.
func (it Item) Text() string {
	var b strings.Builder
	b.WriteString(it.Label)
	b.WriteString(": ")
	b.WriteString(it.Value)
	if it.Detail != "" {
		b.WriteString(" ")
		b.WriteString(it.Detail)
	}
	return b.String()
}

func meter(fraction float64) *float64 {
	return &fraction
}

// levelFor grades a usage fraction against the warn/crit thresholds shared
// by every meter.
func levelFor(fraction float64) Level {
	switch {
	case fraction >= 0.90:
		return LevelCrit
	case fraction >= 0.75:
		return LevelWarn
	default:
		return LevelNormal
	}
}

func newSection(key, title string, items []Item, data map[string]any) (Section, bool) {
	if len(items) == 0 {
		return Section{}, false
	}
	lines := make([]string, len(items))
	for i, it := range items {
		lines[i] = it.Text()
	}
	return Section{Key: key, Title: title, Items: items, Lines: lines, Data: data}, true
}
