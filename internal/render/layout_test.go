package render

import (
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

func TestRenderer_Render(t *testing.T) {
	tests := []struct {
		name         string
		output       banner.Output
		cfg          config.Config
		wantContains []string
	}{
		{
			name: "basic header and sections",
			output: banner.Output{
				Header: banner.Header{
					Art:   "ASCII ART",
					Lines: []string{"Linux 6.0 (x86_64)"},
				},
				Sections: []banner.Section{
					{
						Key:   "system",
						Title: "System",
						Lines: []string{"Uptime: 1d 2h 30m"},
					},
				},
			},
			cfg:          config.Default(),
			wantContains: []string{"ASCII ART", "Linux 6.0 (x86_64)", "System", "Uptime: 1d 2h 30m"},
		},
		{
			name: "multiple sections with ordering",
			output: banner.Output{
				Header: banner.Header{
					Art: "HOST",
				},
				Sections: []banner.Section{
					{
						Key:   "resources",
						Title: "Resources",
						Lines: []string{"Mem: 8GB"},
					},
					{
						Key:   "system",
						Title: "System",
						Lines: []string{"Uptime: 1d"},
					},
				},
			},
			cfg: config.Config{
				Layout: config.LayoutConfig{
					Sections: []string{"system", "resources"},
				},
			},
			wantContains: []string{"System", "Resources"},
		},
		{
			name: "empty sections are skipped",
			output: banner.Output{
				Header: banner.Header{
					Art: "HOST",
				},
				Sections: []banner.Section{
					{
						Key:   "system",
						Title: "System",
						Lines: []string{},
					},
					{
						Key:   "network",
						Title: "Network",
						Lines: []string{"Primary: 192.168.1.1"},
					},
				},
			},
			cfg:          config.Default(),
			wantContains: []string{"Network", "Primary: 192.168.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(terminal.Env{}) // No color for deterministic output
			result := r.Render(tt.output, tt.cfg)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Render() result missing %q\nGot:\n%s", want, result)
				}
			}
		})
	}
}

func TestRenderer_RenderCompact(t *testing.T) {
	tests := []struct {
		name         string
		output       banner.Output
		cfg          config.Config
		wantContains []string
		separator    string
	}{
		{
			name: "compact mode with separator",
			output: banner.Output{
				Header: banner.Header{
					Hostname: "host",
					Art:      "line1\nline2\nline3",
					Lines:    []string{"Linux 6.0"},
				},
				Sections: []banner.Section{
					{
						Key:   "system",
						Title: "System",
						Lines: []string{"Uptime: 1d"},
					},
				},
			},
			cfg: config.Config{
				Layout: config.LayoutConfig{
					Compact: true,
				},
			},
			wantContains: []string{"HOST", "Linux 6.0", "System", "Uptime: 1d"},
			separator:    " | ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(terminal.Env{})
			result := r.Render(tt.output, tt.cfg)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Render() compact result missing %q\nGot: %s", want, result)
				}
			}

			if !strings.Contains(result, tt.separator) {
				t.Errorf("Render() compact result missing separator %q", tt.separator)
			}
		})
	}
}

func TestOrderSections(t *testing.T) {
	tests := []struct {
		name     string
		sections []banner.Section
		desired  []string
		wantKeys []string
	}{
		{
			name: "ordered by desired list",
			sections: []banner.Section{
				{Key: "resources"},
				{Key: "system"},
				{Key: "network"},
			},
			desired:  []string{"system", "network", "resources"},
			wantKeys: []string{"system", "network", "resources"},
		},
		{
			name: "partial ordering",
			sections: []banner.Section{
				{Key: "resources"},
				{Key: "system"},
				{Key: "network"},
			},
			desired:  []string{"system"},
			wantKeys: []string{"system"},
		},
		{
			name: "alphabetical fallback when desired is empty",
			sections: []banner.Section{
				{Key: "resources"},
				{Key: "system"},
				{Key: "network"},
			},
			desired:  []string{},
			wantKeys: []string{"network", "resources", "system"},
		},
		{
			name: "desired order with missing keys",
			sections: []banner.Section{
				{Key: "system"},
			},
			desired:  []string{"network", "system", "resources"},
			wantKeys: []string{"system"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orderSections(tt.sections, tt.desired)

			if len(result) != len(tt.wantKeys) {
				t.Fatalf("got %d sections, want %d", len(result), len(tt.wantKeys))
			}

			for i, want := range tt.wantKeys {
				if result[i].Key != want {
					t.Errorf("section[%d].Key = %q, want %q", i, result[i].Key, want)
				}
			}
		})
	}
}

