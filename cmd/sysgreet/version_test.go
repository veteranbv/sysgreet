package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildInfo(t *testing.T) {
	tests := []struct {
		name    string
		info    *debug.BuildInfo
		v, c, d string
		wantV   string
		wantC   string
		wantD   string
	}{
		{
			name: "go install binary resolves module version",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.0"},
			},
			v: "dev", c: "none", d: "unknown",
			wantV: "v1.2.0", wantC: "none", wantD: "unknown",
		},
		{
			name: "local git build resolves vcs metadata",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234"},
					{Key: "vcs.time", Value: "2026-07-15T00:00:00Z"},
				},
			},
			v: "dev", c: "none", d: "unknown",
			wantV: "dev", wantC: "abc1234", wantD: "2026-07-15T00:00:00Z",
		},
		{
			name: "goreleaser ldflags win over build info",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.0"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234"},
				},
			},
			v: "v1.2.0", c: "deadbeef", d: "2026-07-15",
			wantV: "v1.2.0", wantC: "deadbeef", wantD: "2026-07-15",
		},
		{
			name: "dirty checkout marks the commit",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			v: "dev", c: "none", d: "unknown",
			wantV: "dev", wantC: "abc1234-dirty", wantD: "unknown",
		},
		{
			name: "ldflags commit is never marked dirty",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			v: "v1.2.0", c: "deadbeef", d: "2026-07-15",
			wantV: "v1.2.0", wantC: "deadbeef", wantD: "2026-07-15",
		},
		{
			name: "empty build info keeps defaults",
			info: &debug.BuildInfo{},
			v:    "dev", c: "none", d: "unknown",
			wantV: "dev", wantC: "none", wantD: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, gotC, gotD := resolveBuildInfo(tt.info, tt.v, tt.c, tt.d)
			if gotV != tt.wantV || gotC != tt.wantC || gotD != tt.wantD {
				t.Errorf("resolveBuildInfo() = (%q, %q, %q), want (%q, %q, %q)",
					gotV, gotC, gotD, tt.wantV, tt.wantC, tt.wantD)
			}
		})
	}
}
