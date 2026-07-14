package config

import (
	os "os"
	"path/filepath"
	testing "testing"
)

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, path, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if path != "" {
		t.Fatalf("expected no config path, got %q", path)
	}
	def := Default()
	if cfg.Layout.Compact != def.Layout.Compact || len(cfg.Layout.Sections) != len(def.Layout.Sections) {
		t.Fatalf("expected defaults, got %+v", cfg)
	}
	if cfg.ASCII.Font != "ANSI Regular" {
		t.Fatalf("expected ANSI Regular font by default, got %s", cfg.ASCII.Font)
	}
	if cfg.Version != SchemaVersion {
		t.Fatalf("expected schema version %s, got %s", SchemaVersion, cfg.Version)
	}
	if cfg.CreatedAt != "" {
		t.Fatalf("expected empty created_at for in-memory defaults, got %q", cfg.CreatedAt)
	}
}

func TestLoad_YAMLOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := dir + "/.config/sysgreet"
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte(`display:
  hostname: false
ascii:
  font: "standard"
network:
  max_interfaces: 1
version: v42
created_at: 2024-01-01T00:00:00Z
`)
	if err := os.WriteFile(cfgDir+"/config.yaml", content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, used, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if used == "" {
		t.Fatalf("expected config path to be returned")
	}
	if cfg.Display.Hostname {
		t.Fatalf("expected hostname disabled")
	}
	if cfg.ASCII.Font != "standard" {
		t.Fatalf("expected font override, got %s", cfg.ASCII.Font)
	}
	if cfg.Network.MaxInterfaces != 1 {
		t.Fatalf("expected max interfaces 1, got %d", cfg.Network.MaxInterfaces)
	}
	if cfg.Version != "v42" {
		t.Fatalf("expected version v42, got %s", cfg.Version)
	}
	if cfg.CreatedAt != "2024-01-01T00:00:00Z" {
		t.Fatalf("expected created_at preserved, got %s", cfg.CreatedAt)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SYSGREET_DISPLAY_REMOTE_IP", "false")
	t.Setenv("SYSGREET_LAYOUT_SECTIONS", "header,resources")
	t.Setenv("SYSGREET_NETWORK_MAX_INTERFACES", "5")

	cfg, _, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Display.RemoteIP {
		t.Fatalf("env override for remote ip failed")
	}
	if len(cfg.Layout.Sections) != 2 || cfg.Layout.Sections[1] != "resources" {
		t.Fatalf("layout sections override failed: %+v", cfg.Layout.Sections)
	}
	if cfg.Network.MaxInterfaces != 5 {
		t.Fatalf("expected max interfaces 5, got %d", cfg.Network.MaxInterfaces)
	}
}

func TestDefaultSectionsOrder(t *testing.T) {
	cfg := Default()
	if len(cfg.Layout.Sections) == 0 {
		t.Fatalf("expected default sections")
	}
	if cfg.Layout.Sections[0] != "header" {
		t.Fatalf("expected header first, got %s", cfg.Layout.Sections[0])
	}
}

func TestDefaultWritePathUsesEnv(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.yaml")
	t.Setenv("SYSGREET_CONFIG", custom)
	got := DefaultWritePath()
	if got != custom {
		t.Fatalf("expected %s, got %s", custom, got)
	}
}

func TestDefaultWritePathFallsBackToHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SYSGREET_CONFIG", "")
	got := DefaultWritePath()
	expected := filepath.Join(dir, ".config", "sysgreet", "config.yaml")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestLoad_MaxWidthFromYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := dir + "/.config/sysgreet"
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte(`layout:
  max_width: 100
`)
	if err := os.WriteFile(cfgDir+"/config.yaml", content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, _, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Layout.MaxWidth != 100 {
		t.Fatalf("expected max_width 100, got %d", cfg.Layout.MaxWidth)
	}
}

func TestLoad_MaxWidthFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SYSGREET_LAYOUT_MAX_WIDTH", "90")

	cfg, _, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Layout.MaxWidth != 90 {
		t.Fatalf("expected max_width 90 from env, got %d", cfg.Layout.MaxWidth)
	}
}

func TestLoad_MaxWidthRejectsNegative(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SYSGREET_LAYOUT_MAX_WIDTH", "-5")

	cfg, _, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Layout.MaxWidth != 0 {
		t.Fatalf("negative max_width should be ignored, got %d", cfg.Layout.MaxWidth)
	}
}

func TestLoad_ExplicitConfigPathIsExclusive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := dir + "/.config/sysgreet"
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A default-path config that must NOT be picked up when an explicit
	// path is requested but missing.
	decoy := []byte("ascii:\n  font: \"slant\"\n")
	if err := os.WriteFile(cfgDir+"/config.yaml", decoy, 0o644); err != nil {
		t.Fatalf("write decoy config: %v", err)
	}
	t.Setenv("SYSGREET_CONFIG", dir+"/does-not-exist.yaml")

	cfg, used, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if used != "" {
		t.Fatalf("no config should have been loaded, got %q", used)
	}
	if cfg.ASCII.Font != Default().ASCII.Font {
		t.Fatalf("expected built-in defaults, got font %q from decoy config", cfg.ASCII.Font)
	}
}
