package sshlib

import (
	"testing"
)

func TestNormalizeTunnelUnit(t *testing.T) {
	if got := normalizeTunnelUnit(TunnelDeviceAny); got != sshTunnelDeviceAny {
		t.Fatalf("normalizeTunnelUnit(any) = %d, want %d", got, sshTunnelDeviceAny)
	}

	if got := normalizeTunnelUnit(7); got != 7 {
		t.Fatalf("normalizeTunnelUnit(7) = %d, want 7", got)
	}
}

func TestParseTunnelInterfaceUnit(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "tun", input: "tun0", want: 0},
		{name: "tap", input: "tap12", want: 12},
		{name: "invalid", input: "tun", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTunnelInterfaceUnit(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTunnelInterfaceUnit(%q) error = nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseTunnelInterfaceUnit(%q) error = %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("parseTunnelInterfaceUnit(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestUtunControlUnit(t *testing.T) {
	if got := utunControlUnit(TunnelDeviceAny); got != 0 {
		t.Fatalf("utunControlUnit(any) = %d, want 0", got)
	}

	if got := utunControlUnit(0); got != 1 {
		t.Fatalf("utunControlUnit(0) = %d, want 1", got)
	}

	if got := utunControlUnit(5); got != 6 {
		t.Fatalf("utunControlUnit(5) = %d, want 6", got)
	}
}

func TestUtunPacketFamily(t *testing.T) {
	tests := []struct {
		name   string
		packet []byte
		want   uint32
	}{
		{name: "empty", packet: nil, want: utunFamilyUnspec},
		{name: "ipv4", packet: []byte{0x45, 0x00}, want: utunFamilyIPv4},
		{name: "ipv6", packet: []byte{0x60, 0x00}, want: utunFamilyIPv6},
		{name: "unknown", packet: []byte{0x10, 0x00}, want: utunFamilyUnspec},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utunPacketFamily(tt.packet); got != tt.want {
				t.Fatalf("utunPacketFamily(%v) = %d, want %d", tt.packet, got, tt.want)
			}
		})
	}
}
