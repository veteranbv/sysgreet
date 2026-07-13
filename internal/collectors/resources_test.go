package collectors

import (
	"context"
	"testing"
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

func TestGatherRunsCollectorsConcurrently(t *testing.T) {
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
	elapsed := time.Since(start)

	if elapsed > 120*time.Millisecond {
		t.Errorf("Gather took %v; collectors do not appear to run concurrently", elapsed)
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
}

func TestGatherToleratesHangingCollector(t *testing.T) {
	// One collector blocks without honoring ctx; the others finish. Gather
	// must return at the deadline with the finished results applied.
	providers := Providers{
		System:  hangingSystemCollector{},
		Session: slowSessionCollector{},
	}
	start := time.Now()
	snap := providers.Gather(context.Background())
	if elapsed := time.Since(start); elapsed > gatherTimeout+150*time.Millisecond {
		t.Errorf("Gather took %v; timeout did not bound a hanging collector", elapsed)
	}
	if snap.Session.RemoteAddr != "203.0.113.7" {
		t.Errorf("results from finished collectors should survive a timeout, got %+v", snap.Session)
	}
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
