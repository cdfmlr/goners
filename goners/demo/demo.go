package main

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func readingLivePackets() {
	if handle, err := pcap.OpenLive("en1", 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("tcp"); err != nil { // optional
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet) // Do something with a packet here.
		}
	}
}

func handlePacket(packet gopacket.Packet) {
	fmt.Printf("%#v\n", packet.String())
}

func main() {
	readingLivePackets()
}
