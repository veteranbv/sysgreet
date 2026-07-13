package banner

import (
	"context"
	"errors"
	"strings"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

// Section represents a rendered section of the banner body.
type Section struct {
	Key   string
	Title string
	Lines []string
	Data  map[string]any
}

// Header captures ASCII art metadata for the hostname banner.
type Header struct {
	Hostname string // plain hostname, for compact and JSON output
	Art      string
	Font     string
	Color    string
	Lines    []string
}

// Output encapsulates the full banner content ready for layout rendering.
type Output struct {
	Header   Header
	Sections []Section
}

// Builder generates a banner section when enabled.
type Builder interface {
	Key() string
	Enabled(cfg config.Config) bool
	Build(snap collectors.Snapshot, cfg config.Config) (Section, bool)
}

// Banner orchestrates collectors and builders to produce the final output.
type Banner struct {
	providers collectors.Providers
	ascii     *ascii.Renderer
	builders  []Builder
}

// New creates a Banner orchestrator.
func New(providers collectors.Providers, renderer *ascii.Renderer, builders []Builder) (*Banner, error) {
	if renderer == nil {
		return nil, errors.New("ascii renderer is required")
	}
	return &Banner{providers: providers, ascii: renderer, builders: builders}, nil
}

// Build produces the banner output using the provided configuration and
// terminal environment.
func (b *Banner) Build(ctx context.Context, cfg config.Config, env terminal.Env) (Output, collectors.Snapshot, error) {
	snap := b.providers.Gather(ctx)
	return b.BuildWithSnapshot(snap, cfg, env), snap, nil
}

// BuildWithSnapshot renders banner output using a pre-collected snapshot (e.g., for demo mode).
func (b *Banner) BuildWithSnapshot(snap collectors.Snapshot, cfg config.Config, env terminal.Env) Output {
	header := b.buildHeader(snap, cfg, env)
	sections := b.buildSections(snap, cfg)
	return Output{Header: header, Sections: sections}
}

func (b *Banner) buildHeader(snap collectors.Snapshot, cfg config.Config, env terminal.Env) Header {
	name := snap.System.Hostname
	if strings.TrimSpace(name) == "" {
		name = "sysgreet"
	}
	art, err := b.ascii.Render(name, ascii.RenderOptions{
		Font:          cfg.ASCII.Font,
		Color:         cfg.ASCII.Color,
		Gradient:      cfg.ASCII.Gradient,
		Monochrome:    cfg.ASCII.Monochrome,
		Uppercase:     true,
		MaxWidth:      env.Width,
		ShortenDomain: true,
		Profile:       env.Profile,
	})
	if err != nil {
		// Fallback to plain text when ASCII rendering fails.
		art = ascii.Art{Text: name, Font: "plain", Color: "reset"}
	}
	lines := []string{}
	if art.Shortened {
		// The domain was dropped from the art to fit the terminal; keep
		// the full name visible on its own line.
		lines = append(lines, name)
	}
	if cfg.Display.OS {
		line := snap.System.OS
		if snap.System.OSVersion != "" {
			line += " " + snap.System.OSVersion
		}
		if snap.System.Arch != "" {
			line += " (" + snap.System.Arch + ")"
		}
		lines = append(lines, line)
	}
	return Header{Hostname: name, Art: art.Text, Font: art.Font, Color: art.Color, Lines: lines}
}

func (b *Banner) buildSections(snap collectors.Snapshot, cfg config.Config) []Section {
	var sections []Section
	for _, builder := range b.builders {
		if builder == nil {
			continue
		}
		if !builder.Enabled(cfg) {
			continue
		}
		if section, ok := builder.Build(snap, cfg); ok && len(section.Lines) > 0 {
			sections = append(sections, section)
		}
	}
	return sections
}
