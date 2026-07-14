package bootstrap

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/veteranbv/sysgreet/internal/config"
)

type configYAML struct {
	Display struct {
		Hostname    bool `yaml:"hostname"`
		OS          bool `yaml:"os"`
		IPAddresses bool `yaml:"ip_addresses"`
		RemoteIP    bool `yaml:"remote_ip"`
		Uptime      bool `yaml:"uptime"`
		User        bool `yaml:"user"`
		Memory      bool `yaml:"memory"`
		Disk        bool `yaml:"disk"`
		Load        bool `yaml:"load"`
		Datetime    bool `yaml:"datetime"`
		LastLogin   bool `yaml:"last_login"`
	} `yaml:"display"`
	ASCII struct {
		Font       string `yaml:"font"`
		Color      string `yaml:"color"`
		Monochrome bool   `yaml:"monochrome"`
	} `yaml:"ascii"`
	Layout struct {
		Compact  bool     `yaml:"compact"`
		Sections []string `yaml:"sections"`
	} `yaml:"layout"`
	Network struct {
		ShowInterfaceNames bool `yaml:"show_interface_names"`
		MaxInterfaces      int  `yaml:"max_interfaces"`
	} `yaml:"network"`
	Version   string `yaml:"version"`
	CreatedAt string `yaml:"created_at"`
}

func TestBootstrapCreatesDefaultConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	stderr := &bytes.Buffer{}
	result, err := Bootstrap(ctx, cfgPath, IO{Stderr: stderr}, Options{Interactive: true})
	if err != nil {
		t.Fatalf("Bootstrap error: %v", err)
	}
	if result.Action != ActionCreated {
		t.Fatalf("expected action %s, got %s", ActionCreated, result.Action)
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var cfg configYAML
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}

	if cfg.ASCII.Font != "ANSI Regular" {
		t.Fatalf("expected ascii.font=ANSI Regular, got %q", cfg.ASCII.Font)
	}

	if cfg.Version == "" {
		t.Fatalf("expected version field to be set")
	}
	if cfg.CreatedAt == "" {
		t.Fatalf("expected created_at field to be set")
	}
	if _, err := time.Parse(time.RFC3339, cfg.CreatedAt); err != nil {
		t.Fatalf("created_at not RFC3339: %v", err)
	}
}

func TestBootstrapWritesTOMLForTOMLPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	_, err := Bootstrap(context.Background(), cfgPath, IO{}, Options{FlagPolicy: "overwrite", Interactive: false})
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	t.Setenv("SYSGREET_CONFIG", cfgPath)
	cfg, used, err := config.Load()
	if err != nil {
		t.Fatalf("bootstrapped .toml config does not parse: %v", err)
	}
	if used != cfgPath {
		t.Fatalf("expected %s to be loaded, got %q", cfgPath, used)
	}
	if cfg.ASCII.Font != config.Default().ASCII.Font {
		t.Fatalf("bootstrapped config lost defaults: %+v", cfg.ASCII)
	}
}
