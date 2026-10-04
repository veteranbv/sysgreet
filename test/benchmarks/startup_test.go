package benchmarks

import (
	"context"
	"testing"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/render"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

// BenchmarkStartup measures everything a login pays for after process start:
// real collectors, art, layout, and rendering at a typical terminal width.
func BenchmarkStartup(b *testing.B) {
	cfg := config.Default()
	cfg.ASCII.Font = "ANSI Regular"
	env := render.ApplyConfig(terminal.Env{Width: 120, Profile: terminal.ProfileANSI}, cfg)
	art := newArt(b)
	providers := collectors.Providers{
		System:    collectors.NewSystemCollector(),
		Network:   collectors.NewNetworkCollector(cfg.Network.MaxInterfaces),
		Resources: collectors.NewResourceCollector(),
		Session:   collectors.NewSessionCollector(),
		LastLogin: collectors.NewLastLoginCollector(),
	}
	hostBanner, err := banner.New(providers, art, banner.BuildersForConfig(cfg))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	for b.Loop() {
		out, _, err := hostBanner.Build(ctx, cfg, env)
		if err != nil {
			b.Fatal(err)
		}
		if render.NewRenderer(env).Render(out, cfg) == "" {
			b.Fatal("empty banner")
		}
	}
}

// BenchmarkRender isolates the CPU side (art and layout) from the host by
// rendering the demo snapshot.
func BenchmarkRender(b *testing.B) {
	cfg := config.Default()
	cfg.ASCII.Font = "ANSI Regular"
	env := render.ApplyConfig(terminal.Env{Width: 120, Profile: terminal.ProfileANSI}, cfg)
	hostBanner, err := banner.New(collectors.Providers{}, newArt(b), banner.BuildersForConfig(cfg))
	if err != nil {
		b.Fatal(err)
	}
	snap := collectors.DemoSnapshot()

	for b.Loop() {
		out := hostBanner.BuildWithSnapshot(snap, cfg, env)
		if render.NewRenderer(env).Render(out, cfg) == "" {
			b.Fatal("empty banner")
		}
	}
}

func newArt(b *testing.B) *ascii.Renderer {
	b.Helper()
	r, err := ascii.NewRenderer()
	if err != nil {
		b.Fatal(err)
	}
	return r
}
