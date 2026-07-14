package ascii

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/veteranbv/sysgreet/assets"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

const (
	resetColor = "reset"

	// defaultFont is used when a configured font does not exist. A typo in
	// the config should degrade predictably, not roll a random font on
	// every login.
	defaultFont = "ANSI Regular"

	// plainFont is reported when the width ladder falls all the way back
	// to a single-line header.
	plainFont = "plain"
)

// narrowFonts are tried in order (narrowest last) when the configured font
// is too wide for the terminal.
var narrowFonts = []string{"standard", "Small"}

var supportedColors = []string{"red", "green", "yellow", "blue", "purple", "cyan", "gray", "white"}

var debugEnabled = os.Getenv("SYSGREET_DEBUG") != ""

func debugf(format string, args ...any) {
	if !debugEnabled {
		return
	}
	log.Printf("sysgreet ascii: "+format, args...)
}

// Renderer produces ASCII art banners from embedded FIGlet fonts.
type Renderer struct {
	fonts map[string][]byte
	order []string
	rnd   *rand.Rand
}

// NewRenderer loads embedded fonts into memory.
func NewRenderer() (*Renderer, error) {
	paths, err := fs.Glob(assets.FontsFS, "fonts/*.flf")
	if err != nil {
		return nil, err
	}
	fonts := make(map[string][]byte)
	var order []string
	for _, p := range paths {
		data, err := assets.FontsFS.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("load font %s: %w", p, err)
		}
		name := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		fonts[name] = data
		order = append(order, name)
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("no fonts embedded")
	}
	return &Renderer{
		fonts: fonts,
		order: order,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // G404: Non-security use (visual randomization only)
	}, nil
}

// Fonts returns the list of available fonts.
func (r *Renderer) Fonts() []string {
	return append([]string{}, r.order...)
}

// Render creates ASCII art for the input string. When opts.MaxWidth is set
// and the art would overflow, the renderer walks a fallback ladder — drop
// the domain from the name, try narrower fonts, and finally emit a plain
// single-line header — so output never wraps mid-glyph.
func (r *Renderer) Render(text string, opts RenderOptions) (Art, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "sysgreet"
	}
	if opts.Uppercase {
		text = strings.ToUpper(text)
	}

	// FIGlet fonts cover printable ASCII; rendering anything else would
	// silently drop characters, so such text goes straight to the plain
	// header, which preserves it verbatim.
	if !fontRenderable(text) {
		return r.plainHeader(text, text, opts), nil
	}

	texts := []string{text}
	if opts.ShortenDomain {
		if short := shortName(text); short != text {
			texts = append(texts, short)
		}
	}

	for _, font := range r.fontLadder(opts.Font) {
		for _, candidate := range texts {
			rows := r.rows(candidate, font)
			width := maxRowWidth(rows)
			if width == 0 {
				// FIGlet fonts only cover ASCII; text that renders to
				// nothing (e.g. CJK) falls through to the plain header.
				continue
			}
			if opts.MaxWidth > 0 && width > opts.MaxWidth {
				continue
			}
			art, color := r.colorize(rows, opts)
			return Art{
				Text:      art,
				Font:      font,
				Color:     color,
				Width:     width,
				Shortened: candidate != text,
			}, nil
		}
	}

	return r.plainHeader(texts[len(texts)-1], text, opts), nil
}

// fontLadder returns the fonts to try, widest first: the requested font,
// then the built-in narrow fallbacks.
func (r *Renderer) fontLadder(requested string) []string {
	ladder := []string{r.resolveFont(requested)}
	for _, name := range narrowFonts {
		if _, ok := r.fonts[name]; !ok {
			continue
		}
		if name == ladder[0] {
			continue
		}
		ladder = append(ladder, name)
	}
	return ladder
}

// resolveFont maps a configured font name to an embedded font. Unknown
// names fall back to the default deterministically.
func (r *Renderer) resolveFont(name string) string {
	if name == "" || name == "random" {
		return r.randomFont()
	}
	if _, ok := r.fonts[name]; ok {
		return name
	}
	debugf("unknown font %q, falling back to %q", name, defaultFont)
	if _, ok := r.fonts[defaultFont]; ok {
		return defaultFont
	}
	return r.order[0]
}

// rows renders text with a font in non-strict mode: characters the font
// does not cover are dropped rather than aborting the process (go-figure's
// strict mode calls log.Fatal on non-ASCII input).
func (r *Renderer) rows(text, font string) []string {
	fig := figure.NewFigureWithFont(text, bytes.NewReader(r.fonts[font]), false)
	return fig.Slicify()
}

func maxRowWidth(rows []string) int {
	widest := 0
	for _, row := range rows {
		if n := terminal.DisplayWidth(row); n > widest {
			widest = n
		}
	}
	return widest
}

// fontRenderable reports whether every rune falls in the printable ASCII
// range the embedded FIGlet fonts cover.
func fontRenderable(text string) bool {
	for _, r := range text {
		if r < ' ' || r > '~' {
			return false
		}
	}
	return true
}

