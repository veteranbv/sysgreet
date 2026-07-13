package linux

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/render"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

func TestYAMLConfigDisablesNetworkSections(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `display:
  memory: false
  disk: false
  load: false
  uptime: false
  user: false
  datetime: false
  last_login: false
layout:
  sections: ["network"]
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("SYSGREET_CONFIG", cfgPath)

	cfg, _, err := config.Load()
	if err != nil {
		t.Fatalf("Load config: %v", err)
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
		banner.ResourceSectionBuilder{},
	}

	b, err := banner.New(providers, renderer, builders)
	if err != nil {
		t.Fatalf("banner.New: %v", err)
	}
	out, _, err := b.Build(context.Background(), cfg, terminal.Env{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	txt := render.NewRenderer(terminal.Env{}).Render(out, cfg)
	if strings.Contains(txt, "System") {
		t.Fatalf("expected system section to be disabled, got %s", txt)
	}
	if !strings.Contains(txt, "Network") {
		t.Fatalf("expected network section present")
	}
	if strings.Contains(txt, "Mem:") {
		t.Fatalf("expected memory section disabled")
	}
}

type stubSystemCollector struct{}

type stubNetworkCollector struct{}

type stubResourceCollector struct{}

type stubSessionCollector struct{}

func (stubSystemCollector) CollectSystem(ctx context.Context) (collectors.SystemInfo, error) {
	return collectors.SystemInfo{Hostname: "stub", OS: "Linux", Arch: "amd64"}, nil
}

func (stubNetworkCollector) CollectNetwork(ctx context.Context) (collectors.NetworkInfo, error) {
	return collectors.NetworkInfo{Primary: &collectors.Address{IP: "10.0.0.1", Interface: "eth0"}}, nil
}

func (stubResourceCollector) CollectResources(ctx context.Context) (collectors.ResourceInfo, error) {
	return collectors.ResourceInfo{}, nil
}

func (stubSessionCollector) CollectSession(ctx context.Context) (collectors.SessionInfo, error) {
	return collectors.SessionInfo{}, nil
}