func TestRenderer_HighlightResource(t *testing.T) {
	tests := []struct {
		name         string
		section      banner.Section
		line         string
		disableColor bool
		wantContains string
	}{
		{
			name: "memory high usage (red)",
			section: banner.Section{
				Key: "resources",
				Data: map[string]any{
					"memory_used_percent": 95,
				},
			},
			line:         "Mem: 15GB used / 16GB",
			disableColor: false,
			wantContains: "Mem:",
		},
		{
			name: "memory warning usage (yellow)",
			section: banner.Section{
				Key: "resources",
				Data: map[string]any{
					"memory_used_percent": 80,
				},
			},
			line:         "Mem: 12GB used / 16GB",
			disableColor: false,
			wantContains: "Mem:",
		},
		{
			name: "disk normal usage (no color)",
			section: banner.Section{
				Key: "resources",
				Data: map[string]any{
					"disk_used_percent": 50,
				},
			},
			line:         "Disk: 250GB used / 500GB",
			disableColor: false,
			wantContains: "Disk:",
		},
		{
			name: "no data map",
			section: banner.Section{
				Key:  "resources",
				Data: nil,
			},
			line:         "Mem: 8GB used / 16GB",
			disableColor: false,
			wantContains: "Mem:",
		},
		{
			name: "color disabled",
			section: banner.Section{
				Key: "resources",
				Data: map[string]any{
					"memory_used_percent": 95,
				},
			},
			line:         "Mem: 15GB used / 16GB",
			disableColor: true,
			wantContains: "Mem: 15GB used / 16GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(envFor(tt.disableColor))
			result := r.highlightResource(tt.section, tt.line)

			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("highlightResource() = %q, want to contain %q", result, tt.wantContains)
			}

			// When color is disabled, output should equal input
			if tt.disableColor && result != tt.line {
				t.Errorf("highlightResource() with disabled color = %q, want %q", result, tt.line)
			}
		})
	}
}

func TestRenderer_WrapForPercent(t *testing.T) {
	tests := []struct {
		name         string
		pct          int
		line         string
		disableColor bool
		wantColor    bool
	}{
		{"critical threshold", 95, "Mem: 95%", false, true},
		{"warning threshold", 80, "Mem: 80%", false, true},
		{"normal threshold", 50, "Mem: 50%", false, false},
		{"zero percent", 0, "Mem: 0%", false, false},
		{"color disabled", 95, "Mem: 95%", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(envFor(tt.disableColor))
			result := r.wrapForPercent(tt.pct, tt.line)

			hasColor := result != tt.line
			if hasColor != tt.wantColor {
				t.Errorf("wrapForPercent(%d, %q) hasColor=%v, want %v", tt.pct, tt.line, hasColor, tt.wantColor)
			}
		})
	}
}

// envFor maps the old disable-color boolean used across these tests to a
// terminal environment.
func envFor(disable bool) terminal.Env {
	if disable {
		return terminal.Env{}
	}
	return terminal.Env{Profile: terminal.ProfileANSI}
}

