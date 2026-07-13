package banner

import (
	"fmt"
	"strings"
	"time"

	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
)

// SystemSectionBuilder renders uptime, user, datetime, and last-login information.
type SystemSectionBuilder struct{}

// Key returns the section identifier.
func (SystemSectionBuilder) Key() string { return "system" }

// Enabled returns true when any system display flags are enabled.
func (SystemSectionBuilder) Enabled(cfg config.Config) bool {
	return cfg.Display.Uptime || cfg.Display.User || cfg.Display.Datetime || cfg.Display.LastLogin
}

// Build renders the system section lines.
func (SystemSectionBuilder) Build(snap collectors.Snapshot, cfg config.Config) (Section, bool) {
	var lines []string
	if cfg.Display.Uptime && snap.System.Uptime > 0 {
		lines = append(lines, fmt.Sprintf("Uptime: %s", humanDuration(snap.System.Uptime)))
	}
	if cfg.Display.User && strings.TrimSpace(snap.System.CurrentUser) != "" {
		userLine := fmt.Sprintf("User: %s", snap.System.CurrentUser)
		if snap.System.HomeDir != "" {
			userLine += " " + snap.System.HomeDir
		}
		lines = append(lines, userLine)
	}
	// A timed-out system collector leaves Datetime zero; omit the line
	// rather than print the epoch.
	if cfg.Display.Datetime && !snap.System.Datetime.IsZero() {
		lines = append(lines, fmt.Sprintf("Time: %s", snap.System.Datetime.Format(time.RFC1123)))
	}
	if cfg.Display.LastLogin && snap.LastLogin != nil {
		lines = append(lines, fmt.Sprintf("Last login: %s (%s)", snap.LastLogin.Timestamp.Format(time.RFC1123), snap.LastLogin.Source))
	}
	if len(lines) == 0 {
		return Section{}, false
	}
	return Section{Key: "system", Title: "System", Lines: lines}, true
}

// NetworkSectionBuilder renders primary/secondary IP data and remote address.
type NetworkSectionBuilder struct{}

// Key returns unique identifier.
func (NetworkSectionBuilder) Key() string { return "network" }

// Enabled returns true if network details are enabled in config.
func (NetworkSectionBuilder) Enabled(cfg config.Config) bool {
	return cfg.Display.IPAddresses || cfg.Display.RemoteIP
}

// Build builds network section lines.
func (NetworkSectionBuilder) Build(snap collectors.Snapshot, cfg config.Config) (Section, bool) {
	var lines []string
	if cfg.Display.IPAddresses {
		if snap.Network.Primary != nil {
			lines = append(lines, fmt.Sprintf("Primary: %s", formatAddress(*snap.Network.Primary, cfg.Network.ShowInterfaceNames)))
		}
		for _, addr := range snap.Network.Additional {
			lines = append(lines, fmt.Sprintf("Secondary: %s", formatAddress(addr, cfg.Network.ShowInterfaceNames)))
		}
	}
	if cfg.Display.RemoteIP && snap.Session.RemoteAddr != "" {
		lines = append(lines, fmt.Sprintf("Remote: %s", snap.Session.RemoteAddr))
	}
	if len(lines) == 0 {
		return Section{}, false
	}
	return Section{Key: "network", Title: "Network", Lines: lines}, true
}

func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "unknown"
	}
	minutes := int(d.Minutes())
	days := minutes / (60 * 24)
	hours := (minutes / 60) % 24
	mins := minutes % 60
	segments := []string{}
	if days > 0 {
		segments = append(segments, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		segments = append(segments, fmt.Sprintf("%dh", hours))
	}
	if mins > 0 || len(segments) == 0 {
		segments = append(segments, fmt.Sprintf("%dm", mins))
	}
	return strings.Join(segments, " ")
}

func formatAddress(addr collectors.Address, withInterface bool) string {
	if !withInterface || strings.TrimSpace(addr.Interface) == "" {
		return addr.IP
	}
	return fmt.Sprintf("%s (%s)", addr.IP, addr.Interface)
}
