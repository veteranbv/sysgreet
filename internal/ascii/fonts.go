package ascii

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/veteranbv/sysgreet/assets"
)

const resetColor = "reset"

var colorCodes = map[string]string{
	"reset":       "\033[0m",
	"red":         "\033[31m",
	"green":       "\033[32m",
	"yellow":      "\033[33m",
	"blue":        "\033[34m",
	"purple":      "\033[35m",
	"cyan":        "\033[36m",
	"gray":        "\033[37m",
	"white":       "\033[37m",
	"brightblue":  "\033[94m",
	"brightcyan":  "\033[96m",
	"brightwhite": "\033[97m",
}

var supportedColors = []string{"red", "green", "yellow", "blue", "purple", "cyan", "gray", "white"}

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

// Render creates ASCII art for the input string using the specified font and color.
// If fontName or colorName equal "random", selections are randomized from the embedded sets.
// If gradient is provided, it will apply different colors to each line of the banner.
func (r *Renderer) Render(text, fontName, colorName string, monochrome bool) (string, string, string, error) {
	return r.RenderWithGradient(text, fontName, colorName, nil, monochrome)
}

// RenderWithGradient creates ASCII art with optional gradient coloring per line.
func (r *Renderer) RenderWithGradient(text, fontName, colorName string, gradient []string, monochrome bool) (string, string, string, error) {
	if fontName == "" || fontName == "random" {
		fontName = r.randomFont()
	} else if _, ok := r.fonts[fontName]; !ok {
		fontName = r.randomFont()
	}

	fontData := r.fonts[fontName]
	fig := figure.NewFigureWithFont(text, bytes.NewReader(fontData), true)
	rows := fig.Slicify()

	color := resetColor
	var asciiArt string

	if !monochrome && len(gradient) > 0 {
		// Apply gradient coloring per line
		var coloredRows []string
		for i, row := range rows {
			gradientColor := gradient[i%len(gradient)]
			if code, ok := colorCodes[strings.ToLower(gradientColor)]; ok {
				coloredRows = append(coloredRows, fmt.Sprintf("%s%s%s", code, row, colorCodes[resetColor]))
			} else {
				coloredRows = append(coloredRows, row)
			}
		}
		asciiArt = strings.Join(coloredRows, "\n")
		color = strings.Join(gradient, ",")
	} else if !monochrome {
		// Single color mode
		asciiArt = strings.Join(rows, "\n")
		color = r.pickColor(colorName)
		if code, ok := colorCodes[color]; ok && color != resetColor {
			asciiArt = fmt.Sprintf("%s%s%s", code, asciiArt, colorCodes[resetColor])
		}
	} else {
		asciiArt = strings.Join(rows, "\n")
	}

	return asciiArt, fontName, color, nil
}

func (r *Renderer) randomFont() string {
	return r.order[r.rnd.Intn(len(r.order))]
}

func (r *Renderer) pickColor(name string) string {
	if name == "" || name == "random" {
		return supportedColors[r.rnd.Intn(len(supportedColors))]
	}
	name = strings.ToLower(name)
	if _, ok := colorCodes[name]; ok {
		return name
	}
	return resetColor
}