func TestRenderer_CompactIsSingleLine(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{
			Hostname: "pve1",
			Art:      "██\n██\n██",
			Lines:    []string{"Linux 6.8"},
		},
		Sections: []banner.Section{
			{Key: "system", Title: "System", Lines: []string{"Uptime: 1d"}},
		},
	}
	cfg := config.Config{Layout: config.LayoutConfig{Compact: true}}
	result := NewRenderer(terminal.Env{}).Render(out, cfg)

	if strings.Contains(result, "\n") {
		t.Fatalf("compact output must be a single line, got:\n%s", result)
	}
	if !strings.Contains(result, "PVE1") {
		t.Errorf("compact output missing hostname, got: %s", result)
	}
	if strings.Contains(result, "██") {
		t.Errorf("compact output must not include ASCII art, got: %s", result)
	}
}

func TestRenderer_ClipsBodyLinesToWidth(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{Hostname: "host", Art: "HOST"},
		Sections: []banner.Section{
			{
				Key:   "system",
				Title: "System",
				Lines: []string{"Last login: Fri, 10 Oct 2025 09:45:00 PDT from 203.0.113.10"},
			},
		},
	}
	cfg := config.Default()
	result := NewRenderer(terminal.Env{Width: 40}).Render(out, cfg)

	for _, line := range strings.Split(result, "\n") {
		if n := len([]rune(line)); n > 40 {
			t.Errorf("line exceeds width 40 (%d): %q", n, line)
		}
	}
	if !strings.Contains(result, "…") {
		t.Errorf("expected clipped line to end with ellipsis, got:\n%s", result)
	}
}

func TestRenderer_UnconstrainedLeavesLinesAlone(t *testing.T) {
	long := "Last login: Fri, 10 Oct 2025 09:45:00 PDT from 203.0.113.10"
	out := banner.Output{
		Header: banner.Header{Hostname: "host", Art: "HOST"},
		Sections: []banner.Section{
			{Key: "system", Title: "System", Lines: []string{long}},
		},
	}
	result := NewRenderer(terminal.Env{}).Render(out, config.Default())
	if !strings.Contains(result, long) {
		t.Errorf("unconstrained render should not clip lines, got:\n%s", result)
	}
}

func TestRenderJSON(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{
			Hostname: "pve1",
			Art:      "ART",
			Lines:    []string{"Linux 6.8 (x86_64)"},
		},
		Sections: []banner.Section{
			{Key: "resources", Title: "Resources", Lines: []string{"Mem: 23% used"}, Data: map[string]any{"memory_used_percent": 23}},
			{Key: "system", Title: "System", Lines: []string{"Uptime: 1d"}},
		},
	}
	cfg := config.Default()
	doc, err := RenderJSON(out, cfg)
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}
	for _, want := range []string{`"hostname": "pve1"`, `"Linux 6.8 (x86_64)"`, `"memory_used_percent": 23`} {
		if !strings.Contains(doc, want) {
			t.Errorf("JSON missing %q:\n%s", want, doc)
		}
	}
	if strings.Contains(doc, "ART") {
		t.Errorf("JSON should not include ASCII art:\n%s", doc)
	}
	// Sections must follow the configured order: system before resources.
	if strings.Index(doc, `"key": "system"`) > strings.Index(doc, `"key": "resources"`) {
		t.Errorf("JSON sections not in layout order:\n%s", doc)
	}
}

func TestRenderer_CompactClipsToWidth(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{
			Hostname: "pve1",
			Lines:    []string{"Linux 6.8 (x86_64)"},
		},
		Sections: []banner.Section{
			{Key: "system", Title: "System", Lines: []string{
				"Uptime: 4d 12h 30m",
				"Last login: Fri, 10 Oct 2025 09:45:00 PDT (203.0.113.10)",
			}},
		},
	}
	cfg := config.Config{Layout: config.LayoutConfig{Compact: true}}
	result := NewRenderer(terminal.Env{Width: 50}).Render(out, cfg)

	if n := len([]rune(result)); n > 50 {
		t.Errorf("compact line exceeds width 50 (%d): %q", n, result)
	}
	if !strings.HasSuffix(result, "…") {
		t.Errorf("expected clipped compact line to end with ellipsis: %q", result)
	}
}

