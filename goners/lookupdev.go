package goners

import (
	"net"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"
)

// Device is a view to net.Interface.
type Device struct {
	Index        int    `json:"index"`         // net.Interface.Index: positive integer that starts at one, zero is never used
	Name         string `json:"name"`          // net.Interface.Name: e.g., "en0", "lo0", "eth0.100"
	HardwareAddr string `json:"hardware_addr"` // net.Interface.HardwareAddr.String(): MAC address
	Addrs        []Addr `json:"addrs"`         // a list of addresses for the interface: IP addresses

	// the underlying net.Interface
	netInterface net.Interface
}

// NewDevice converts net.Interface into a Device
func NewDevice(dev net.Interface) Device {
	d := Device{netInterface: dev}

	d.Index = dev.Index
	d.Name = dev.Name
	d.HardwareAddr = dev.HardwareAddr.String()

	addrs, err := dev.Addrs()
	if err != nil {
		slog.Error("NewDevice: get dev.Addrs() failed",
			"err", err,
		)
	}
	d.Addrs = make([]Addr, 0, len(addrs))

	for _, addr := range addrs {
		d.Addrs = append(d.Addrs, NewAddr(addr))
	}

	return d
}

// Addr is a view to net.Addr + net.IP
type Addr struct {
	NetworkName string `json:"network_name"` // net.Addr.Network(), for example, "tcp", "udp"
	IP          string `json:"ip"`           // IP v4 or v6 address
	Prefix      int    `json:"prefix"`       // subnet prefix
	IPType      IPType `json:"ip_type"`      // IP address types: use TypeString() to get a readable version.

	netAddr net.Addr // underlying net.Addr
	netIp   net.IP   // underlying net.IP
}

// NewAddr converts net.Addr into an Addr
func NewAddr(addr net.Addr) Addr {
	a := Addr{netAddr: addr}
	a.NetworkName = addr.Network()

	var err error = nil
	ipSplit := strings.Split(addr.String(), "/")
	if len(ipSplit) >= 1 {
		a.IP = ipSplit[0]
	}
	if len(ipSplit) >= 2 {
		a.Prefix, err = strconv.Atoi(ipSplit[1])
	}
	if len(ipSplit) > 3 || err != nil {
		slog.Warn("NewAddr: unexpected addr format.",
			"addr.String()", addr.String(),
			"err", err,
		)
	}

	a.netIp = net.ParseIP(a.IP)
	if a.netIp == nil {
		slog.Warn("NewAddr: ParseIP failed")
	}

	for _, t := range ipTypeMappings {
		if t.checker(a.netIp) {
			a.IPType |= t.ipType
		}
	}

	return a
}

// TypeString converts a.Type into a human-readable IP type report.
func (a Addr) TypeString() string {
	var sb strings.Builder

	// for i, s := range ipTypeToString {
	// 	if (1<<i)&a.Type != 0 {
	// 		if sb.Len() > 0 {
	// 			sb.WriteString(", ")
	// 		}
	// 		sb.WriteString(s)
	// 	}
	// }
	for _, t := range ipTypeMappings {
		if t.ipType&a.IPType != 0 {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(t.str)
		}
	}
	return sb.String()
}

func (a Addr) IsIPv4() bool {
	return a.netIp.To4() != nil
}

// IP address types: bitset in an int64.
type IPType int

// IP address types mask.
const (
	// IPv4 "0.0.0.0" or IPv6 "::".
	IPTypeUnspecified IPType = 1 << iota
	IPTypeLoopback
	// ip is a private address: RFC 1918 (IPv4) and RFC 4193 (IPv6).
	IPTypePrivate
	IPTypeMulticast
	IPTypeInterfaceLocalMulticast
	IPTypeLinkLocalMulticast
	IPTypeLinkLocalUnicast
	// global unicast address, including IPv4 private address space or local IPv6 unicast address space.
	IPTypeGlobalUnicast
)

// ipTypeMapping maps ipType to its checker fn & string
type ipTypeMapping struct {
	ipType  IPType
	checker func(net.IP) bool
	str     string
}

// ipTypeCheckers maps
var ipTypeMappings = []ipTypeMapping{
	{IPTypeUnspecified, net.IP.IsUnspecified, "Unspecified"},
	{IPTypeLoopback, net.IP.IsLoopback, "Loopback"},
	{IPTypePrivate, net.IP.IsPrivate, "Private"},
	{IPTypeMulticast, net.IP.IsMulticast, "Multicast"},
	{IPTypeInterfaceLocalMulticast, net.IP.IsInterfaceLocalMulticast, "InterfaceLocalMulticast"},
	{IPTypeLinkLocalMulticast, net.IP.IsLinkLocalMulticast, "LinkLocalMulticast"},
	{IPTypeLinkLocalUnicast, net.IP.IsLinkLocalUnicast, "LinkLocalUnicast"},
	{IPTypeGlobalUnicast, net.IP.IsGlobalUnicast, "GlobalUnicast"},
}

// LookupDevices lists local network interfaces into a []Device.
func LookupDevices() ([]Device, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(netInterfaces))
	for _, ni := range netInterfaces {
		devices = append(devices, NewDevice(ni))
	}

	return devices, nil
}
