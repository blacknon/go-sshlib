package sshlib

import (
	"bytes"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
)

type stubListener struct {
	closed bool
}

func (l *stubListener) Accept() (net.Conn, error) { return nil, errors.New("not implemented") }
func (l *stubListener) Close() error {
	l.closed = true
	return nil
}
func (l *stubListener) Addr() net.Addr { return &net.UnixAddr{Name: "stub", Net: "unix"} }

func TestShortSocketTokenStable(t *testing.T) {
	path := "/tmp/control.sock"
	got1 := shortSocketToken(path)
	got2 := shortSocketToken(path)
	if got1 != got2 {
		t.Fatalf("shortSocketToken() mismatch: %q != %q", got1, got2)
	}
	if len(got1) != 8 {
		t.Fatalf("shortSocketToken() len = %d, want 8", len(got1))
	}
}

func TestEncodeDecodeControlAddr(t *testing.T) {
	tcpAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222}
	encoded := encodeControlAddr(tcpAddr)
	decoded := decodeControlAddr(encoded)
	gotTCP, ok := decoded.(*net.TCPAddr)
	if !ok {
		t.Fatalf("decodeControlAddr(tcp) type = %T, want *net.TCPAddr", decoded)
	}
	if gotTCP.String() != tcpAddr.String() {
		t.Fatalf("decodeControlAddr(tcp) = %q, want %q", gotTCP.String(), tcpAddr.String())
	}

	unixAddr := &net.UnixAddr{Name: "/tmp/test.sock", Net: "unix"}
	encoded = encodeControlAddr(unixAddr)
	decoded = decodeControlAddr(encoded)
	gotUnix, ok := decoded.(*net.UnixAddr)
	if !ok {
		t.Fatalf("decodeControlAddr(unix) type = %T, want *net.UnixAddr", decoded)
	}
	if gotUnix.String() != unixAddr.String() || gotUnix.Network() != unixAddr.Network() {
		t.Fatalf("decodeControlAddr(unix) = %v, want %v", gotUnix, unixAddr)
	}
}

func TestDecodeControlAddrFallbackToStaticAddr(t *testing.T) {
	decoded := decodeControlAddr(controlAddr{Network: "custom", Address: "endpoint"})
	got, ok := decoded.(staticAddr)
	if !ok {
		t.Fatalf("decodeControlAddr(custom) type = %T, want staticAddr", decoded)
	}
	if got.Network() != "custom" || got.String() != "endpoint" {
		t.Fatalf("decodeControlAddr(custom) = %#v", got)
	}
}

func TestWriteReadStreamFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("hello")

	if err := writeStreamFrame(&buf, streamFrameStdout, payload); err != nil {
		t.Fatalf("writeStreamFrame() error = %v", err)
	}

	frameType, got, err := readStreamFrame(&buf)
	if err != nil {
		t.Fatalf("readStreamFrame() error = %v", err)
	}
	if frameType != streamFrameStdout {
		t.Fatalf("readStreamFrame() frameType = %d, want %d", frameType, streamFrameStdout)
	}
	if string(got) != string(payload) {
		t.Fatalf("readStreamFrame() payload = %q, want %q", got, payload)
	}
}

func TestEnsureControlPathCreatesParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "control", "master.sock")
	if err := ensureControlPath(path); err != nil {
		t.Fatalf("ensureControlPath() error = %v", err)
	}

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat(parent dir) error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("parent path is not a directory: %s", filepath.Dir(path))
	}
}

func TestCleanupStaleControlSocketIgnoresMissingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.sock")
	if err := cleanupStaleControlSocket(path); err != nil {
		t.Fatalf("cleanupStaleControlSocket() error = %v", err)
	}
}

func TestCleanupStaleControlSocketRejectsNonSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-socket")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := cleanupStaleControlSocket(path)
	if err == nil {
		t.Fatal("cleanupStaleControlSocket() error = nil, want non-nil")
	}
}

func TestControlMasterCloseListenerRemovesRegisteredListener(t *testing.T) {
	listener := &stubListener{}
	m := &controlMaster{
		listeners: map[uint64]net.Listener{
			7: listener,
		},
	}

	if err := m.closeListener(7); err != nil {
		t.Fatalf("closeListener() error = %v", err)
	}
	if !listener.closed {
		t.Fatal("closeListener() did not close listener")
	}
	if _, ok := m.listeners[7]; ok {
		t.Fatal("closeListener() did not remove listener from map")
	}
}

func TestControlMasterLookupListenerMissing(t *testing.T) {
	m := &controlMaster{listeners: map[uint64]net.Listener{}}

	_, err := m.lookupListener(42)
	if err == nil {
		t.Fatal("lookupListener() error = nil, want non-nil")
	}
}
