package collectors

import (
	"context"
	"net/netip"
	"os"
	"strings"
)

// EnvSessionCollector reads SSH metadata from environment variables.
type EnvSessionCollector struct{}

// NewSessionCollector returns the default session collector.
func NewSessionCollector() SessionCollector {
	return EnvSessionCollector{}
}

// CollectSession implements SessionCollector.
func (EnvSessionCollector) CollectSession(ctx context.Context) (SessionInfo, error) {
	if addr := parseSSHEnv(os.Getenv("SSH_CONNECTION")); addr != "" {
		return SessionInfo{RemoteAddr: addr, Source: "SSH_CONNECTION"}, nil
	}
	if addr := parseSSHEnv(os.Getenv("SSH_CLIENT")); addr != "" {
		return SessionInfo{RemoteAddr: addr, Source: "SSH_CLIENT"}, nil
	}
	return SessionInfo{}, nil
}

func parseSSHEnv(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	// Dual-stack sshd reports IPv4 clients as ::ffff:a.b.c.d.
	if addr, err := netip.ParseAddr(fields[0]); err == nil {
		return addr.Unmap().WithZone("").String()
	}
	return fields[0]
}
