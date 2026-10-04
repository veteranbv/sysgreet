package collectors

import (
	"context"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

// DefaultSystemCollector gathers host metadata via gopsutil.
type DefaultSystemCollector struct{}

// NewSystemCollector returns the default implementation for all platforms.
func NewSystemCollector() SystemCollector {
	return DefaultSystemCollector{}
}

// CollectSystem implements SystemCollector.
func (DefaultSystemCollector) CollectSystem(ctx context.Context) (SystemInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		// gopsutil fails the whole call when one probe (for example
		// virtualization detection) fails; keep whatever it gathered.
		recordError("system", err)
		if info == nil {
			info = &host.InfoStat{}
		}
	}
	if info.Hostname == "" {
		info.Hostname, _ = os.Hostname()
	}
	if info.Platform == "" {
		info.Platform = runtime.GOOS
	}

	var currentUser, homeDir string
	isRoot := false
	if u, err := user.Current(); err == nil {
		currentUser = u.Username
		homeDir = u.HomeDir
		isRoot = u.Uid == "0"
	}

	// Convert uptime safely from uint64 to int64 for time.Duration
	// Cap at max int64 to prevent overflow (292 years)
	uptime := info.Uptime
	if uptime > uint64(1<<63-1) {
		uptime = uint64(1<<63 - 1)
	}

	return SystemInfo{
		Hostname:    info.Hostname,
		OS:          osName(info.Platform, info.PlatformVersion),
		OSVersion:   info.PlatformVersion,
		Arch:        runtime.GOARCH,
		Uptime:      time.Duration(uptime) * time.Second, //nolint:gosec // G115: Overflow protected above (capped at max int64)
		CurrentUser: currentUser,
		HomeDir:     homeDir,
		IsRoot:      isRoot,
		Datetime:    time.Now(),
	}, nil
}

// osName returns a human name for the OS. On Linux the distribution's own
// PRETTY_NAME ("Ubuntu 24.04.4 LTS") beats anything assembled from
// gopsutil's ID fields.
func osName(platform, version string) string {
	if runtime.GOOS == "linux" {
		if name := osReleasePrettyName("/etc/os-release"); name != "" {
			return name
		}
	}
	return prettyPlatform(platform, version)
}

func osReleasePrettyName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "PRETTY_NAME="); ok {
			return strings.Trim(v, `"'`)
		}
	}
	return ""
}

var platformNames = map[string]string{
	"darwin":    "macOS",
	"rhel":      "RHEL",
	"opensuse":  "openSUSE",
	"linuxmint": "Linux Mint",
	"raspbian":  "Raspberry Pi OS",
	"freebsd":   "FreeBSD",
	"openbsd":   "OpenBSD",
	"netbsd":    "NetBSD",
}

// prettyPlatform names a platform ID without its family, which gopsutil
// reports as e.g. "debian" for Ubuntu or "Standalone Workstation" for macOS.
func prettyPlatform(platform, version string) string {
	key := strings.ToLower(strings.TrimSpace(platform))
	name, ok := platformNames[key]
	if !ok {
		if strings.HasPrefix(key, "opensuse") {
			name = "openSUSE"
		} else {
			name = titleCase(platform)
		}
	}
	if version != "" && !strings.Contains(name, version) {
		name += " " + version
	}
	return name
}

// titleCase capitalizes a lowercase ID ("ubuntu"); names that already carry
// capitals ("Microsoft Windows 11 Pro") are left alone.
func titleCase(s string) string {
	if s == "" || s != strings.ToLower(s) {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
