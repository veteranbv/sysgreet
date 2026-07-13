package render

import (
	"sort"
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
	width     int
}

// NewRenderer instantiates a renderer for the given terminal environment.
// A zero env.Width leaves lines unclipped.
func NewRenderer(env terminal.Env) Renderer {
	return Renderer{colorizer: NewColorizer(env.Profile), width: env.Width}
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

	var builder strings.Builder
	builder.WriteString("\n")
	builder.WriteString(out.Header.Art)
	if len(out.Header.Lines) > 0 {
		builder.WriteString("\n")
		for _, line := range out.Header.Lines {
			builder.WriteString(r.clip(line, 0))
			builder.WriteString("\n")
		}
	}

	// The indent has to fit inside the width cap too; drop it on absurdly
	// narrow terminals rather than overflow.
	indent := bodyIndent
	if r.width > 0 && r.width <= len(bodyIndent) {
		indent = ""
	}

	sections := orderSections(out.Sections, cfg.Layout.Sections)
	for _, section := range sections {
		if len(section.Lines) == 0 {
			continue
		}
		builder.WriteString("\n")
		builder.WriteString(r.clip(section.Title, 0))
		builder.WriteString("\n")
		for _, line := range section.Lines {
			formatted := r.clip(line, len(indent))
			if section.Key == "resources" {
				formatted = r.highlightResource(section, formatted)
			}
			builder.WriteString(indent)
			builder.WriteString(formatted)
			builder.WriteString("\n")
		}
	}
	return strings.TrimRight(builder.String(), "\n")
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

// clip truncates a line to the terminal width, accounting for indent and
// marking the cut with an ellipsis. Zero width leaves the line untouched.
func (r Renderer) clip(line string, indent int) string {
	if r.width <= 0 {
		return line
	}
	limit := r.width - indent
	if limit < 1 {
		limit = 1
	}
	runes := []rune(line)
	if len(runes) <= limit {
		return line
	}
	if limit == 1 {
		return "…"
	}
	return string(runes[:limit-1]) + "…"
}

func orderSections(sections []banner.Section, desired []string) []banner.Section {
	lookup := make(map[string]banner.Section)
	keys := []string{}
	for _, s := range sections {
		lookup[s.Key] = s
		keys = append(keys, s.Key)
	}
	var ordered []banner.Section
	for _, key := range desired {
		if sec, ok := lookup[key]; ok {
			ordered = append(ordered, sec)
		}
	}
	if len(ordered) == 0 {
		sort.Strings(keys)
		for _, key := range keys {
			ordered = append(ordered, lookup[key])
		}
	}
	return ordered
}

func (r Renderer) highlightResource(section banner.Section, line string) string {
	data := section.Data
	if data == nil {
		return line
	}
	switch {
	case strings.HasPrefix(line, "Mem:"):
		if pct, ok := data["memory_used_percent"].(int); ok {
			return r.wrapForPercent(pct, line)
		}
	case strings.HasPrefix(line, "Disk:"):
		if pct, ok := data["disk_used_percent"].(int); ok {
			return r.wrapForPercent(pct, line)
		}
	case strings.HasPrefix(line, "CPU:") && strings.Contains(line, "%"):
		if pct, ok := data["cpu_usage_percent"].(float64); ok {
			return r.wrapForPercent(int(pct+0.5), line)
		}
	}
	return line
}

func (r Renderer) wrapForPercent(pct int, line string) string {
	color := ""
	switch {
	case pct >= 90:
		color = "red"
	case pct >= 75:
		color = "yellow"
	case pct >= 0:
		color = "green"
	}
	if color == "" || color == "green" {
		return line
	}
	return r.colorizer.Wrap(color, line)
}
