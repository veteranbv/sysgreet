package ascii

import (
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/terminal"
)

func mustRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}
	return r
}

func TestRendererRenderSpecificFont(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("host", RenderOptions{Font: "standard", Monochrome: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if art.Font != "standard" {
		t.Fatalf("expected font 'standard', got %s", art.Font)
	}
	if len(art.Text) == 0 {
		t.Fatalf("expected non-empty art output")
	}
	if art.Text == "host" {
		t.Fatalf("expected ASCII art, got plain text")
	}
	if art.Width <= 0 {
		t.Fatalf("expected positive art width, got %d", art.Width)
	}
}

func TestRendererRandomFontSelection(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("host", RenderOptions{Font: "random", Color: "random", Profile: terminal.ProfileANSI})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	found := false
	for _, candidate := range r.Fonts() {
		if art.Font == candidate {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("random font %s not in available list", art.Font)
	}
	if art.Color == "reset" {
		t.Fatalf("expected random color selection")
	}
}

func TestRendererUnknownFontFallsBackDeterministically(t *testing.T) {
	r := mustRenderer(t)
	first, err := r.Render("host", RenderOptions{Font: "no-such-font", Monochrome: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if first.Font != defaultFont {
		t.Fatalf("expected fallback to %q, got %q", defaultFont, first.Font)
	}
	// A typo must not roll a new font per invocation.
	for i := 0; i < 3; i++ {
		art, err := r.Render("host", RenderOptions{Font: "no-such-font", Monochrome: true})
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}
		if art.Font != first.Font {
			t.Fatalf("unknown font fallback not deterministic: %q vs %q", art.Font, first.Font)
		}
	}
}

func TestRendererFitsWithinMaxWidth(t *testing.T) {
	r := mustRenderer(t)
	for _, width := range []int{40, 60, 80, 120} {
		art, err := r.Render("proxmox-node-01", RenderOptions{
			Font:      "ANSI Regular",
			Uppercase: true,
			MaxWidth:  width,
			Profile:   terminal.ProfileNoColor,
		})
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}
		if art.Width > width {
			t.Errorf("width %d: art is %d columns wide", width, art.Width)
		}
		for _, line := range strings.Split(art.Text, "\n") {
			if n := len([]rune(terminal.Strip(line))); n > width {
				t.Errorf("width %d: line overflows at %d columns: %q", width, n, line)
			}
		}
	}
}

func TestRendererShortensDomainToFit(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("pve1.home.lan", RenderOptions{
		Font:          "ANSI Regular",
		Uppercase:     true,
		MaxWidth:      60,
		ShortenDomain: true,
		Profile:       terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !art.Shortened {
		t.Fatalf("expected domain to be dropped at width 60, got font %q width %d", art.Font, art.Width)
	}
	if art.Width > 60 {
		t.Fatalf("shortened art still overflows: %d columns", art.Width)
	}
}

func TestRendererNeverShortensLiteralText(t *testing.T) {
	r := mustRenderer(t)
	// --text content must survive verbatim: a dotted string squeezed into
	// a tiny width may shrink or clip, but never lose everything after
	// the first dot.
	art, err := r.Render("v2.1.0-rc.1", RenderOptions{
		Font:     "ANSI Regular",
		MaxWidth: 30,
		Profile:  terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if art.Shortened {
		t.Fatalf("literal text was shortened at a dot: %q", art.Text)
	}
	if !strings.Contains(terminal.Strip(art.Text), "v2.1") {
		t.Fatalf("literal text lost its dotted content: %q", art.Text)
	}
	if art.Width > 30 {
		t.Fatalf("literal text overflows: %d columns", art.Width)
	}
}

func TestRendererFallsBackToNarrowFont(t *testing.T) {
	r := mustRenderer(t)
	// Wide enough for the name in Small but not in ANSI Regular.
	art, err := r.Render("media-server", RenderOptions{
		Font:      "ANSI Regular",
		Uppercase: true,
		MaxWidth:  70,
		Profile:   terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if art.Font == "ANSI Regular" {
		t.Fatalf("expected a narrower font at width 70, art is %d columns", art.Width)
	}
	if art.Width > 70 {
		t.Fatalf("fallback art still overflows: %d columns", art.Width)
	}
}

func TestRendererPlainHeaderAtTinyWidth(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("media-server-vault-01", RenderOptions{
		Font:      "ANSI Regular",
		Uppercase: true,
		MaxWidth:  30,
		Profile:   terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if art.Font != plainFont {
		t.Fatalf("expected plain header at width 30, got font %q", art.Font)
	}
	if strings.Contains(art.Text, "\n") {
		t.Fatalf("plain header must be a single line, got %q", art.Text)
	}
	if art.Width > 30 {
		t.Fatalf("plain header overflows: %d columns", art.Width)
	}
	if !strings.Contains(art.Text, "MEDIA-SERVER-VAULT-01") {
		t.Fatalf("plain header lost the hostname: %q", art.Text)
	}
}

func TestRendererUnconstrainedKeepsConfiguredFont(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("media-server-vault-01", RenderOptions{
		Font:       "ANSI Regular",
		Uppercase:  true,
		Monochrome: true,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if art.Font != "ANSI Regular" {
		t.Fatalf("unconstrained render should keep the configured font, got %q", art.Font)
	}
	if art.Shortened {
		t.Fatalf("unconstrained render should not shorten the name")
	}
}

func TestRendererNoColorProfileEmitsNoEscapes(t *testing.T) {
	r := mustRenderer(t)
	art, err := r.Render("host", RenderOptions{
		Font:     "standard",
		Gradient: []string{"brightblue", "blue", "cyan"},
		Profile:  terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(art.Text, "\033[") {
		t.Fatalf("NO_COLOR output contains escape sequences: %q", art.Text)
	}
}

func TestRendererGradientProfiles(t *testing.T) {
	r := mustRenderer(t)
	gradient := []string{"brightblue", "blue", "cyan", "brightcyan", "white"}

	ansi, err := r.Render("host", RenderOptions{Font: "standard", Gradient: gradient, Profile: terminal.ProfileANSI})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(ansi.Text, "\033[94m") {
		t.Errorf("ANSI gradient missing 16-color escape: %q", ansi.Text)
	}
	if strings.Contains(ansi.Text, "\033[38;2;") {
		t.Errorf("ANSI profile must not emit truecolor escapes")
	}

	tc, err := r.Render("host", RenderOptions{Font: "standard", Gradient: gradient, Profile: terminal.ProfileTrueColor})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(tc.Text, "\033[38;2;") {
		t.Errorf("truecolor gradient missing 24-bit escape: %q", tc.Text)
	}
}

func TestShortName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"pve1.home.lan", "pve1"},
		{"plain-host", "plain-host"},
		{".weird", ".weird"},
		{"host.", "host"},
	}
	for _, tt := range tests {
		if got := shortName(tt.in); got != tt.want {
			t.Errorf("shortName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLerpStops(t *testing.T) {
	stops := []string{"blue", "white"}
	start, ok := lerpStops(stops, 0)
	if !ok {
		t.Fatal("lerpStops failed at t=0")
	}
	blue, _ := terminal.RGB("blue")
	if start != blue {
		t.Errorf("t=0 should equal first stop, got %v want %v", start, blue)
	}
	end, ok := lerpStops(stops, 1)
	if !ok {
		t.Fatal("lerpStops failed at t=1")
	}
	white, _ := terminal.RGB("white")
	if end != white {
		t.Errorf("t=1 should equal last stop, got %v want %v", end, white)
	}
}

func TestRendererNonASCIITextUsesPlainHeader(t *testing.T) {
	r := mustRenderer(t)
	// go-figure's strict mode log.Fatals on non-ASCII, and non-strict mode
	// silently drops characters; both are wrong. Non-ASCII text must render
	// via the plain header with its content intact.
	for _, text := range []string{"日本語のホスト", "web-サーバ", "café-host"} {
		art, err := r.Render(text, RenderOptions{
			Font:     "ANSI Regular",
			MaxWidth: 40,
			Profile:  terminal.ProfileNoColor,
		})
		if err != nil {
			t.Fatalf("Render(%q) error = %v", text, err)
		}
		if art.Font != plainFont {
			t.Errorf("Render(%q): expected plain header, got font %q", text, art.Font)
		}
		if !strings.Contains(art.Text, text) {
			t.Errorf("Render(%q): text not preserved verbatim: %q", text, art.Text)
		}
		if art.Width > 40 {
			t.Errorf("Render(%q): overflows at %d columns", text, art.Width)
		}
	}
}

func TestPlainHeaderUsesAvailableWidth(t *testing.T) {
	r := mustRenderer(t)
	// 70 columns of hostname in a 100-column terminal: the ruled header
	// must widen to fit the name rather than clip it at the preferred
	// 60-column rule length.
	name := strings.Repeat("a", 70)
	art, err := r.Render(name+".home.lan", RenderOptions{
		Font:     "ANSI Regular",
		MaxWidth: 100,
		Profile:  terminal.ProfileNoColor,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(art.Text, name) {
		t.Fatalf("plain header clipped a name that fits the terminal: %q", art.Text)
	}
	if art.Width > 100 {
		t.Fatalf("plain header overflows: %d columns", art.Width)
	}
}
