package collectors

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestResourceCollectorReturnsMetrics(t *testing.T) {
	collector := NewResourceCollector()
	info, err := collector.CollectResources(context.Background())
	if err != nil {
		t.Fatalf("CollectResources error: %v", err)
	}
	if info.Memory.Total == 0 {
		t.Fatalf("expected memory total > 0")
	}
	if info.Memory.Available == 0 {
		t.Fatalf("expected memory available > 0")
	}
	if info.Disk.Total == 0 {
		t.Fatalf("expected disk total > 0")
	}
	if info.Disk.Used == 0 {
		t.Fatalf("expected disk used > 0")
	}
	if info.CPU.Mode == "" {
		t.Fatalf("expected CPU mode set")
	}
}

// The gather tests run on synctest's virtual clock: the stubs' sleeps and
// Gather's deadline are exact, so the bounds below cannot flake on a slow
// runner.

func TestGatherRunsCollectorsConcurrently(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Each stub sleeps 30ms; serial execution would take 150ms.
		start := time.Now()
		providers := Providers{
			System:    slowSystemCollector{},
			Network:   slowNetworkCollector{},
			Resources: slowResourceCollector{},
			Session:   slowSessionCollector{},
			LastLogin: slowLastLoginCollector{},
		}
		snap := providers.Gather(context.Background())

		if elapsed := time.Since(start); elapsed != 30*time.Millisecond {
			t.Errorf("Gather took %v; concurrent collectors should finish together at 30ms", elapsed)
		}
		if snap.System.Hostname != "slow" {
			t.Errorf("system snapshot missing, got %+v", snap.System)
		}
		if snap.Network.Primary == nil {
			t.Error("network snapshot missing")
		}
		if snap.Session.RemoteAddr != "203.0.113.7" {
			t.Errorf("session snapshot missing, got %+v", snap.Session)
		}
		if snap.LastLogin == nil {
			t.Error("last login snapshot missing")
		}
	})
}

func TestGatherToleratesHangingCollector(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// One collector blocks without honoring ctx; the others finish.
		// Gather must return exactly at the deadline with the finished
		// results applied.
		providers := Providers{
			System:  hangingSystemCollector{},
			Session: slowSessionCollector{},
		}
		start := time.Now()
		snap := providers.Gather(context.Background())
		if elapsed := time.Since(start); elapsed != gatherTimeout {
			t.Errorf("Gather took %v; want exactly the %v deadline", elapsed, gatherTimeout)
		}
		if snap.Session.RemoteAddr != "203.0.113.7" {
			t.Errorf("results from finished collectors should survive a timeout, got %+v", snap.Session)
		}

		// Let the abandoned collector finish. synctest fails the test if
		// its goroutine is still blocked when the bubble ends, so this
		// proves a straggler drains into Gather's buffered channel instead
		// of leaking.
		time.Sleep(time.Minute)
		synctest.Wait()
	})
}

type slowSystemCollector struct{}

func (slowSystemCollector) CollectSystem(ctx context.Context) (SystemInfo, error) {
	time.Sleep(30 * time.Millisecond)
	return SystemInfo{Hostname: "slow"}, nil
}

type slowNetworkCollector struct{}

func (slowNetworkCollector) CollectNetwork(ctx context.Context) (NetworkInfo, error) {
	time.Sleep(30 * time.Millisecond)
	return NetworkInfo{Primary: &Address{IP: "10.0.0.1", Interface: "eth0"}}, nil
}

type slowResourceCollector struct{}

func (slowResourceCollector) CollectResources(ctx context.Context) (ResourceInfo, error) {
	time.Sleep(30 * time.Millisecond)
	return ResourceInfo{}, nil
}

type slowSessionCollector struct{}

func (slowSessionCollector) CollectSession(ctx context.Context) (SessionInfo, error) {
	time.Sleep(30 * time.Millisecond)
	return SessionInfo{RemoteAddr: "203.0.113.7"}, nil
}

type slowLastLoginCollector struct{}

func (slowLastLoginCollector) CollectLastLogin(ctx context.Context) (*LastLoginInfo, error) {
	time.Sleep(30 * time.Millisecond)
	return &LastLoginInfo{Timestamp: time.Now()}, nil
}

// hangingSystemCollector ignores context cancellation entirely, modeling a
// blocking syscall; Gather must still return at its deadline.
type hangingSystemCollector struct{}

func (hangingSystemCollector) CollectSystem(ctx context.Context) (SystemInfo, error) {
	time.Sleep(10 * time.Second)
	return SystemInfo{}, nil
}

func TestParseSSHEnv(t *testing.T) {
	tests := map[string]string{
		"203.0.113.5 50000 192.0.2.10 22":    "203.0.113.5",
		"::ffff:203.0.113.9 50000 ::1 22":    "203.0.113.9",
		"2001:db8::5 50000 2001:db8::1 22":   "2001:db8::5",
		"fe80::1%eth0 50000 fe80::2%eth0 22": "fe80::1",
		"":                                   "",
	}
	for in, want := range tests {
		if got := parseSSHEnv(in); got != want {
			t.Errorf("parseSSHEnv(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPrettyPlatform(t *testing.T) {
	tests := []struct{ platform, version, want string }{
		{"darwin", "15.0", "macOS 15.0"},
		{"rhel", "9.4", "RHEL 9.4"},
		{"opensuse-leap", "15.6", "openSUSE 15.6"},
		{"Microsoft Windows 11 Pro", "10.0.26100", "Microsoft Windows 11 Pro 10.0.26100"},
	}
	for _, tt := range tests {
		if got := prettyPlatform(tt.platform, tt.version); got != tt.want {
			t.Errorf("prettyPlatform(%q, %q) = %q, want %q", tt.platform, tt.version, got, tt.want)
		}
	}
}
