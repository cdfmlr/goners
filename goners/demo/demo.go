package main

import (
	"fmt"
	"net"
	"strings"

	// "os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func readingLivePackets() {
	if handle, err := pcap.OpenLive("lo0", 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("tcp portrange 9000-9100"); err != nil { // optional
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet) // Do something with a packet here.
			// os.Exit(0)
		}
	}
}

func handlePacket(packet gopacket.Packet) {
	// fmt.Println("------ packet.String ------\n")
	fmt.Printf("%v\n", packet.String())

	// fmt.Println("------ packet.Dump ------\n")
	fmt.Printf("%v\n", packet.Dump())

	// fmt.Println("------ My ------\n")

	// fmt.Printf("Time: %v\n",
	// 	packet.Metadata().Timestamp,
	// )
	// fmt.Printf("Length: %v (Captured %v)\n",
	// 	packet.Metadata().Length,
	// 	packet.Metadata().CaptureLength,
	// )

	// var FullPacketData = false
	// for i, l := range packet.Layers() {
	// 	switch FullPacketData {
	// 	case true:
	// 		fmt.Printf("--- Layer %d ---\n%s", i+1, gopacket.LayerDump(l))
	// 	default:
	// 		fmt.Printf("--- Layer %d ---\n%s", i+1, gopacket.LayerString(l))
	// 	}
	// }

	return
}

// func (p *packet) packetString() string {
// 	var b bytes.Buffer
// 	fmt.Fprintf(&b, "PACKET: %d bytes", len(p.Data()))
// 	if p.metadata.Truncated {
// 		b.WriteString(", truncated")
// 	}
// 	if p.metadata.Length > 0 {
// 		fmt.Fprintf(&b, ", wire length %d cap length %d", p.metadata.Length, p.metadata.CaptureLength)
// 	}
// 	if !p.metadata.Timestamp.IsZero() {
// 		fmt.Fprintf(&b, " @ %v", p.metadata.Timestamp)
// 	}
// 	b.WriteByte('\n')
// 	for i, l := range p.layers {
// 		fmt.Fprintf(&b, "- Layer %d (%02d bytes) = %s\n", i+1, len(l.LayerContents()), LayerString(l))
// 	}
// 	return b.String()
// }

// func (p *packet) packetDump() string {
// 	var b bytes.Buffer
// 	fmt.Fprintf(&b, "-- FULL PACKET DATA (%d bytes) ------------------------------------\n%s", len(p.data), hex.Dump(p.data))
// 	for i, l := range p.layers {
// 		fmt.Fprintf(&b, "--- Layer %d ---\n%s", i+1, LayerDump(l))
// 	}
// 	return b.String()
// }

func ifconfig() {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, netInterface := range netInterfaces {
		fmt.Printf("%d: %s  %s\n", netInterface.Index, netInterface.Name, netInterface.HardwareAddr.String())

		addrs, err := netInterface.Addrs()
		if err != nil {
			panic(err)
		}

		for _, addr := range addrs {
			// fmt.Printf("    %s (%s)\n", addr.String(), addr.Network())
			fmt.Printf("    %s ", addr.Network())

			ipStr := strings.Split(addr.String(), "/")[0]
			ip := net.ParseIP(ipStr)
			if ip.To4() != nil {
				fmt.Printf("[IPv4] ")
			} else {
				fmt.Printf("[IPv6] ")
			}

			fmt.Printf("%s  # ", addr.String())

			if ip.IsGlobalUnicast() {
				fmt.Print("GlobalUnicast, ")
			}
			if ip.IsInterfaceLocalMulticast() {
				fmt.Print("InterfaceLocalMulticast, ")
			}
			if ip.IsLinkLocalMulticast() {
				fmt.Print("LinkLocalMulticast, ")
			}
			if ip.IsLinkLocalUnicast() {
				fmt.Print("LinkLocalUnicast, ")
			}
			if ip.IsLoopback() {
				fmt.Print("Loopback, ")
			}
			if ip.IsMulticast() {
				fmt.Print("Multicast, ")
			}
			if ip.IsPrivate() {
				fmt.Print("Private, ")
			}
			if ip.IsUnspecified() {
				fmt.Print("Unspecified, ")
			}
			fmt.Println()
		}
	} 

	// fmt.Printf("%v", interfaces)
}

func main() {
	ifconfig()
	readingLivePackets()
}
