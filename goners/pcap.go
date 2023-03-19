package goners

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packet is a View to gopacket.Packet
type Packet struct {
	DeviceIndex int       `json:"device_index"`
	Timestamp   time.Time `json:"timestamp"` // gopacket.Packet.Metadata().Timestamp

	Length        int `json:"length"`         // gopacket.Packet.Metadata().Length
	CaptureLength int `json:"capture_length"` // gopacket.Packet.Metadata().CaptureLength

	Layers []Layer

	packet gopacket.Packet
}

func NewPacket(packet gopacket.Packet) *Packet {
	p := Packet{
		packet: packet,

		DeviceIndex:   packet.Metadata().InterfaceIndex,
		Timestamp:     packet.Metadata().Timestamp,
		Length:        packet.Metadata().Length,
		CaptureLength: packet.Metadata().CaptureLength,
	}

	packetLayers := packet.Layers()
	p.Layers = make([]Layer, 0, len(packetLayers))
	for _, layer := range packetLayers {
		p.Layers = append(p.Layers, NewLayer(layer))
	}

	return &p
}

// To pretty print this, a tty with 96+ chars width is required.
func (p Packet) String() string {
	var sb strings.Builder

	src, dst := p.Flow()
	// TODO: PACKET -> type (e.g. HTTP, TCP, ...)
	sb.WriteString(fmt.Sprintf("%v: %v -> %v @ %v\n",
		"PACKET", src, dst, p.Timestamp))
	sb.WriteString(fmt.Sprintf("\tLength: %v (Captured %v) from device %v\n",
		p.Length, p.CaptureLength, p.DeviceIndex))

	for i, l := range p.Layers {
		sb.WriteString(fmt.Sprintf("  Layer %v ", i+1))
		sb.WriteString(strings.TrimSuffix(strings.ReplaceAll(l.String(), "\n", "\n\t"), "\t"))
	}

	return sb.String()
}

// Flow returns the most high-level readable description
// to the packet flow: src -> dst.
func (p Packet) Flow() (src string, dst string) {
	if len(p.Layers) >= 1 { // Link: MAC -> MAC
		src = p.Layers[0].Src
		dst = p.Layers[0].Dst
	}
	if len(p.Layers) >= 2 { // Network: IP -> IP
		src = p.Layers[1].Src
		dst = p.Layers[1].Dst
	}
	if len(p.Layers) >= 3 { // Transport: IP:port -> IP:port
		if strings.Contains(src, ":") {
			src = fmt.Sprintf("[%s]", src)
		}
		if strings.Contains(dst, ":") {
			dst = fmt.Sprintf("[%s]", dst)
		}
		src = fmt.Sprintf("%s:%s", src, p.Layers[2].Src)
		dst = fmt.Sprintf("%s:%s", dst, p.Layers[2].Dst)
	}
	return src, dst
}

// PacketType returns the most high-level protocol.
func (p Packet) PacketType() string {
	if len(p.Layers) == 0 {
		return "UNK"
	}
	return p.Layers[len(p.Layers)-1].LayerType
}

type PacketView Packet

func (p Packet) MarshalJSON() ([]byte, error) {
	src, dst := p.Flow()

	return json.Marshal(struct {
		PacketView
		Src string `json:"src"`
		Dst string `json:"dst"`
	}{
		PacketView: PacketView(p),
		Src:        src,
		Dst:        dst,
	})
}

// Layer : LinkLayer, NetworkLayer, TransportLayer, ApplicationLayer
//
// LinkLayer: SrcMAC, DstMAC
// NetworkLayer: SrcIP, DstIP
// TransportLayer: SrcPort, DstPort
// ApplicationLayer: Payload
type Layer struct {
	LayerType string `json:"layer_type"`

	Src string `json:"src"`
	Dst string `json:"dst"`

	Payload []byte `json:"payload"`

	layer gopacket.Layer
}

func NewLayer(layer gopacket.Layer) Layer {
	l := Layer{
		layer:     layer,
		LayerType: layer.LayerType().String(),
		Payload:   layer.LayerPayload(),
	}

	switch lnt := layer.(type) { // Link|Network|Transport
	case gopacket.LinkLayer:
		l.Src = lnt.LinkFlow().Src().String()
		l.Dst = lnt.LinkFlow().Dst().String()
	case gopacket.NetworkLayer:
		l.Src = lnt.NetworkFlow().Src().String()
		l.Dst = lnt.NetworkFlow().Dst().String()
	case gopacket.TransportLayer:
		l.Src = lnt.TransportFlow().Src().String()
		l.Dst = lnt.TransportFlow().Dst().String()
	}

	return l
}

