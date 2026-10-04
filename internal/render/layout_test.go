package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

func frac(f float64) *float64 { return &f }

// section builds a section the way the builders do, with Lines derived
// from the items.
func section(key, title string, items ...banner.Item) banner.Section {
	s := banner.Section{Key: key, Title: title, Items: items}
	for _, it := range items {
		s.Lines = append(s.Lines, it.Text())
	}
	return s
}

func demoOutput() banner.Output {
	return banner.Output{
		Header: banner.Header{Hostname: "pve1", Art: "PVE1", Lines: []string{"Ubuntu 24.04.4 LTS (amd64)"}},
		Sections: []banner.Section{
			section("system", "System",
				banner.Item{Label: "Uptime", Value: "4d 12h"},
				banner.Item{Label: "User", Value: "root", Level: banner.LevelCrit},
				banner.Item{Label: "Last login", Value: "26h ago from 192.168.1.20"},
			),
			section("network", "Network",
				banner.Item{Label: "eth0", Value: "192.168.1.42"},
				banner.Item{Label: "From", Value: "192.168.1.20"},
			),
			section("resources", "Resources",
				banner.Item{Label: "Mem", Value: "23%", Detail: "3.7/16.0 GiB", Meter: frac(0.23)},
				banner.Item{Label: "Disk", Value: "87%", Detail: "412.0/476.0 GiB", Meter: frac(0.87), Level: banner.LevelWarn},
				banner.Item{Label: "Load", Value: "0.45", Detail: "8 cores", Meter: frac(0.06)},
			),
		},
	}
}

func plain(t *testing.T, width int) string {
	t.Helper()
	return NewRenderer(terminal.Env{Width: width}).Render(demoOutput(), config.Default())
}

func assertFits(t *testing.T, out string, width int) {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if n := terminal.DisplayWidth(terminal.Strip(line)); width > 0 && n > width {
			t.Errorf("line exceeds width %d (%d): %q", width, n, line)
		}
	}
}

func TestRender_AlignsLabelsWithinSection(t *testing.T) {
	out := plain(t, 0)
	for _, want := range []string{
		"  Uptime      4d 12h",
		"  User        root",
		"  Last login  26h ago from 192.168.1.20",
		"  eth0  192.168.1.42",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing aligned row %q in:\n%s", want, out)
		}
	}
}

func TestRender_MetersAndRightAlignedValues(t *testing.T) {
	out := plain(t, 0)
	for _, want := range []string{
		"  Mem   ██░░░░░░░░  23%  3.7/16.0 GiB",
		"  Disk  █████████░  87%  412.0/476.0 GiB",
		"  Load  █░░░░░░░░░ 0.45  8 cores",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing meter row %q in:\n%s", want, out)
		}
	}
}

func TestRender_UnknownWidthStacksSections(t *testing.T) {
	out := plain(t, 0)
	sys := strings.Index(out, "System")
	net := strings.Index(out, "Network")
	res := strings.Index(out, "Resources")
	if sys < 0 || net < 0 || res < 0 || sys >= net || net >= res {
		t.Fatalf("expected System, Network, Resources stacked in order:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "System") && strings.Contains(line, "Network") {
			t.Fatalf("piped output (width 0) must not place sections side by side:\n%s", out)
		}
	}
}

func TestRender_WideTerminalPlacesSectionsSideBySide(t *testing.T) {
	out := plain(t, 140)
	assertFits(t, out, 140)
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "System") && strings.Contains(line, "Network") && strings.Contains(line, "Resources") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected all three section titles on one row at 140 columns:\n%s", out)
	}
	if strings.Count(out, "\n") > 10 {
		t.Errorf("the side-by-side layout should be short, got %d lines:\n%s", strings.Count(out, "\n")+1, out)
	}
}

func TestRender_MediumTerminalWrapsRows(t *testing.T) {
	out := plain(t, 80)
	assertFits(t, out, 80)
	var titleRow string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "System") {
			titleRow = line
		}
	}
	if !strings.Contains(titleRow, "Network") || strings.Contains(titleRow, "Resources") {
		t.Fatalf("at 80 columns expect System+Network on one row, Resources below; got row %q in:\n%s", titleRow, out)
	}
}

