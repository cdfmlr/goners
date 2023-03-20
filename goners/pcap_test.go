package goners

import (
	"context"
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
		ctx     context.Context
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
			args:    args{ctx: context.Background(), device: "lo0", bpf: "tcp", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: false,
		},
		{
			name:    "badDev",
			args:    args{ctx: context.Background(), device: "noexists", bpf: "tcp", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: true,
		},
		{
			name:    "badBpf",
			args:    args{ctx: context.Background(), device: "lo0", bpf: "好久不见呀 我又来了", snaplen: 16384 + 1, promisc: true, timeout: BlockForever},
			wantErr: true,
		},
		{ // 这个测试没啥用，bpf 写错了哈哈哈。timeout 是控制 libpcap 的，不是用来 timeout 后 stop capturing
			name:    "timeout",
			args:    args{ctx: context.Background(), device: "lo0", bpf: "1+1=2", snaplen: 16384 + 1, promisc: true, timeout: time.Second * 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CaptureLivePackets(tt.args.ctx, tt.args.device, tt.args.bpf, tt.args.snaplen, tt.args.promisc, tt.args.timeout)
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

// sudo go test . -run TestCaptureLivePacketsCancel -v
func TestCaptureLivePacketsCancel(t *testing.T) {
	const expectedRunTime = time.Second * 3

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	got, err := CaptureLivePackets(
		ctxWithCancel,
		"lo0",
		"tcp",
		16000,
		true,
		BlockForever)
	if err != nil {
		t.Fatalf("CaptureLivePackets() error = %v, wantErr = false", err)
	}

	stop := time.NewTimer(expectedRunTime)
	go func() {
		<-stop.C
		cancel()
	}()

	counter := atomic.Int64{}
	startTime, duration := time.Now(), time.Nanosecond

LOOP:
	for { // block untail cancel -> stop pcap -> chan close
		select {
		case p, noclosed := <-got:
			_, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			counter.Add(1)
			if !noclosed {
				break LOOP
			}
		default:
			duration = time.Since(startTime)
			if duration-expectedRunTime > time.Second {
				t.Errorf("Slow stop: duration=%v, expected < %v (+1s)",
					duration, expectedRunTime)
				return
			}
		}
	}

	duration = time.Since(startTime)
	t.Logf("captured %v packets. stoped: duration=%v", counter.Load(), duration)
}
