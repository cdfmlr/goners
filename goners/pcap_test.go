package goners

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

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
			for k, v := range l.Fields() {
				fmt.Printf("\tfield %q: %q\n", k, v)
			}
		}

		break
	}
}

// sudo go test . -run TestCaptureLivePackets -v
func TestCaptureLivePackets(t *testing.T) {
	type args struct {
		device  string
		bpf     string
		snaplen int32
		promisc bool
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "good",
			args:    args{device: "lo0", bpf: "tcp", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: false,
		},
		{
			name:    "badDev",
			args:    args{device: "noexists", bpf: "tcp", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: true,
		},
		{
			name:    "badBpf",
			args:    args{device: "lo0", bpf: "好久不见呀 我又来了", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: true,
		},
		{
			name:    "timeout",
			args:    args{device: "lo0", bpf: "1+1=2", snaplen: 16384 + 1, promisc: true, timeout: time.Second * 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CaptureLivePackets(tt.args.device, tt.args.bpf, tt.args.snaplen, tt.args.promisc, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("CaptureLivePackets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			timeoutDuration := tt.args.timeout
			if timeoutDuration == BlockForever {
				timeoutDuration = time.Second * 10
			}
			timeout := time.NewTimer(timeoutDuration)

			counter := atomic.Int64{}
		LOOP:
			for {
				select {
				case p := <-got:
					_, err := json.Marshal(p)
					if err != nil {
						t.Fatal(err)
					}
					// t.Log(string(j))
					counter.Add(1)
				case <-timeout.C:
					break LOOP
				}
			}

			t.Logf("captured %v packets.", counter.Load())
		})
	}
}
