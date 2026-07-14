package render

import (
	"github.com/veteranbv/sysgreet/internal/terminal"
)

// Colorizer wraps text in ANSI sequences when the terminal profile allows.
type Colorizer struct {
	profile terminal.Profile
}

// NewColorizer creates a colorizer for the given terminal profile.
func NewColorizer(profile terminal.Profile) Colorizer {
	return Colorizer{profile: profile}
}

// Wrap applies color when the profile supports it.
func (c Colorizer) Wrap(color, text string) string {
	return terminal.Wrap(c.profile, color, text)
}

// Strip removes ANSI escape sequences from a string.
func Strip(input string) string {
	return terminal.Strip(input)
}
