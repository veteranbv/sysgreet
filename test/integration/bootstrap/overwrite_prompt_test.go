package bootstrap_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOverwritePromptFlows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bootstrap prompt integration in short mode")
	}

	binaryPath := buildBinary(t)
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("ascii:\n  font: standard\n"), 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	run := func(input string) (stdout, stderr string) {
		cmd := exec.Command(binaryPath, "--init-config")
		cmd.Env = append(os.Environ(), "SYSGREET_CONFIG="+cfgPath, "SYSGREET_ASSUME_TTY=1", "SYSGREET_CONFIG_POLICY=prompt")
		cmd.Stdin = strings.NewReader(input)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("run sysgreet: %v\nstdout: %s\nstderr: %s", err, outBuf.String(), errBuf.String())
		}
		return outBuf.String(), errBuf.String()
	}

	// Keep existing config
	_, stderr := run("k\n")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(content), "standard") {
		t.Fatalf("expected existing font to remain, got %s", content)
	}
	if strings.Contains(stderr, "overwrote") {
		t.Fatalf("unexpected overwrite message: %s", stderr)
	}
	matches, err := filepath.Glob(filepath.Join(cfgDir, "config.yaml.bak-*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no backup files after keep, found %v", matches)
	}

	// Overwrite config
	prevContent := string(content)
	_, stderr = run("o\n")
	content, err = os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(content), "ANSI Regular") {
		t.Fatalf("expected overwrite to set ANSI Regular font, got %s", content)
	}
	matches, err = filepath.Glob(filepath.Join(cfgDir, "config.yaml.bak-*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one backup file, found %v", matches)
	}
	backupData, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupData) != prevContent {
		t.Fatalf("backup does not match previous content\nwant: %s\n got: %s", prevContent, string(backupData))
	}
	if !strings.Contains(stderr, "overwrote") {
		t.Fatalf("expected overwrite message in stderr, got %s", stderr)
	}
}
