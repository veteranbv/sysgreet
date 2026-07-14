package terminal

import (
	"os"
	"path/filepath"
	"testing"
)

// pipeFile returns a non-terminal file for detection tests.
func pipeFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestDetectProfileExplicitDisable(t *testing.T) {
	if got := DetectProfile(pipeFile(t), true); got != ProfileNoColor {
		t.Errorf("explicit disable: got %v, want ProfileNoColor", got)
	}
}

func TestDetectProfileNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := DetectProfile(pipeFile(t), false); got != ProfileNoColor {
		t.Errorf("NO_COLOR set: got %v, want ProfileNoColor", got)
	}
}

func TestDetectProfileDumbTerm(t *testing.T) {
	t.Setenv("TERM", "dumb")
	if got := DetectProfile(pipeFile(t), false); got != ProfileNoColor {
		t.Errorf("TERM=dumb: got %v, want ProfileNoColor", got)
	}
}

func TestDetectProfileNonTerminal(t *testing.T) {
	// A regular file is not a terminal, so color must be off even with a
	// color-friendly environment.
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "truecolor")
	if got := DetectProfile(pipeFile(t), false); got != ProfileNoColor {
		t.Errorf("non-terminal output: got %v, want ProfileNoColor", got)
	}
}

func TestWidthFallsBackToColumnsEnv(t *testing.T) {
	t.Setenv("COLUMNS", "72")
	if got := Width(pipeFile(t)); got != 72 {
		t.Errorf("COLUMNS fallback: got %d, want 72", got)
	}
}

func TestWidthUnknown(t *testing.T) {
	t.Setenv("COLUMNS", "")
	if got := Width(pipeFile(t)); got != 0 {
		t.Errorf("unknown width: got %d, want 0", got)
	}
}

func TestWidthIgnoresGarbageColumns(t *testing.T) {
	t.Setenv("COLUMNS", "not-a-number")
	if got := Width(pipeFile(t)); got != 0 {
		t.Errorf("garbage COLUMNS: got %d, want 0", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		color   string
		text    string
		want    string
	}{
		{"ansi red", ProfileANSI, "red", "ERROR", "\033[31mERROR\033[0m"},
		{"truecolor uses named code", ProfileTrueColor, "cyan", "X", "\033[36mX\033[0m"},
		{"no color passthrough", ProfileNoColor, "red", "ERROR", "ERROR"},
		{"unknown color passthrough", ProfileANSI, "magenta", "TEXT", "TEXT"},
		{"case insensitive", ProfileANSI, "RED", "E", "\033[31mE\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.profile, tt.color, tt.text); got != tt.want {
				t.Errorf("Wrap(%v, %q, %q) = %q, want %q", tt.profile, tt.color, tt.text, got, tt.want)
			}
		})
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"red code", "\033[31mERROR\033[0m", "ERROR"},
		{"multiple codes", "\033[31mRED\033[0m and \033[32mGREEN\033[0m", "RED and GREEN"},
		{"truecolor code", "\033[38;2;36;114;200mBLUE\033[0m", "BLUE"},
		{"no codes", "plain text", "plain text"},
		{"empty", "", ""},
		{"compound sequence", "\033[1;31mBOLD RED\033[0m", "BOLD RED"},
		{"only codes", "\033[31m\033[0m", ""},
		{"unicode preserved", "Hello 世界 🌍", "Hello 世界 🌍"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Strip(tt.input); got != tt.want {
				t.Errorf("Strip(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestForegroundRGB(t *testing.T) {
	if got := ForegroundRGB(36, 114, 200); got != "\033[38;2;36;114;200m" {
		t.Errorf("ForegroundRGB = %q", got)
	}
}

func TestCodeAndRGBCoverSameNames(t *testing.T) {
	for name := range ansiCodes {
		if _, ok := rgbValues[name]; !ok {
			t.Errorf("color %q has an ANSI code but no RGB value", name)
		}
	}
	for name := range rgbValues {
		if _, ok := ansiCodes[name]; !ok {
			t.Errorf("color %q has an RGB value but no ANSI code", name)
		}
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"hello", 5},
		{"", 0},
		{"日本語", 6},
		{"héllo", 5},
		{"user-日本", 9},
	}
	for _, tt := range tests {
		if got := DisplayWidth(tt.in); got != tt.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestClip(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, "hello"}, // zero width = unconstrained
		{"日本語です", 4, "日…"},
		{"日本語です", 5, "日本…"},
	}
	for _, tt := range tests {
		got := Clip(tt.in, tt.width)
		if got != tt.want {
			t.Errorf("Clip(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
		}
		if tt.width > 0 && DisplayWidth(got) > tt.width {
			t.Errorf("Clip(%q, %d) = %q occupies %d columns", tt.in, tt.width, got, DisplayWidth(got))
		}
	}
}