// Dump is my version of gopacket.LayerDump
func (l Layer) Dump() string {
	var b bytes.Buffer
	if d, ok := l.layer.(gopacket.Dumper); ok {
		dump := d.Dump()
		if dump != "" {
			b.WriteString(dump)
			if dump[len(dump)-1] != '\n' {
				b.WriteByte('\n')
			}
		}
	}
	b.WriteString(hex.Dump(l.layer.LayerContents()))
	return b.String()
}

// Fields is a map[string]string version of gopacket.layerString
func (l Layer) Fields() map[string]string {
	fields := make(map[string]string)

	v := reflect.ValueOf(l.layer)

	// unwrap: interface, ptr
	for unwrapTimes := 0; unwrapTimes < 3; unwrapTimes++ {
		switch v.Type().Kind() {
		case reflect.Interface, reflect.Ptr:
			if v.IsNil() {
				return fields
			}
			v = v.Elem()
		}
	}

	if v.Type().Kind() != reflect.Struct {
		fields["value"] = fmt.Sprintf("%v", v.Interface())
		return fields
	}

	// assert v.Type().Kind() == reflect.Struct
	for i := 0; i < v.NumField(); i++ {
		ftype := v.Type().Field(i)
		if ftype.Anonymous { // embedded field
			continue
		}
		if ftype.PkgPath != "" { // unexported field
			continue
		}
		key := ftype.Name
		value := v.Field(i).Interface()
		fields[key] = fmt.Sprintf("%v", value)
	}

	return fields
}

const maxLineWidth = 80
const prettyFieldLen = 25 // 25 * 3 < 80

func (l Layer) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v: src %v -> dst %v\n",
		l.LayerType, l.Src, l.Dst))

	// fields:
	// |    k: v    k: v    k: v    k: v    |
	// |    longK: ---longV---              |
	sb.WriteString("Fields:\n")
	line := make([]string, 0, 4)
	longFields := make(map[string]string)
	for k, v := range l.Fields() {
		if len(k)+2+len(v) >= prettyFieldLen {
			longFields[k] = v
			continue
		}
		line = append(line,
			fmt.Sprintf("%-16s", fmt.Sprintf("%v: %v", k, v)))
		if len(line) >= maxLineWidth/prettyFieldLen-1 {
			sb.WriteString("\t")
			sb.WriteString(strings.Join(line, "\t"))
			sb.WriteString("\n")
			line = []string{}
		}
	}
	if len(line) != 0 {
		sb.WriteString("\t")
		sb.WriteString(strings.Join(line, "\t"))
		sb.WriteString("\n")
	}
	for k, v := range longFields {
		sb.WriteString("\t")
		sb.WriteString(fmt.Sprintf("%v: %v\n", k, v))
	}

	// Dump content
	sb.WriteString("Dump:\n\t")
	sb.WriteString(strings.TrimSuffix(strings.ReplaceAll(l.Dump(), "\n", "\n\t"), "\t"))

	return sb.String()
}

// LayerView is Layer: workaround for add Dump() & Fields() retvalue into Layer's json.
type LayerView Layer

func (l Layer) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		LayerView
		Dump   string            `json:"dump"`
		Fields map[string]string `json:"fields"`
	}{
		LayerView: LayerView(l),
		Dump:      l.Dump(),
		Fields:    l.Fields(),
	})
}

const BlockForever = pcap.BlockForever

var ChanBufSize = 16

func CaptureLivePackets(
	device string, bpf string, snaplen int32, promisc bool, timeout time.Duration,
) (chan *Packet, error) {
	chOut := make(chan *Packet, ChanBufSize)

	handle, err := pcap.OpenLive(device, snaplen, promisc, timeout)
	if err != nil {
		return nil, err
	}

	bpf = strings.TrimSpace(bpf)
	if bpf != "" {
		if err := handle.SetBPFFilter(bpf); err != nil {
			return nil, err
		}
	}

	go func() {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			p := NewPacket(packet)
			chOut <- p
		}
		close(chOut)
	}()

	return chOut, nil
}
