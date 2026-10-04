package network

import (
	"net"
	"testing"

	gnet "github.com/shirou/gopsutil/v4/net"
)

func iface(name string, addrs ...string) gnet.InterfaceStat {
	s := gnet.InterfaceStat{Name: name, Flags: []string{"up"}}
	for _, a := range addrs {
		s.Addrs = append(s.Addrs, gnet.InterfaceAddr{Addr: a})
	}
	return s
}

func TestSelectAddresses_DefaultRouteWins(t *testing.T) {
	// A libvirt host with k3s: name ranking alone picked virbr0.
	stats := []gnet.InterfaceStat{
		iface("lo", "127.0.0.1/8"),
		iface("virbr0", "192.168.122.1/24"),
		iface("cni0", "10.42.0.1/24"),
		iface("wlp2s0", "192.168.1.50/24", "fe80::1/64"),
	}
	primary, rest, err := selectAddresses(stats, net.ParseIP("192.168.1.50"), 3)
	if err != nil {
		t.Fatal(err)
	}
	if primary == nil || primary.Interface != "wlp2s0" || primary.IP != "192.168.1.50" {
		t.Fatalf("primary = %+v, want wlp2s0 192.168.1.50", primary)
	}
	if len(rest) != 0 {
		t.Fatalf("container bridges must be hidden, got %+v", rest)
	}
}

func TestSelectAddresses_VPNDefaultRouteIsPrimary(t *testing.T) {
	stats := []gnet.InterfaceStat{
		iface("eth0", "192.168.1.42/24"),
		iface("wg0", "10.8.0.2/24"),
	}
	primary, rest, _ := selectAddresses(stats, net.ParseIP("10.8.0.2"), 3)
	if primary == nil || primary.Interface != "wg0" {
		t.Fatalf("a full-tunnel VPN carries the default route; got %+v", primary)
	}
	if len(rest) != 1 || rest[0].Interface != "eth0" {
		t.Fatalf("expected eth0 as the other interface, got %+v", rest)
	}
}

func TestSelectAddresses_OneAddressPerInterface(t *testing.T) {
	stats := []gnet.InterfaceStat{
		iface("eth0", "192.168.1.42/24", "192.168.1.43/24", "192.168.1.44/24"),
		iface("eth1", "10.0.0.5/24"),
	}
	primary, rest, _ := selectAddresses(stats, net.ParseIP("192.168.1.42"), 3)
	if primary.IP != "192.168.1.42" {
		t.Fatalf("primary = %+v", primary)
	}
	if len(rest) != 1 || rest[0].Interface != "eth1" {
		t.Fatalf("aliases of the primary must not fill the other slots, got %+v", rest)
	}
}

func TestSelectAddresses_IPv6OnlyHost(t *testing.T) {
	stats := []gnet.InterfaceStat{
		iface("eth0", "fe80::1/64", "2001:db8::42/64"),
	}
	primary, _, _ := selectAddresses(stats, nil, 3)
	if primary == nil || primary.IP != "2001:db8::42" {
		t.Fatalf("an IPv6-only host must still show its global address, got %+v", primary)
	}
}

func TestSelectAddresses_NoRouteFallsBackToRanking(t *testing.T) {
	stats := []gnet.InterfaceStat{
		iface("ppp0", "100.64.0.9/32"),
		iface("enp3s0", "192.168.1.9/24"),
	}
	primary, _, _ := selectAddresses(stats, nil, 3)
	if primary == nil || primary.Interface != "enp3s0" {
		t.Fatalf("ethernet should outrank ppp without route info, got %+v", primary)
	}
}

func TestSelectAddresses_LimitsAdditional(t *testing.T) {
	stats := []gnet.InterfaceStat{
		iface("eth0", "192.168.1.1/24"),
		iface("eth1", "192.168.2.1/24"),
		iface("eth2", "192.168.3.1/24"),
		iface("eth3", "192.168.4.1/24"),
	}
	_, rest, _ := selectAddresses(stats, net.ParseIP("192.168.1.1"), 2)
	if len(rest) != 2 {
		t.Fatalf("expected 2 additional, got %+v", rest)
	}
}

func TestSelectAddresses_DownInterfaceHidden(t *testing.T) {
	down := iface("eth1", "10.0.0.5/24")
	down.Flags = nil
	stats := []gnet.InterfaceStat{iface("eth0", "192.168.1.1/24"), down}
	_, rest, _ := selectAddresses(stats, net.ParseIP("192.168.1.1"), 3)
	if len(rest) != 0 {
		t.Fatalf("down interfaces must be hidden, got %+v", rest)
	}
}