func TestApplyConfig(t *testing.T) {
	base := terminal.Env{Width: 120, Profile: terminal.ProfileANSI}

	capped := ApplyConfig(base, config.Config{Layout: config.LayoutConfig{MaxWidth: 80}})
	if capped.Width != 80 {
		t.Errorf("max_width should cap detected width: got %d", capped.Width)
	}

	wider := ApplyConfig(base, config.Config{Layout: config.LayoutConfig{MaxWidth: 200}})
	if wider.Width != 120 {
		t.Errorf("max_width above terminal width must not widen: got %d", wider.Width)
	}

	unknown := ApplyConfig(terminal.Env{Profile: terminal.ProfileANSI}, config.Config{Layout: config.LayoutConfig{MaxWidth: 80}})
	if unknown.Width != 80 {
		t.Errorf("max_width should apply when terminal width is unknown: got %d", unknown.Width)
	}

	mono := ApplyConfig(base, config.Config{ASCII: config.ASCIIConfig{Monochrome: true}})
	if mono.Profile != terminal.ProfileNoColor {
		t.Errorf("monochrome config should force ProfileNoColor, got %v", mono.Profile)
	}

	untouched := ApplyConfig(base, config.Config{})
	if untouched != base {
		t.Errorf("empty config should leave env unchanged: %+v", untouched)
	}
}

func TestRenderer_ClipsSectionTitles(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{Hostname: "vm", Art: "VM"},
		Sections: []banner.Section{
			{Key: "resources", Title: "Resources", Lines: []string{"Mem: 4%"}},
		},
	}
	result := NewRenderer(terminal.Env{Width: 8}).Render(out, config.Default())
	for _, line := range strings.Split(result, "\n") {
		if n := len([]rune(line)); n > 8 {
			t.Errorf("line exceeds width 8 (%d): %q", n, line)
		}
	}
}

func TestRenderer_CompactDoesNotDuplicateHostname(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{
			Hostname: "pve1.home.lan",
			// The width ladder shortened the art, so the full hostname was
			// added as an info line.
			Lines: []string{"pve1.home.lan", "Linux 6.8 (x86_64)"},
		},
	}
	cfg := config.Config{Layout: config.LayoutConfig{Compact: true}}
	result := NewRenderer(terminal.Env{}).Render(out, cfg)

	if strings.Count(strings.ToLower(result), "pve1.home.lan") != 1 {
		t.Fatalf("compact output repeats the hostname: %q", result)
	}
	if !strings.Contains(result, "Linux 6.8 (x86_64)") {
		t.Fatalf("compact output lost the OS line: %q", result)
	}
}

func TestRenderer_TinyWidthsNeverOverflow(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{Hostname: "vm", Art: "…"},
		Sections: []banner.Section{
			{Key: "system", Title: "System", Lines: []string{"Uptime: 4d 12h"}},
		},
	}
	for _, width := range []int{1, 2, 3, 4} {
		result := NewRenderer(terminal.Env{Width: width}).Render(out, config.Default())
		for _, line := range strings.Split(result, "\n") {
			if n := len([]rune(line)); n > width {
				t.Errorf("width %d: line has %d columns: %q", width, n, line)
			}
		}
	}
}

func TestRenderer_ClipsWideRunesByColumns(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{Hostname: "vm", Art: "VM"},
		Sections: []banner.Section{
			{Key: "system", Title: "System", Lines: []string{"User: 田中太郎 /home/田中太郎"}},
		},
	}
	result := NewRenderer(terminal.Env{Width: 20}).Render(out, config.Default())
	for _, line := range strings.Split(result, "\n") {
		if n := terminal.DisplayWidth(line); n > 20 {
			t.Errorf("line occupies %d columns at width 20: %q", n, line)
		}
	}
}
