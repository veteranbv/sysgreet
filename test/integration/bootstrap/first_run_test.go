package bootstrap_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sysgreet")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../../cmd/sysgreet")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, output)
	}
	return binaryPath
}

type configDoc struct {
	ASCII struct {
		Font string `yaml:"font"`
	} `yaml:"ascii"`
	Version   string `yaml:"version"`
	CreatedAt string `yaml:"created_at"`
}

// runSysgreet runs the binary with a clean environment rooted at home, so
// nothing from the test runner's own SSH session or config leaks in.
func runSysgreet(t *testing.T, bin, home string, stdin string, env []string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + home}, env...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	timer := time.AfterFunc(10*time.Second, func() { _ = cmd.Process.Kill() })
	defer timer.Stop()
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestNormalRunHasNoSideEffects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bootstrap integration in short mode")
	}

	bin := buildBinary(t)
	home := t.TempDir()
	cfgPath := filepath.Join(home, ".config", "sysgreet", "config.yaml")

	stdout, stderr, err := runSysgreet(t, bin, home, "", []string{"SYSGREET_ASSUME_TTY=1"})
	if err != nil {
		t.Fatalf("banner run failed: %v\nstderr: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected a banner on stdout")
	}
	if stderr != "" {
		t.Fatalf("a normal login must not write to stderr, got: %s", stderr)
	}
	if _, err := os.Stat(cfgPath); err == nil {
		t.Fatalf("a normal run must not create %s", cfgPath)
	}
}

func TestInitConfigCreatesConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bootstrap integration in short mode")
	}

	bin := buildBinary(t)
	home := t.TempDir()
	cfgPath := filepath.Join(home, ".config", "sysgreet", "config.yaml")

	stdout, stderr, err := runSysgreet(t, bin, home, "", nil, "--init-config")
	if err != nil {
		t.Fatalf("--init-config failed: %v\nstderr: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("--init-config should not print a banner, got: %s", stdout)
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
	var parsed configDoc
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	if parsed.ASCII.Font != "ANSI Regular" {
		t.Fatalf("expected ascii font ANSI Regular, got %q", parsed.ASCII.Font)
	}
	if parsed.Version == "" || parsed.CreatedAt == "" {
		t.Fatalf("expected version and created_at metadata, got %+v", parsed)
	}
	if !strings.Contains(stderr, "created default config") {
		t.Fatalf("expected stderr to report config creation, got: %s", stderr)
	}
}
