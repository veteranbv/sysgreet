package windows

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/render"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

func TestSessionCollectorFallsBackToSSHClient(t *testing.T) {
	t.Setenv("SSH_CONNECTION", "")
	t.Setenv("SSH_CLIENT", "198.51.100.9 2222 22")

	collector := collectors.NewSessionCollector()
	info, err := collector.CollectSession(context.Background())
	if err != nil {
		t.Fatalf("CollectSession error: %v", err)
	}
	if info.RemoteAddr != "198.51.100.9" {
		t.Fatalf("expected remote addr 198.51.100.9, got %s", info.RemoteAddr)
	}
	if info.Source != "SSH_CLIENT" {
		t.Fatalf("expected source SSH_CLIENT, got %s", info.Source)
	}
}

func TestSessionCollectorNoEnvReturnsEmpty(t *testing.T) {
	t.Setenv("SSH_CONNECTION", "")
	t.Setenv("SSH_CLIENT", "")

	collector := collectors.NewSessionCollector()
	info, err := collector.CollectSession(context.Background())
	if err != nil {
		t.Fatalf("CollectSession error: %v", err)
	}
	if info.RemoteAddr != "" {
		t.Fatalf("expected empty remote addr, got %s", info.RemoteAddr)
	}
	if info.Source != "" {
		t.Fatalf("expected empty source, got %s", info.Source)
	}
}

func TestResourceCollectorUsesCPUUsageOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific test")
	}
	collector := collectors.NewResourceCollector()
	info, err := collector.CollectResources(context.Background())
	if err != nil {
		t.Fatalf("CollectResources error: %v", err)
	}
	if info.CPU.Mode != "usage" {
		t.Fatalf("expected CPU mode usage, got %s", info.CPU.Mode)
	}
}

func TestTOMLConfigOverridesFontAndSections(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific test")
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `display.user = false
display.memory = false
display.disk = false
display.load = false
ascii.font = "slant"
ascii.monochrome = true
layout.sections = ["network", "system"]
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("SYSGREET_CONFIG", cfgPath)
	cfg, _, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	renderer, err := ascii.NewRenderer()
	if err != nil {
		t.Fatalf("ascii renderer: %v", err)
	}
	providers := collectors.Providers{
		System:    stubSystemCollector{},
		Network:   stubNetworkCollector{},
		Resources: stubResourceCollector{},
		Session:   stubSessionCollector{},
	}
	builders := []banner.Builder{
		banner.SystemSectionBuilder{},
		banner.NetworkSectionBuilder{},
	}
	b, err := banner.New(providers, renderer, builders)
	if err != nil {
		t.Fatalf("banner.New: %v", err)
	}
	out, _, err := b.Build(context.Background(), cfg, terminal.Env{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	result := render.NewRenderer(terminal.Env{}).Render(out, cfg)
	if strings.Contains(result, "Mem:") {
		t.Fatalf("expected memory data suppressed, got %s", result)
	}
	if !strings.Contains(result, "Network") {
		t.Fatalf("expected network section present")
	}
}

type stubSystemCollector struct{}

type stubNetworkCollector struct{}

type stubResourceCollector struct{}

type stubSessionCollector struct{}

func (stubSystemCollector) CollectSystem(ctx context.Context) (collectors.SystemInfo, error) {
	return collectors.SystemInfo{Hostname: "stub", OS: "Windows", Arch: "amd64"}, nil
}

func (stubNetworkCollector) CollectNetwork(ctx context.Context) (collectors.NetworkInfo, error) {
	return collectors.NetworkInfo{Primary: &collectors.Address{IP: "172.16.0.5", Interface: "Ethernet0"}}, nil
}

func (stubResourceCollector) CollectResources(ctx context.Context) (collectors.ResourceInfo, error) {
	return collectors.ResourceInfo{}, nil
}

func (stubSessionCollector) CollectSession(ctx context.Context) (collectors.SessionInfo, error) {
	return collectors.SessionInfo{}, nil
}
