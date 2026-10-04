package network

import (
	"context"
	"net"
	"sort"
	"strings"

	gnet "github.com/shirou/gopsutil/v3/net"
)

// Address represents an IP address bound to an interface.
type Address struct {
	Interface string
	IP        string
}

// virtualPrefixes are container, VM, and overlay bridges whose addresses
// say nothing about how to reach the host.
var virtualPrefixes = []string{
	"docker", "veth", "br-", "vbox", "vmnet", "virbr", "cni", "flannel",
	"cali", "lxcbr", "lxdbr", "incusbr", "podman", "fwbr", "fwpr", "fwln",
	"utun", "tap", "wg", "zt",
}

// routeProbes are documentation addresses (RFC 5737, RFC 3849). Connecting
// a UDP socket to one only performs a route lookup; no packet is sent.
var routeProbes = []string{"192.0.2.1:9", "[2001:db8::1]:9"}

// CollectAddresses returns the address carrying the default route as
// primary, then up to maxAdditional other interfaces, one address each.
func CollectAddresses(ctx context.Context, maxAdditional int) (primary *Address, additional []Address, err error) {
	stats, err := gnet.InterfacesWithContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	return selectAddresses(stats, defaultRouteIP(), maxAdditional)
}

func selectAddresses(stats []gnet.InterfaceStat, routeIP net.IP, maxAdditional int) (*Address, []Address, error) {
	// The default route wins even through an interface the virtual filter
	// would hide: a full-tunnel VPN really is the path traffic takes.
	if routeIP != nil {
		for _, stat := range stats {
			for _, a := range stat.Addrs {
				if ip := addrIP(a.Addr); ip != nil && ip.Equal(routeIP) {
					primary := &Address{Interface: stat.Name, IP: ip.String()}
					return primary, others(stats, stat.Name, maxAdditional), nil
				}
			}
		}
	}

	candidates := others(stats, "", -1)
	if len(candidates) == 0 {
		return nil, nil, nil
	}
	primary := candidates[0]
	rest := candidates[1:]
	if maxAdditional > 0 && len(rest) > maxAdditional {
		rest = rest[:maxAdditional]
	}
	return &primary, rest, nil
}

// others lists one address per eligible interface except skip, best-ranked
// first. limit <= 0 means no limit.
func others(stats []gnet.InterfaceStat, skip string, limit int) []Address {
	var out []Address
	for _, stat := range stats {
		if stat.Name == skip || !isUp(stat) || isLoopback(stat) || isVirtual(stat) {
			continue
		}
		if ip := preferredIP(stat); ip != "" {
			out = append(out, Address{Interface: stat.Name, IP: ip})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return rankInterface(out[i].Interface) < rankInterface(out[j].Interface)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// preferredIP picks the interface's first usable IPv4 address, falling back
// to a global IPv6 address so IPv6-only hosts still show one.
func preferredIP(stat gnet.InterfaceStat) string {
	var v6 string
	for _, a := range stat.Addrs {
		ip := addrIP(a.Addr)
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		if ip.To4() != nil {
			return ip.String()
		}
		if v6 == "" && ip.IsGlobalUnicast() {
			v6 = ip.String()
		}
	}
	return v6
}

func addrIP(cidr string) net.IP {
	if ip, _, err := net.ParseCIDR(cidr); err == nil {
		return ip
	}
	return net.ParseIP(cidr)
}

// defaultRouteIP asks the OS which local address it would use to reach the
// internet, trying IPv4 then IPv6.
func defaultRouteIP() net.IP {
	for _, target := range routeProbes {
		conn, err := net.Dial("udp", target)
		if err != nil {
			continue
		}
		addr, ok := conn.LocalAddr().(*net.UDPAddr)
		_ = conn.Close()
		if ok && !addr.IP.IsLoopback() && !addr.IP.IsUnspecified() {
			return addr.IP
		}
	}
	return nil
}

func isUp(stat gnet.InterfaceStat) bool {
	for _, flag := range stat.Flags {
		if strings.EqualFold(flag, "up") {
			return true
		}
	}
	return false
}

func isLoopback(stat gnet.InterfaceStat) bool {
	for _, flag := range stat.Flags {
		if strings.EqualFold(flag, "loopback") {
			return true
		}
	}
	return strings.HasPrefix(strings.ToLower(stat.Name), "lo")
}

func isVirtual(stat gnet.InterfaceStat) bool {
	lower := strings.ToLower(stat.Name)
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

func rankInterface(name string) int {
	lower := strings.ToLower(name)
	for _, prefix := range []string{"eth", "en", "wl", "wifi", "wi-fi", "vmbr", "bond"} {
		if strings.HasPrefix(lower, prefix) {
			return 0
		}
	}
	if strings.HasPrefix(lower, "em") {
		return 1
	}
	return 5
}