func TestRender_NeverOverflowsAnyWidth(t *testing.T) {
	// The art is fitted upstream by the ascii ladder; this covers the body.
	out := demoOutput()
	out.Header.Art = ""
	for _, w := range []int{1, 2, 3, 8, 20, 30, 40, 50, 60, 79, 100, 200} {
		assertFits(t, NewRenderer(terminal.Env{Width: w}).Render(out, config.Default()), w)
	}
}

func TestRender_NarrowDropsDetailBeforeClipping(t *testing.T) {
	out := plain(t, 30)
	if strings.Contains(out, "GiB") {
		t.Errorf("details should be dropped at 30 columns:\n%s", out)
	}
	if !strings.Contains(out, "87%") {
		t.Errorf("the value must survive when details are dropped:\n%s", out)
	}
}

func TestRender_ColorsFollowLevels(t *testing.T) {
	out := NewRenderer(terminal.Env{Profile: terminal.ProfileANSI}).Render(demoOutput(), config.Default())
	if !strings.Contains(out, "\033[33m█████████") {
		t.Errorf("a warn-level meter should be yellow:\n%q", out)
	}
	if !strings.Contains(out, "\033[32m██") {
		t.Errorf("a normal meter should be green:\n%q", out)
	}
	if !strings.Contains(out, "\033[1m\033[31mroot") {
		t.Errorf("root should be bold red:\n%q", out)
	}
}

func TestRender_NoColorProfileEmitsNoEscapes(t *testing.T) {
	if out := plain(t, 120); strings.Contains(out, "\033[") {
		t.Fatalf("no-color output contains escapes:\n%q", out)
	}
}

func TestRender_TitleUsesGradientAccent(t *testing.T) {
	cfg := config.Default()
	cfg.ASCII.Gradient = []string{"purple", "blue"}
	out := NewRenderer(terminal.Env{Profile: terminal.ProfileANSI}).Render(demoOutput(), cfg)
	if !strings.Contains(out, "\033[35mSystem") {
		t.Fatalf("section titles should take the first gradient color:\n%q", out)
	}
}

func TestRender_HostnameDisabledOmitsArt(t *testing.T) {
	out := demoOutput()
	out.Header.Art = ""
	got := NewRenderer(terminal.Env{}).Render(out, config.Default())
	if strings.HasPrefix(got, "\n\n") {
		t.Fatalf("no art should leave no empty art block:\n%q", got)
	}
}

func TestRender_ClipsWideRunesByColumns(t *testing.T) {
	out := banner.Output{
		Header: banner.Header{Hostname: "vm", Art: "VM"},
		Sections: []banner.Section{
			section("system", "System", banner.Item{Label: "User", Value: "田中太郎 /home/田中太郎"}),
		},
	}
	assertFits(t, NewRenderer(terminal.Env{Width: 20}).Render(out, config.Default()), 20)
}

func TestOrderSections(t *testing.T) {
	all := []banner.Section{{Key: "system"}, {Key: "network"}, {Key: "resources"}}
	tests := []struct {
		name    string
		desired []string
		want    []string
	}{
		{"configured order", []string{"resources", "system", "network"}, []string{"resources", "system", "network"}},
		{"subset", []string{"network"}, []string{"network"}},
		{"header only keeps builder order", []string{"header"}, []string{"system", "network", "resources"}},
		{"empty keeps builder order", nil, []string{"system", "network", "resources"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orderSections(all, tt.desired)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d sections, want %d", len(got), len(tt.want))
			}
			for i, key := range tt.want {
				if got[i].Key != key {
					t.Errorf("section[%d] = %q, want %q", i, got[i].Key, key)
				}
			}
		})
	}
}

