package render

import (
	"strings"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

// bodyIndent prefixes every section line in the full layout.
const bodyIndent = "  "

// Renderer formats banner output into terminal-friendly text.
type Renderer struct {
	colorizer Colorizer
	profile   terminal.Profile
	width     int
	accent    string // color for section titles
}

// NewRenderer instantiates a renderer for the given terminal environment.
// A zero env.Width leaves lines unclipped and stacks sections vertically.
func NewRenderer(env terminal.Env) Renderer {
	return Renderer{
		colorizer: NewColorizer(env.Profile),
		profile:   env.Profile,
		width:     env.Width,
		accent:    "cyan",
	}
}

// ApplyConfig folds config-driven constraints into the detected terminal
// environment: layout.max_width caps the width and ascii.monochrome forces
// plain output everywhere, including resource threshold highlights.
func ApplyConfig(env terminal.Env, cfg config.Config) terminal.Env {
	if max := cfg.Layout.MaxWidth; max > 0 && (env.Width == 0 || max < env.Width) {
		env.Width = max
	}
	if cfg.ASCII.Monochrome {
		env.Profile = terminal.ProfileNoColor
	}
	return env
}

// Render produces the final banner string.
func (r Renderer) Render(out banner.Output, cfg config.Config) string {
	if cfg.Layout.Compact {
		return r.renderCompact(out, cfg)
	}
	r.accent = accentColor(cfg)

	var lines []string
	if out.Header.Art != "" {
		lines = append(lines, "", out.Header.Art)
	}
	if len(out.Header.Lines) > 0 {
		lines = append(lines, "")
		for _, line := range out.Header.Lines {
			lines = append(lines, terminal.Dim(r.profile, r.clip(line, 0)))
		}
	}
	if body := r.renderBody(orderSections(out.Sections, cfg.Layout.Sections)); len(body) > 0 {
		lines = append(lines, "")
		lines = append(lines, body...)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// accentColor ties section titles to the banner: the first gradient stop,
// else the single art color.
func accentColor(cfg config.Config) string {
	for _, c := range append(append([]string{}, cfg.ASCII.Gradient...), cfg.ASCII.Color) {
		if _, ok := terminal.Code(c); ok {
			return strings.ToLower(c)
		}
	}
	return "cyan"
}

// renderCompact emits a single pipe-separated line using the plain hostname
// rather than the multi-line art.
func (r Renderer) renderCompact(out banner.Output, cfg config.Config) string {
	parts := []string{strings.ToUpper(out.Header.Hostname)}
	for _, line := range out.Header.Lines {
		// The full-hostname fallback line exists to supplement shortened
		// art; compact output already leads with the full name.
		if line == out.Header.Hostname {
			continue
		}
		parts = append(parts, line)
	}
	sections := orderSections(out.Sections, cfg.Layout.Sections)
	for _, section := range sections {
		if len(section.Lines) == 0 {
			continue
		}
		parts = append(parts, section.Title)
		parts = append(parts, section.Lines...)
	}
	return r.clip(strings.Join(parts, " | "), 0)
}

// clip truncates a line to the terminal width in display columns,
// accounting for indent and marking the cut with an ellipsis. Zero width
// leaves the line untouched.
func (r Renderer) clip(line string, indent int) string {
	if r.width <= 0 {
		return line
	}
	limit := r.width - indent
	if limit < 1 {
		limit = 1
	}
	return terminal.Clip(line, limit)
}

func orderSections(sections []banner.Section, desired []string) []banner.Section {
	lookup := make(map[string]banner.Section)
	for _, s := range sections {
		lookup[s.Key] = s
	}
	var ordered []banner.Section
	for _, key := range desired {
		if sec, ok := lookup[key]; ok {
			ordered = append(ordered, sec)
		}
	}
	if len(ordered) == 0 {
		// Nothing configured matched (for example sections: [header]);
		// keep the order the builders produced.
		return sections
	}
	return ordered
}