// shortName drops everything after the first dot, so pve1.home.lan renders
// as pve1. The full name stays available for an info line.
func shortName(text string) string {
	if idx := strings.IndexByte(text, '.'); idx > 0 {
		return text[:idx]
	}
	return text
}

// plainHeader is the bottom of the ladder: a one-line ruled header that fits
// any terminal. Something like ═════ PVE1 ═════.
func (r *Renderer) plainHeader(name, fullText string, opts RenderOptions) Art {
	// Preferred rule length; long names widen it rather than get clipped,
	// as long as the terminal allows.
	const maxRule = 60
	avail := opts.MaxWidth

	label := name
	if avail > 0 {
		label = terminal.Clip(label, avail)
	}
	labelWidth := terminal.DisplayWidth(label)

	target := maxRule
	if labelWidth+4 > target {
		target = labelWidth + 4
	}
	if avail > 0 && target > avail {
		target = avail
	}

	line := label
	if fill := target - labelWidth - 2; fill >= 2 {
		left := fill / 2
		right := fill - left
		line = strings.Repeat("═", left) + " " + label + " " + strings.Repeat("═", right)
	}

	color := resetColor
	if !opts.Monochrome && opts.Profile != terminal.ProfileNoColor {
		color = r.headerColor(opts)
		line = terminal.Wrap(opts.Profile, color, line)
	}
	return Art{
		Text:      line,
		Font:      plainFont,
		Color:     color,
		Width:     terminal.DisplayWidth(terminal.Strip(line)),
		Shortened: name != fullText,
	}
}

// headerColor picks the accent color for the plain header: the first
// gradient stop when a gradient is configured, otherwise the single color.
func (r *Renderer) headerColor(opts RenderOptions) string {
	if len(opts.Gradient) > 0 {
		if _, ok := terminal.Code(opts.Gradient[0]); ok {
			return strings.ToLower(opts.Gradient[0])
		}
	}
	return r.pickColor(opts.Color)
}

// colorize joins the art rows, applying gradient or single-color styling
// according to the terminal profile.
func (r *Renderer) colorize(rows []string, opts RenderOptions) (string, string) {
	if opts.Monochrome || opts.Profile == terminal.ProfileNoColor {
		return strings.Join(rows, "\n"), resetColor
	}

	if len(opts.Gradient) > 0 {
		codes := gradientCodes(opts.Gradient, len(rows), opts.Profile)
		colored := make([]string, len(rows))
		for i, row := range rows {
			if codes[i] == "" {
				colored[i] = row
				continue
			}
			colored[i] = codes[i] + row + terminal.Reset
		}
		return strings.Join(colored, "\n"), strings.Join(opts.Gradient, ",")
	}

	color := r.pickColor(opts.Color)
	art := strings.Join(rows, "\n")
	if code, ok := terminal.Code(color); ok && color != resetColor {
		art = code + art + terminal.Reset
	}
	return art, color
}

// gradientCodes returns one escape sequence per row. On truecolor terminals
// the configured stops are interpolated into a smooth 24-bit gradient; on
// 16-color terminals the stops cycle per row as before.
func gradientCodes(stops []string, n int, profile terminal.Profile) []string {
	codes := make([]string, n)
	if profile == terminal.ProfileTrueColor {
		for i := range codes {
			t := 0.0
			if n > 1 {
				t = float64(i) / float64(n-1)
			}
			if rgb, ok := lerpStops(stops, t); ok {
				codes[i] = terminal.ForegroundRGB(rgb[0], rgb[1], rgb[2])
			}
		}
		return codes
	}
	for i := range codes {
		if code, ok := terminal.Code(stops[i%len(stops)]); ok {
			codes[i] = code
		}
	}
	return codes
}

// lerpStops interpolates the named gradient stops at position t in [0, 1].
func lerpStops(stops []string, t float64) ([3]uint8, bool) {
	if len(stops) == 1 {
		return terminal.RGB(stops[0])
	}
	pos := t * float64(len(stops)-1)
	seg := int(pos)
	if seg >= len(stops)-1 {
		seg = len(stops) - 2
	}
	frac := pos - float64(seg)
	from, okFrom := terminal.RGB(stops[seg])
	to, okTo := terminal.RGB(stops[seg+1])
	if !okFrom || !okTo {
		return [3]uint8{}, false
	}
	var rgb [3]uint8
	for c := 0; c < 3; c++ {
		rgb[c] = uint8(float64(from[c]) + (float64(to[c])-float64(from[c]))*frac + 0.5)
	}
	return rgb, true
}

func (r *Renderer) randomFont() string {
	return r.order[r.rnd.Intn(len(r.order))]
}

func (r *Renderer) pickColor(name string) string {
	if name == "" || name == "random" {
		return supportedColors[r.rnd.Intn(len(supportedColors))]
	}
	name = strings.ToLower(name)
	if _, ok := terminal.Code(name); ok {
		return name
	}
	return resetColor
}
