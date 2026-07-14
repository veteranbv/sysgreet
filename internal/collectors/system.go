package collectors

import (
	"context"
	"time"
)

// gatherTimeout bounds total collection time. A login banner that is a few
// metrics short beats one that hangs the shell.
const gatherTimeout = 250 * time.Millisecond

// SystemInfo captures host identity and session metadata.
type SystemInfo struct {
	Hostname    string
	OS          string
	OSVersion   string
	Arch        string
	Uptime      time.Duration
	CurrentUser string
	HomeDir     string
	Datetime    time.Time
}

// Address represents an IP address bound to an interface.
type Address struct {
	IP        string
	Interface string
}

// NetworkInfo summarizes primary and secondary addresses.
type NetworkInfo struct {
	Primary    *Address
	Additional []Address
}

// SessionInfo provides connection-specific metadata.
type SessionInfo struct {
	RemoteAddr string
	Source     string
}

// MemoryInfo captures memory usage snapshot.
type MemoryInfo struct {
	Total     uint64
	Available uint64
}

// DiskInfo captures disk usage snapshot.
type DiskInfo struct {
	Total uint64
	Used  uint64
}

// CPUInfo captures load or usage metrics.
type CPUInfo struct {
	Load1  float64
	Load5  float64
	Load15 float64
	Usage  float64
	Mode   string // "load" or "usage"
}

// ResourceInfo bundles resource metrics for display.
type ResourceInfo struct {
	Memory MemoryInfo
	Disk   DiskInfo
	CPU    CPUInfo
}

// LastLoginInfo contains last successful login data.
type LastLoginInfo struct {
	Timestamp time.Time
	Source    string
}

// Snapshot aggregates all collector outputs for banner rendering.
type Snapshot struct {
	System    SystemInfo
	Network   NetworkInfo
	Session   SessionInfo
	Resources ResourceInfo
	LastLogin *LastLoginInfo
}

// SystemCollector defines host identity collection behavior.
type SystemCollector interface {
	CollectSystem(ctx context.Context) (SystemInfo, error)
}

// NetworkCollector defines network snapshot behavior.
type NetworkCollector interface {
	CollectNetwork(ctx context.Context) (NetworkInfo, error)
}

// ResourceCollector defines resource metrics behavior.
type ResourceCollector interface {
	CollectResources(ctx context.Context) (ResourceInfo, error)
}

// SessionCollector defines remote session detection behavior.
type SessionCollector interface {
	CollectSession(ctx context.Context) (SessionInfo, error)
}

// LastLoginCollector defines retrieval of last-login metadata.
type LastLoginCollector interface {
	CollectLastLogin(ctx context.Context) (*LastLoginInfo, error)
}

// Providers groups all collectors used to build a snapshot.
type Providers struct {
	System    SystemCollector
	Network   NetworkCollector
	Resources ResourceCollector
	Session   SessionCollector
	LastLogin LastLoginCollector
}

// Gather builds a Snapshot using the configured providers. Collectors run
// concurrently and share a single deadline; whatever has finished by then is
// used and the rest is abandoned, so a stuck collector can never delay a
// login even if it ignores context cancellation. Missing collectors and
// failures are tolerated to allow graceful degradation.
func (p Providers) Gather(ctx context.Context) Snapshot {
	ctx, cancel := context.WithTimeout(ctx, gatherTimeout)
	defer cancel()

	// Each collector sends an apply function; the snapshot is only ever
	// touched from this goroutine, so an abandoned straggler cannot race
	// the caller. The channel is buffered so stragglers finish into it
	// instead of leaking blocked forever.
	results := make(chan func(*Snapshot), 5)
	launched := 0
	launch := func(fn func() func(*Snapshot)) {
		launched++
		go func() {
			results <- fn()
		}()
	}

	if p.System != nil {
		launch(func() func(*Snapshot) {
			sys, err := p.System.CollectSystem(ctx)
			if err != nil {
				recordError("system", err)
				return nil
			}
			return func(s *Snapshot) { s.System = sys }
		})
	}
	if p.Network != nil {
		launch(func() func(*Snapshot) {
			netInfo, err := p.Network.CollectNetwork(ctx)
			if err != nil {
				recordError("network", err)
				return nil
			}
			return func(s *Snapshot) { s.Network = netInfo }
		})
	}
	if p.Resources != nil {
		launch(func() func(*Snapshot) {
			res, err := p.Resources.CollectResources(ctx)
			if err != nil {
				recordError("resources", err)
				return nil
			}
			return func(s *Snapshot) { s.Resources = res }
		})
	}
	if p.Session != nil {
		launch(func() func(*Snapshot) {
			session, err := p.Session.CollectSession(ctx)
			if err != nil {
				recordError("session", err)
				return nil
			}
			return func(s *Snapshot) { s.Session = session }
		})
	}
	if p.LastLogin != nil {
		launch(func() func(*Snapshot) {
			last, err := p.LastLogin.CollectLastLogin(ctx)
			if err != nil {
				recordError("last_login", err)
				return nil
			}
			return func(s *Snapshot) { s.LastLogin = last }
		})
	}

	var snap Snapshot
	for i := 0; i < launched; i++ {
		select {
		case apply := <-results:
			if apply != nil {
				apply(&snap)
			}
		case <-ctx.Done():
			recordError("gather", ctx.Err())
			return snap
		}
	}
	return snap
}

// DemoSnapshot returns a realistic demo snapshot for screenshots and testing.
func DemoSnapshot() Snapshot {
	now := time.Now()
	return Snapshot{
		System: SystemInfo{
			Hostname:    "sysgreet",
			OS:          "Linux Server",
			OSVersion:   "6.8.0",
			Arch:        "x86_64",
			Uptime:      4*24*time.Hour + 12*time.Hour + 33*time.Minute,
			CurrentUser: "demo",
			HomeDir:     "/home/demo",
			Datetime:    now,
		},
		Network: NetworkInfo{
			Primary: &Address{
				IP:        "192.168.1.42",
				Interface: "eth0",
			},
			Additional: []Address{
				{IP: "10.8.0.2", Interface: "tun0"},
			},
		},
		Session: SessionInfo{
			RemoteAddr: "203.0.113.5",
			Source:     "ssh",
		},
		Resources: ResourceInfo{
			Memory: MemoryInfo{
				Total:     16 * 1024 * 1024 * 1024,           // 16 GB
				Available: 12*1024*1024*1024 + 300*1024*1024, // 12.3 GB
			},
			Disk: DiskInfo{
				Total: 512 * 1024 * 1024 * 1024, // 512 GB
				Used:  210 * 1024 * 1024 * 1024, // 210 GB
			},
			CPU: CPUInfo{
				Load1:  0.45,
				Load5:  0.52,
				Load15: 0.60,
				Mode:   "load",
			},
		},
		LastLogin: &LastLoginInfo{
			Timestamp: now.Add(-2 * time.Hour),
			Source:    "203.0.113.10",
		},
	}
}