func TestRenderCompact(t *testing.T) {
	cfg := config.Config{Layout: config.LayoutConfig{Compact: true}}
	out := NewRenderer(terminal.Env{}).Render(demoOutput(), cfg)
	if strings.Contains(out, "\n") {
		t.Fatalf("compact output must be one line:\n%s", out)
	}
	for _, want := range []string{"PVE1", "Uptime: 4d 12h", "Mem: 23% 3.7/16.0 GiB", " | "} {
		if !strings.Contains(out, want) {
			t.Errorf("compact output missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "█") {
		t.Errorf("compact output must not include art or meters: %s", out)
	}
}

func TestRenderCompact_ClipsAndDedupes(t *testing.T) {
	out := demoOutput()
	out.Header.Hostname = "pve1.home.lan"
	out.Header.Lines = append([]string{"pve1.home.lan"}, out.Header.Lines...)
	cfg := config.Config{Layout: config.LayoutConfig{Compact: true}}

	got := NewRenderer(terminal.Env{}).Render(out, cfg)
	if strings.Count(strings.ToLower(got), "pve1.home.lan") != 1 {
		t.Errorf("compact output repeats the hostname: %s", got)
	}
	assertFits(t, NewRenderer(terminal.Env{Width: 50}).Render(out, cfg), 50)
}

func TestRenderJSON(t *testing.T) {
	cfg := config.Default()
	out := demoOutput()
	out.Sections[2].Data = map[string]any{"memory_used_percent": 23}
	doc, err := RenderJSON(out, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Hostname string `json:"hostname"`
		Sections []struct {
			Key   string   `json:"key"`
			Lines []string `json:"lines"`
			Items []struct {
				Label string   `json:"label"`
				Value string   `json:"value"`
				Meter *float64 `json:"meter"`
				Level string   `json:"level"`
			} `json:"items"`
			Data map[string]any `json:"data"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(doc), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, doc)
	}
	if parsed.Hostname != "pve1" || strings.Contains(doc, "PVE1") {
		t.Errorf("JSON should carry the hostname and no art:\n%s", doc)
	}
	if got := parsed.Sections[0].Key; got != "system" {
		t.Errorf("sections must follow the layout order, first is %q", got)
	}
	res := parsed.Sections[2]
	if res.Lines[1] != "Disk: 87% 412.0/476.0 GiB" {
		t.Errorf("lines keep the plain one-line form, got %q", res.Lines[1])
	}
	if res.Items[1].Level != "warn" || res.Items[1].Meter == nil || *res.Items[1].Meter != 0.87 {
		t.Errorf("items carry meter and level, got %+v", res.Items[1])
	}
	if res.Items[0].Level != "" {
		t.Errorf("normal level is omitted, got %q", res.Items[0].Level)
	}
	if res.Data["memory_used_percent"] != float64(23) {
		t.Errorf("data keeps raw numbers, got %v", res.Data)
	}
}

func TestApplyConfig(t *testing.T) {
	base := terminal.Env{Width: 120, Profile: terminal.ProfileANSI}

	if capped := ApplyConfig(base, config.Config{Layout: config.LayoutConfig{MaxWidth: 80}}); capped.Width != 80 {
		t.Errorf("max_width should cap detected width: got %d", capped.Width)
	}
	if wider := ApplyConfig(base, config.Config{Layout: config.LayoutConfig{MaxWidth: 200}}); wider.Width != 120 {
		t.Errorf("max_width above terminal width must not widen: got %d", wider.Width)
	}
	unknown := ApplyConfig(terminal.Env{Profile: terminal.ProfileANSI}, config.Config{Layout: config.LayoutConfig{MaxWidth: 80}})
	if unknown.Width != 80 {
		t.Errorf("max_width should apply when terminal width is unknown: got %d", unknown.Width)
	}
	if mono := ApplyConfig(base, config.Config{ASCII: config.ASCIIConfig{Monochrome: true}}); mono.Profile != terminal.ProfileNoColor {
		t.Errorf("monochrome config should force ProfileNoColor, got %v", mono.Profile)
	}
	if untouched := ApplyConfig(base, config.Config{}); untouched != base {
		t.Errorf("empty config should leave env unchanged: %+v", untouched)
	}
}
