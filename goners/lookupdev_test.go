package goners

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"testing"
)

func TestAddr_TypeString(t *testing.T) {
	type fields struct {
		Type IPType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty", fields{0}, ""},
		{"fristone", fields{IPTypeUnspecified}, "Unspecified"},
		{"middleone", fields{IPTypeMulticast}, "Multicast"},
		{"lastone", fields{IPTypeGlobalUnicast}, "GlobalUnicast"},
		{"combined", fields{IPTypePrivate | IPTypeGlobalUnicast}, "Private, GlobalUnicast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Addr{
				IPType: tt.fields.Type,
			}
			if got := a.TypeString(); got != tt.want {
				t.Errorf("Addr.TypeString() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNewAddr(t *testing.T) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}

	for _, netInterface := range netInterfaces {
		addrs, err := netInterface.Addrs()
		if err != nil {
			t.Fatal(err)
		}

		for _, addr := range addrs {
			got := NewAddr(addr)

			if got.IP == "" {
				t.Error("❌ got.IP == \"\"")
			}
			if got.Prefix == 0 {
				t.Error("❌ got.Prefix == 0")
			}
			if got.NetworkName == "" {
				t.Error("❌ got.NetworkName == 0")
			}
			if got.IPType == 0 {
				t.Error("❌ got.Type == 0")
			}
			if got.TypeString() == "" {
				t.Error("❌ got.TypeString() == 0")
			}

			gotJson, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Error(err)
			}
			t.Logf("✅ gotJson:\n%s, TypeStirng: %+v", gotJson, got.TypeString())

		}
	}
}

func TestNewDevice(t *testing.T) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}

	for _, ni := range netInterfaces {
		got := NewDevice(ni)

		gotVal := reflect.ValueOf(got)
		for i := 0; i < gotVal.NumField(); i++ {
			fieldName := gotVal.Type().Field(i).Name
			if fieldName == "HardwareAddr" {
				// 这个可以为空
				continue
			}

			fieldVal := gotVal.Field(i)
			if fieldVal.IsZero() {
				t.Errorf("❌ field %v is zero.", fieldName)
			}
		}

		gotJson, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Error(err)
		}
		addrIPs := make([]string, 0, len(got.Addrs))
		for _, addr := range got.Addrs {
			addrIPs = append(addrIPs, fmt.Sprintf("%s (%s)", addr.IP, addr.TypeString()))
		}
		t.Logf("✅ got: \n%s\n\t addrs: %v", gotJson, addrIPs)
	}
}

func ExampleLookupDevices() {
	devices, err := LookupDevices()
	if err != nil {
		panic(err)
	}

	devicesJson, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("%v devices:", len(devices))
	fmt.Println(string(devicesJson))
}
