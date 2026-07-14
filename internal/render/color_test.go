package render

import (
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/terminal"
)

func TestColorizer_Wrap(t *testing.T) {
	tests := []struct {
		name         string
		profile      terminal.Profile
		color        string
		text         string
		wantContains []string
		wantEqual    bool
	}{
		{
			name:         "red color enabled",
			profile:      terminal.ProfileANSI,
			color:        "red",
			text:         "ERROR",
			wantContains: []string{"\033[31m", "ERROR", "\033[0m"},
		},
		{
			name:         "yellow color enabled",
			profile:      terminal.ProfileANSI,
			color:        "yellow",
			text:         "WARNING",
			wantContains: []string{"\033[33m", "WARNING", "\033[0m"},
		},
		{
			name:         "green color enabled",
			profile:      terminal.ProfileANSI,
			color:        "green",
			text:         "OK",
			wantContains: []string{"\033[32m", "OK", "\033[0m"},
		},
		{
			name:         "cyan color enabled",
			profile:      terminal.ProfileANSI,
			color:        "cyan",
			text:         "INFO",
			wantContains: []string{"\033[36m", "INFO", "\033[0m"},
		},
		{
			name:      "no-color profile",
			profile:   terminal.ProfileNoColor,
			color:     "red",
			text:      "ERROR",
			wantEqual: true,
		},
		{
			name:      "unknown color is passthrough",
			profile:   terminal.ProfileANSI,
			color:     "magenta",
			text:      "TEXT",
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewColorizer(tt.profile)
			result := c.Wrap(tt.color, tt.text)

			if tt.wantEqual {
				if result != tt.text {
					t.Errorf("Wrap(%q, %q) = %q, want %q", tt.color, tt.text, result, tt.text)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Wrap(%q, %q) = %q, want to contain %q", tt.color, tt.text, result, want)
				}
			}
		})
	}
}

func TestStrip(t *testing.T) {
	input := "\033[31mRED\033[0m and \033[32mGREEN\033[0m"
	if got := Strip(input); got != "RED and GREEN" {
		t.Errorf("Strip(%q) = %q", input, got)
	}
}
