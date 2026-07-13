package banner

import (
	"strings"
	"testing"
	"time"

	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
)

func TestSystemSectionBuilder_Enabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{
			name: "all enabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime:    true,
					User:      true,
					Datetime:  true,
					LastLogin: true,
				},
			},
			want: true,
		},
		{
			name: "uptime only",
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime: true,
				},
			},
			want: true,
		},
		{
			name: "all disabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime:    false,
					User:      false,
					Datetime:  false,
					LastLogin: false,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := SystemSectionBuilder{}
			if got := builder.Enabled(tt.cfg); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemSectionBuilder_Build(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		snap      collectors.Snapshot
		cfg       config.Config
		wantOK    bool
		wantLines int
	}{
		{
			name: "all fields enabled and present",
			snap: collectors.Snapshot{
				System: collectors.SystemInfo{
					Uptime:      time.Hour*48 + time.Minute*30,
					CurrentUser: "testuser",
					HomeDir:     "/home/testuser",
					Datetime:    now,
				},
				LastLogin: &collectors.LastLoginInfo{
					Timestamp: now.Add(-time.Hour),
					Source:    "192.168.1.1",
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime:    true,
					User:      true,
					Datetime:  true,
					LastLogin: true,
				},
			},
			wantOK:    true,
			wantLines: 4,
		},
		{
			name: "uptime only",
			snap: collectors.Snapshot{
				System: collectors.SystemInfo{
					Uptime: time.Hour * 24,
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime: true,
				},
			},
			wantOK:    true,
			wantLines: 1,
		},
		{
			name: "no fields enabled",
			snap: collectors.Snapshot{
				System: collectors.SystemInfo{
					Uptime: time.Hour * 24,
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime:   false,
					User:     false,
					Datetime: false,
				},
			},
			wantOK:    false,
			wantLines: 0,
		},
		{
			name: "zero uptime is skipped",
			snap: collectors.Snapshot{
				System: collectors.SystemInfo{
					Uptime: 0,
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					Uptime: true,
				},
			},
			wantOK:    false,
			wantLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := SystemSectionBuilder{}
			section, ok := builder.Build(tt.snap, tt.cfg)

			if ok != tt.wantOK {
				t.Errorf("Build() ok = %v, want %v", ok, tt.wantOK)
			}

			if len(section.Lines) != tt.wantLines {
				t.Errorf("Build() lines = %d, want %d", len(section.Lines), tt.wantLines)
			}

			if ok && section.Key != "system" {
				t.Errorf("Build() key = %q, want %q", section.Key, "system")
			}
		})
	}
}

func TestNetworkSectionBuilder_Enabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{
			name: "IP addresses enabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: true,
				},
			},
			want: true,
		},
		{
			name: "Remote IP enabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					RemoteIP: true,
				},
			},
			want: true,
		},
		{
			name: "both enabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: true,
					RemoteIP:    true,
				},
			},
			want: true,
		},
		{
			name: "all disabled",
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: false,
					RemoteIP:    false,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NetworkSectionBuilder{}
			if got := builder.Enabled(tt.cfg); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNetworkSectionBuilder_Build(t *testing.T) {
	tests := []struct {
		name      string
		snap      collectors.Snapshot
		cfg       config.Config
		wantOK    bool
		wantLines int
	}{
		{
			name: "primary and secondary IPs",
			snap: collectors.Snapshot{
				Network: collectors.NetworkInfo{
					Primary: &collectors.Address{
						IP:        "192.168.1.100",
						Interface: "eth0",
					},
					Additional: []collectors.Address{
						{IP: "10.0.0.5", Interface: "wlan0"},
					},
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: true,
				},
				Network: config.NetworkConfig{
					ShowInterfaceNames: true,
				},
			},
			wantOK:    true,
			wantLines: 2,
		},
		{
			name: "remote IP only",
			snap: collectors.Snapshot{
				Session: collectors.SessionInfo{
					RemoteAddr: "203.0.113.5",
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					RemoteIP: true,
				},
			},
			wantOK:    true,
			wantLines: 1,
		},
		{
			name: "no network data",
			snap: collectors.Snapshot{},
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: true,
					RemoteIP:    true,
				},
			},
			wantOK:    false,
			wantLines: 0,
		},
		{
			name: "interface names hidden",
			snap: collectors.Snapshot{
				Network: collectors.NetworkInfo{
					Primary: &collectors.Address{
						IP:        "192.168.1.100",
						Interface: "eth0",
					},
				},
			},
			cfg: config.Config{
				Display: config.DisplayConfig{
					IPAddresses: true,
				},
				Network: config.NetworkConfig{
					ShowInterfaceNames: false,
				},
			},
			wantOK:    true,
			wantLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NetworkSectionBuilder{}
			section, ok := builder.Build(tt.snap, tt.cfg)

			if ok != tt.wantOK {
				t.Errorf("Build() ok = %v, want %v", ok, tt.wantOK)
			}

			if len(section.Lines) != tt.wantLines {
				t.Errorf("Build() lines = %d, want %d", len(section.Lines), tt.wantLines)
			}

			if ok && section.Key != "network" {
				t.Errorf("Build() key = %q, want %q", section.Key, "network")
			}
		})
	}
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "unknown"},
		{"negative", -time.Hour, "unknown"},
		{"minutes only", time.Minute * 45, "45m"},
		{"hours and minutes", time.Hour*2 + time.Minute*30, "2h 30m"},
		{"days and hours", time.Hour*24*3 + time.Hour*5, "3d 5h"},
		{"days hours minutes", time.Hour*24*2 + time.Hour*12 + time.Minute*33, "2d 12h 33m"},
		{"exact days", time.Hour * 24 * 7, "7d"},
		{"exact hours", time.Hour * 3, "3h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanDuration(tt.duration); got != tt.want {
				t.Errorf("humanDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name          string
		addr          collectors.Address
		withInterface bool
		want          string
	}{
		{
			name:          "with interface",
			addr:          collectors.Address{IP: "192.168.1.1", Interface: "eth0"},
			withInterface: true,
			want:          "192.168.1.1 (eth0)",
		},
		{
			name:          "without interface",
			addr:          collectors.Address{IP: "192.168.1.1", Interface: "eth0"},
			withInterface: false,
			want:          "192.168.1.1",
		},
		{
			name:          "empty interface with flag",
			addr:          collectors.Address{IP: "192.168.1.1", Interface: ""},
			withInterface: true,
			want:          "192.168.1.1",
		},
		{
			name:          "whitespace interface with flag",
			addr:          collectors.Address{IP: "192.168.1.1", Interface: "   "},
			withInterface: true,
			want:          "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatAddress(tt.addr, tt.withInterface); got != tt.want {
				t.Errorf("formatAddress(%+v, %v) = %q, want %q", tt.addr, tt.withInterface, got, tt.want)
			}
		})
	}
}

func TestSystemSectionOmitsUnavailableData(t *testing.T) {
	// A timed-out system collector leaves the snapshot zero-valued; the
	// section must not invent an epoch timestamp or unknown user.
	section, ok := SystemSectionBuilder{}.Build(collectors.Snapshot{}, config.Default())
	if !ok {
		return // nothing to render is acceptable
	}
	for _, line := range section.Lines {
		if strings.Contains(line, "0001") || strings.Contains(line, "unknown") {
			t.Fatalf("section renders placeholder data: %q", line)
		}
	}
}
