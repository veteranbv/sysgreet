package ascii

import "github.com/veteranbv/sysgreet/internal/terminal"

// RenderOptions describes how ASCII art should be generated.
type RenderOptions struct {
	Font       string
	Color      string
	Gradient   []string
	Monochrome bool
	Uppercase  bool
	// MaxWidth caps the art width in columns. Zero means unconstrained.
	MaxWidth int
	// ShortenDomain allows the width ladder to drop everything after the
	// first dot (pve1.home.lan -> pve1). Only hostname rendering sets
	// this; user-supplied --text must never be rewritten.
	ShortenDomain bool
	// Profile gates color output for the target terminal.
	Profile terminal.Profile
}

// Art is the result of rendering a headline, along with what the width
// ladder had to do to make it fit.
type Art struct {
	Text  string // final output, colored per Profile
	Font  string // font used; "plain" when the ladder hit bottom
	Color string
	Width int // widest line in columns, ignoring escape sequences
	// Shortened reports that the domain was dropped from the name to fit
	// the terminal, so callers can surface the full name elsewhere.
	Shortened bool
}
