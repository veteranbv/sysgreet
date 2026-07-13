// Package terminal detects the capabilities of the terminal sysgreet writes
// to: how many columns are available and how much color it can display. It
// also owns the ANSI palette shared by the ascii and render packages.
package terminal

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// Profile describes the color capability of the output stream.
type Profile int

const (
	// ProfileNoColor means no escape sequences should be emitted.
	ProfileNoColor Profile = iota
	// ProfileANSI means the classic 16-color palette is available.
	ProfileANSI
	// ProfileTrueColor means 24-bit foreground colors are available.
	ProfileTrueColor
)

// Env captures everything the renderers need to know about the output stream.
type Env struct {
	// Width is the usable column count. Zero means unconstrained (for
	// example when output is piped and COLUMNS is unset).
	Width   int
	Profile Profile
}

// DetectEnv inspects the output stream and environment once at startup.
// disableColor forces ProfileNoColor regardless of terminal capability.
func DetectEnv(out *os.File, disableColor bool) Env {
	return Env{
		Width:   Width(out),
		Profile: DetectProfile(out, disableColor),
	}
}

// Width reports the column count of the terminal attached to out, falling
// back to the COLUMNS environment variable. Returns 0 when unknown so
// callers can render at natural width.
func Width(out *os.File) int {
	if out != nil {
		if w, _, err := term.GetSize(int(out.Fd())); err == nil && w > 0 {
			return w
		}
	}
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(strings.TrimSpace(cols)); err == nil && w > 0 {
			return w
		}
	}
	return 0
}

// DetectProfile determines the color capability of out. Precedence:
// explicit disable, then NO_COLOR (https://no-color.org), then TERM=dumb,
// then whether out is a terminal at all, then COLORTERM for truecolor.
func DetectProfile(out *os.File, disableColor bool) Profile {
	if disableColor {
		return ProfileNoColor
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return ProfileNoColor
	}
	if os.Getenv("TERM") == "dumb" {
		return ProfileNoColor
	}
	if out == nil || !term.IsTerminal(int(out.Fd())) {
		return ProfileNoColor
	}
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		return ProfileTrueColor
	}
	return ProfileANSI
}

// Reset is the ANSI attribute reset sequence.
const Reset = "\033[0m"

// ansiCodes maps the color names accepted in configuration to foreground
// escape sequences.
var ansiCodes = map[string]string{
	"red":         "\033[31m",
	"green":       "\033[32m",
	"yellow":      "\033[33m",
	"blue":        "\033[34m",
	"purple":      "\033[35m",
	"cyan":        "\033[36m",
	"gray":        "\033[37m",
	"white":       "\033[97m",
	"brightblue":  "\033[94m",
	"brightcyan":  "\033[96m",
	"brightwhite": "\033[97m",
}

// rgbValues maps the same color names to 24-bit values for truecolor
// gradients. Values follow the common xterm palette.
var rgbValues = map[string][3]uint8{
	"red":         {205, 49, 49},
	"green":       {13, 188, 121},
	"yellow":      {229, 229, 16},
	"blue":        {36, 114, 200},
	"purple":      {188, 63, 188},
	"cyan":        {17, 168, 205},
	"gray":        {229, 229, 229},
	"white":       {255, 255, 255},
	"brightblue":  {59, 142, 234},
	"brightcyan":  {41, 184, 219},
	"brightwhite": {255, 255, 255},
}

// Code returns the ANSI escape sequence for a named color.
func Code(name string) (string, bool) {
	code, ok := ansiCodes[strings.ToLower(name)]
	return code, ok
}

// RGB returns the 24-bit value for a named color.
func RGB(name string) ([3]uint8, bool) {
	rgb, ok := rgbValues[strings.ToLower(name)]
	return rgb, ok
}

// ForegroundRGB builds a 24-bit foreground escape sequence.
func ForegroundRGB(r, g, b uint8) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// Wrap surrounds text with the named color when the profile allows it.
func Wrap(p Profile, color, text string) string {
	if p == ProfileNoColor {
		return text
	}
	code, ok := ansiCodes[strings.ToLower(color)]
	if !ok {
		return text
	}
	return code + text + Reset
}

// RuneWidth returns the terminal column width of r: 2 for the common East
// Asian wide/fullwidth and emoji ranges, 1 otherwise. Zero-width combining
// marks are rare in banner content and treated as width 1.
func RuneWidth(r rune) int {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r >= 0x2E80 && r <= 0xA4CF, // CJK radicals through Yi
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF, // CJK compatibility ideographs
		r >= 0xFE30 && r <= 0xFE4F, // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFF60, // fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji
		r >= 0x20000 && r <= 0x3FFFD: // CJK extensions
		return 2
	}
	return 1
}

// DisplayWidth returns the terminal column count of s, which must not
// contain escape sequences (Strip first if it might).
func DisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}

// Clip truncates s to at most width display columns, marking dropped
// content with an ellipsis. s must not contain escape sequences.
func Clip(s string, width int) string {
	if width <= 0 || DisplayWidth(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	cols := 0
	for i, r := range s {
		rw := RuneWidth(r)
		if cols+rw > width-1 {
			return s[:i] + "…"
		}
		cols += rw
	}
	return s
}

// Strip removes ANSI CSI escape sequences from a string.
func Strip(input string) string {
	var b strings.Builder
	runes := []rune(input)
	i := 0
	for i < len(runes) {
		if runes[i] == '\033' {
			i++
			for i < len(runes) && ((runes[i] >= '0' && runes[i] <= '9') || runes[i] == '[' || runes[i] == ';') {
				i++
			}
			if i < len(runes) {
				i++
			}
			continue
		}
		b.WriteRune(runes[i])
		i++
	}
	return b.String()
}
