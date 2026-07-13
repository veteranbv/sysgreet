package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBinaryExecution(t *testing.T) {
	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sysgreet")
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/sysgreet")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, buildOutput)
	}

	tests := []struct {
		name         string
		args         []string
		env          map[string]string
		wantErr      bool
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:      "disable flag produces no output",
			args:      []string{"--disable"},
			wantEmpty: true,
			wantErr:   false,
		},
		{
			name:         "default execution produces output",
			args:         []string{},
			wantErr:      false,
			wantContains: []string{},
			wantEmpty:    false,
		},
		{
			name: "monochrome mode",
			args: []string{},
			env: map[string]string{
				"SYSGREET_ASCII_MONOCHROME": "true",
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name: "disable specific sections",
			args: []string{},
			env: map[string]string{
				"SYSGREET_DISPLAY_UPTIME": "false",
				"SYSGREET_DISPLAY_MEMORY": "false",
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name: "compact layout",
			args: []string{},
			env: map[string]string{
				"SYSGREET_LAYOUT_COMPACT": "true",
			},
			wantErr:      false,
			wantContains: []string{"|"}, // Compact mode uses pipe separator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)

			// Use temp config to avoid prompts
			testConfigPath := filepath.Join(tmpDir, "test-config.yaml")

			// Set environment variables
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "SYSGREET_CONFIG="+testConfigPath, "CI=1", "SYSGREET_CONFIG_POLICY=overwrite")
			for k, v := range tt.env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			// Set timeout to prevent hanging
			timeout := time.Second * 5
			timer := time.AfterFunc(timeout, func() {
				if cmd.Process != nil {
					_ = cmd.Process.Kill() // Best effort kill on timeout
				}
			})
			defer timer.Stop()

			output, err := cmd.CombinedOutput()

			if (err != nil) != tt.wantErr {
				t.Errorf("execution error = %v, wantErr %v\nOutput: %s", err, tt.wantErr, output)
				return
			}

			outputStr := string(output)

			if tt.wantEmpty {
				if len(strings.TrimSpace(outputStr)) > 0 {
					t.Errorf("expected empty output, got: %s", outputStr)
				}
				return
			}

			if !tt.wantEmpty && len(strings.TrimSpace(outputStr)) == 0 {
				t.Error("expected non-empty output, got empty")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing %q\nGot: %s", want, outputStr)
				}
			}
		})
	}
}

func TestBinaryStartupTime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sysgreet")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/sysgreet")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	// Run multiple iterations to get average startup time
	iterations := 10
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		cmd := exec.Command(binaryPath, "--disable")
		if err := cmd.Run(); err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}
		duration := time.Since(start)
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(iterations)
	maxAllowed := 80 * time.Millisecond

	if avgDuration > maxAllowed {
		t.Errorf("average startup time %v exceeds maximum allowed %v", avgDuration, maxAllowed)
	}

	t.Logf("Average startup time over %d iterations: %v", iterations, avgDuration)
}

func TestBinaryWithInvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sysgreet")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/sysgreet")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	// Create invalid config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	invalidConfig := "invalid: yaml: content: [[[["
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run with invalid config and "keep" policy - should fail to load the invalid YAML
	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(), "SYSGREET_CONFIG="+configPath, "CI=1", "SYSGREET_CONFIG_POLICY=keep")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected error with invalid config, got success\nOutput: %s", output)
	}

	// Should contain error message
	if !strings.Contains(string(output), "sysgreet:") {
		t.Errorf("error output should contain 'sysgreet:' prefix\nGot: %s", output)
	}
}

func TestBinaryMemoryFootprint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// This test is informational and doesn't fail
	// It helps track memory usage over time
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sysgreet")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/sysgreet")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	// Get binary size
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("failed to stat binary: %v", err)
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	maxSizeMB := 10.0

	if sizeMB > maxSizeMB {
		t.Errorf("binary size %.2f MB exceeds maximum %.2f MB", sizeMB, maxSizeMB)
	}

	t.Logf("Binary size: %.2f MB", sizeMB)
}

// buildTestBinary compiles sysgreet into dir and returns its path.
func buildTestBinary(t *testing.T, dir string) string {
	t.Helper()
	binaryPath := filepath.Join(dir, "sysgreet")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/sysgreet")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, output)
	}
	return binaryPath
}

func TestBinaryJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	tmpDir := t.TempDir()
	binaryPath := buildTestBinary(t, tmpDir)

	cmd := exec.Command(binaryPath, "--json")
	cmd.Env = append(os.Environ(), "SYSGREET_CONFIG="+filepath.Join(tmpDir, "config.yaml"), "CI=1")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("--json run failed: %v", err)
	}

	var doc struct {
		Hostname string `json:"hostname"`
		Sections []struct {
			Key   string   `json:"key"`
			Lines []string `json:"lines"`
		} `json:"sections"`
	}
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\nOutput: %s", err, output)
	}
	if doc.Hostname == "" {
		t.Errorf("JSON output missing hostname: %s", output)
	}
	if strings.Contains(string(output), "\033") {
		t.Errorf("JSON output contains escape sequences")
	}
}

func TestBinaryWidthNeverOverflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	tmpDir := t.TempDir()
	binaryPath := buildTestBinary(t, tmpDir)

	for _, width := range []string{"30", "50", "80"} {
		cmd := exec.Command(binaryPath, "--text", "media-server-vault-01", "--width", width)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("--width %s run failed: %v", width, err)
		}
		limit, _ := strconv.Atoi(width)
		for _, line := range strings.Split(string(output), "\n") {
			if n := len([]rune(line)); n > limit {
				t.Errorf("--width %s: line has %d columns: %q", width, n, line)
			}
		}
	}
}

func TestBinaryListFonts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	tmpDir := t.TempDir()
	binaryPath := buildTestBinary(t, tmpDir)

	output, err := exec.Command(binaryPath, "--list-fonts").Output()
	if err != nil {
		t.Fatalf("--list-fonts run failed: %v", err)
	}
	for _, want := range []string{"ANSI Regular", "standard", "Small"} {
		if !strings.Contains(string(output), want) {
			t.Errorf("--list-fonts missing %q\nGot: %s", want, output)
		}
	}
}

func TestBinaryTextModeHonorsConfigFont(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	tmpDir := t.TempDir()
	binaryPath := buildTestBinary(t, tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	config := "ascii:\n  font: \"standard\"\n  monochrome: true\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd := exec.Command(binaryPath, "--text", "hi")
	cmd.Env = append(os.Environ(), "SYSGREET_CONFIG="+configPath, "CI=1")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("--text run failed: %v", err)
	}
	// The standard font draws with slashes and underscores, not the solid
	// blocks of the default ANSI Regular.
	if strings.Contains(string(output), "█") {
		t.Errorf("--text ignored the configured font\nGot: %s", output)
	}
	if !strings.Contains(string(output), "_") {
		t.Errorf("expected standard-font glyphs in output\nGot: %s", output)
	}
}

func TestBinaryNoColorFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	tmpDir := t.TempDir()
	binaryPath := buildTestBinary(t, tmpDir)

	cmd := exec.Command(binaryPath, "--demo", "--no-color")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("--no-color run failed: %v", err)
	}
	if strings.Contains(string(output), "\033") {
		t.Errorf("--no-color output contains escape sequences")
	}
}
