package goners

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
)

// Packet is a View to gopacket.Packet
type Packet struct {
	DeviceIndex int
	Timestamp   time.Time // gopacket.Packet.Metadata().Timestamp

	Length        int // gopacket.Packet.Metadata().Length
	CaptureLength int // gopacket.Packet.Metadata().CaptureLength

	Layers []Layer

	packet gopacket.Packet
}

func NewPacket(packet gopacket.Packet) Packet {
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

	return p
}

// Layer : LinkLayer, NetworkLayer, TransportLayer, ApplicationLayer
//
// LinkLayer: SrcMAC, DstMAC
// NetworkLayer: SrcIP, DstIP
// TransportLayer: SrcPort, DstPort
// ApplicationLayer: Payload
type Layer struct {
	LayerType string

	Src string
	Dst string

	Payload []byte

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
