package goners

import (
	"encoding/json"
	"testing"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// sudo go test . -run TestNewPacket -v
func TestNewPacket(t *testing.T) {
	handle, err := pcap.OpenLive("lo0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	err = handle.SetBPFFilter("tcp")
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for p := range packetSource.Packets() {
		packet := NewPacket(p)

		packetJson, err := json.MarshalIndent(packet, "", "  ")
		if err != nil {
			t.Error(err)
		}
		t.Logf(string(packetJson))


		for i, l := range packet.Layers {
			t.Logf("--- Layer %v (%v): \n%v\n", i, l.LayerType, l.Dump())
		}

		break
	}
}
