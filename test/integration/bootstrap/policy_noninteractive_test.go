package bootstrap_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A fresh host reached by `ssh host cmd`, cron, or Ansible has no config and
// no terminal. That must render the banner from defaults, not fail.
func TestFreshHostWithoutTerminalRendersBanner(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping policy integration in short mode")
	}

	bin := buildBinary(t)
	home := t.TempDir()

	stdout, stderr, err := runSysgreet(t, bin, home, "", []string{"CI=1"})
	if err != nil {
		t.Fatalf("expected success on a fresh host, got %v\nstderr: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected a banner rendered from defaults")
	}
	if stderr != "" {
		t.Fatalf("expected no stderr chatter, got: %s", stderr)
	}
}

func TestInitConfigPoliciesWithoutTerminal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping policy integration in short mode")
	}

	bin := buildBinary(t)
	home := t.TempDir()
	cfgPath := filepath.Join(home, ".config", "sysgreet", "config.yaml")

	// Missing config: created whatever the policy.
	if _, stderr, err := runSysgreet(t, bin, home, "", []string{"CI=1"}, "--init-config", "--config-policy=keep"); err != nil {
		t.Fatalf("init with keep failed: %v (%s)", err, stderr)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("expected config to be created: %v", err)
	}

	// Existing config, no policy, no terminal: left alone.
	custom := "ascii:\n  font: \"standard\"\n# MY CUSTOM\n"
	if err := os.WriteFile(cfgPath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, stderr, err := runSysgreet(t, bin, home, "", []string{"CI=1"}, "--init-config"); err != nil {
		t.Fatalf("init without policy failed: %v (%s)", err, stderr)
	}
	if got, _ := os.ReadFile(cfgPath); string(got) != custom {
		t.Fatalf("existing config changed without a policy:\n%s", got)
	}

	// Overwrite twice: every earlier version must survive as a backup.
	for i := 0; i < 2; i++ {
		if _, stderr, err := runSysgreet(t, bin, home, "", []string{"CI=1", "SYSGREET_CONFIG_POLICY=overwrite"}, "--init-config"); err != nil {
			t.Fatalf("overwrite %d failed: %v (%s)", i, err, stderr)
		}
	}
	backups, _ := filepath.Glob(cfgPath + ".bak-*")
	if len(backups) != 2 {
		t.Fatalf("expected 2 backups after 2 overwrites, got %v", backups)
	}
	found := false
	for _, b := range backups {
		if data, _ := os.ReadFile(b); strings.Contains(string(data), "MY CUSTOM") {
			found = true
		}
	}
	if !found {
		t.Fatal("the user's original config was lost: no backup contains it")
	}
}

// Shell rc files run for scp, rsync, and sftp too. In a non-interactive SSH
// session the banner must stay silent unless forced.
func TestNonInteractiveSSHSessionIsSilent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping policy integration in short mode")
	}

	bin := buildBinary(t)
	home := t.TempDir()
	ssh := []string{"SSH_CONNECTION=203.0.113.5 50000 192.0.2.10 22"}

	stdout, stderr, err := runSysgreet(t, bin, home, "", ssh)
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("expected silence in a non-interactive SSH session, got err=%v stdout=%q stderr=%q", err, stdout, stderr)
	}

	stdout, _, err = runSysgreet(t, bin, home, "", ssh, "--force")
	if err != nil || strings.TrimSpace(stdout) == "" {
		t.Fatalf("--force should print the banner, got err=%v stdout=%q", err, stdout)
	}

	stdout, _, err = runSysgreet(t, bin, home, "", ssh, "--json")
	if err != nil || !strings.Contains(stdout, `"hostname"`) {
		t.Fatalf("explicit --json must still print, got err=%v stdout=%q", err, stdout)
	}
}

func TestDisableEnvironmentVariable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping policy integration in short mode")
	}

	bin := buildBinary(t)
	stdout, stderr, err := runSysgreet(t, bin, t.TempDir(), "", []string{"SYSGREET_DISABLE=1", "SYSGREET_ASSUME_TTY=1"})
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("SYSGREET_DISABLE=1 must print nothing, got err=%v stdout=%q stderr=%q", err, stdout, stderr)
	}
}
