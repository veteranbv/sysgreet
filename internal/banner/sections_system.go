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

// Build renders the system section.
func (SystemSectionBuilder) Build(snap collectors.Snapshot, cfg config.Config) (Section, bool) {
	var items []Item
	// Exact values for --json; the items carry the human-friendly forms.
	data := map[string]any{}
	if cfg.Display.Uptime && snap.System.Uptime > 0 {
		items = append(items, Item{Label: "Uptime", Value: humanDuration(snap.System.Uptime)})
		data["uptime_seconds"] = int64(snap.System.Uptime.Seconds())
	}
	if cfg.Display.User && strings.TrimSpace(snap.System.CurrentUser) != "" {
		it := Item{Label: "User", Value: snap.System.CurrentUser}
		if snap.System.IsRoot {
			// Knowing you are root before typing anything is the point.
			it.Level = LevelCrit
		}
		items = append(items, it)
	}
	// A timed-out system collector leaves Datetime zero; omit the line
	// rather than print the epoch.
	now := snap.System.Datetime
	if cfg.Display.Datetime && !now.IsZero() {
		items = append(items, Item{Label: "Time", Value: now.Format("Mon 02 Jan 15:04 MST")})
		data["time"] = now.Format(time.RFC3339)
	}
	if cfg.Display.LastLogin && snap.LastLogin != nil {
		if now.IsZero() {
			now = time.Now()
		}
		value := relativeTime(snap.LastLogin.Timestamp, now)
		if src := snap.LastLogin.Source; src != "" {
			value += " from " + src
		}
		items = append(items, Item{Label: "Last login", Value: value})
		data["last_login"] = snap.LastLogin.Timestamp.Format(time.RFC3339)
		if src := snap.LastLogin.Source; src != "" {
			data["last_login_from"] = src
		}
	}
	return newSection("system", "System", items, data)
}

// NetworkSectionBuilder renders the host's addresses and the SSH client.
type NetworkSectionBuilder struct{}

// Key returns unique identifier.
func (NetworkSectionBuilder) Key() string { return "network" }

// Enabled returns true if network details are enabled in config.
func (NetworkSectionBuilder) Enabled(cfg config.Config) bool {
	return cfg.Display.IPAddresses || cfg.Display.RemoteIP
}

// Build renders the network section. The default-route address comes
// first; each address is labeled with its interface.
func (NetworkSectionBuilder) Build(snap collectors.Snapshot, cfg config.Config) (Section, bool) {
	var items []Item
	if cfg.Display.IPAddresses {
		addrs := snap.Network.Additional
		if snap.Network.Primary != nil {
			addrs = append([]collectors.Address{*snap.Network.Primary}, addrs...)
		}
		for _, addr := range addrs {
			label := "IP"
			if cfg.Network.ShowInterfaceNames && strings.TrimSpace(addr.Interface) != "" {
				label = addr.Interface
			}
			items = append(items, Item{Label: label, Value: addr.IP})
		}
	}
	if cfg.Display.RemoteIP && snap.Session.RemoteAddr != "" {
		items = append(items, Item{Label: "From", Value: snap.Session.RemoteAddr})
	}
	return newSection("network", "Network", items, nil)
}

// humanDuration keeps the two largest units: minutes stop mattering once
// a host has been up for days.
func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "unknown"
	}
	minutes := int(d.Minutes())
	parts := []struct {
		n    int
		unit string
	}{
		{minutes / (60 * 24), "d"},
		{(minutes / 60) % 24, "h"},
		{minutes % 60, "m"},
	}
	var segments []string
	for i, p := range parts {
		if p.n == 0 {
			// Skip empty leading units, and stop at the first empty one
			// after them ("7d", not "7d 0h").
			if len(segments) > 0 {
				break
			}
			if i < len(parts)-1 {
				continue
			}
		}
		segments = append(segments, fmt.Sprintf("%d%s", p.n, p.unit))
		if len(segments) == 2 {
			break
		}
	}
	return strings.Join(segments, " ")
}

// relativeTime phrases t relative to now ("3h ago"), falling back to a
// date beyond a week.
func relativeTime(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Mon 02 Jan 2006")
	}
}
