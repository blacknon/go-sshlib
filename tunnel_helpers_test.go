package sshlib

import (
	"bytes"
	"errors"
	"io"
	"syscall"
	"testing"

	"golang.org/x/crypto/ssh"
)

type stubTunnelRW struct {
	bytes.Buffer
	closeCount int
}

func (s *stubTunnelRW) Close() error {
	s.closeCount++
	return nil
}

type stubChannel struct {
	stubTunnelRW
}

func (c *stubChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, nil
}

func (c *stubChannel) Stderr() io.ReadWriter {
	return &bytes.Buffer{}
}

func (c *stubChannel) AckRequest(ok bool, payload []byte) error {
	return nil
}

func (c *stubChannel) CloseWrite() error {
	return nil
}

type errReader struct {
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestTunnelFinishOnlyOnce(t *testing.T) {
	tunnel := &Tunnel{done: make(chan error, 1)}

	tunnel.finish(io.EOF)
	tunnel.finish(errors.New("ignored"))

	err := <-tunnel.done
	if !errors.Is(err, io.EOF) {
		t.Fatalf("finish() error = %v, want %v", err, io.EOF)
	}
}

func TestTunnelCloseClosesMembersOnce(t *testing.T) {
	local := &stubTunnelRW{}
	remote := &stubTunnelRW{}
	channel := &stubChannel{}
	tunnel := &Tunnel{
		local:   local,
		remote:  remote,
		channel: ssh.Channel(channel),
		done:    make(chan error, 1),
	}

	if err := tunnel.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := tunnel.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}

	if local.closeCount != 1 || remote.closeCount != 1 || channel.closeCount != 1 {
		t.Fatalf("close counts = local:%d remote:%d channel:%d", local.closeCount, remote.closeCount, channel.closeCount)
	}
}

func TestTunnelCopyControlInputWritesFrame(t *testing.T) {
	src := bytes.NewBufferString("packet")
	var dst bytes.Buffer
	tunnel := &Tunnel{
		local: &stubTunnelRW{},
		done:  make(chan error, 1),
	}

	tunnel.copyControlInput(&dst, src)

	frameType, payload, err := readStreamFrame(&dst)
	if err != nil {
		t.Fatalf("readStreamFrame() error = %v", err)
	}
	if frameType != streamFrameStdin || string(payload) != "packet" {
		t.Fatalf("frame = type:%d payload:%q", frameType, payload)
	}
	if err := tunnel.Wait(); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestTunnelCopyControlInputReturnsReadError(t *testing.T) {
	tunnel := &Tunnel{
		local: &stubTunnelRW{},
		done:  make(chan error, 1),
	}

	tunnel.copyControlInput(&bytes.Buffer{}, &errReader{err: io.ErrUnexpectedEOF})

	if err := tunnel.Wait(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Wait() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestTunnelCopyPacketsPropagatesWriteError(t *testing.T) {
	local := &stubTunnelRW{}
	tunnel := &Tunnel{
		local: local,
		done:  make(chan error, 1),
	}

	tunnel.copyPackets(&errWriter{err: io.ErrClosedPipe}, bytes.NewBufferString("data"))

	if err := tunnel.Wait(); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Wait() error = %v, want %v", err, io.ErrClosedPipe)
	}
}

func TestValidateTunnelDevice(t *testing.T) {
	if err := validateTunnelDevice(TunnelModeEthernet, "tap0"); err != nil {
		t.Fatalf("validateTunnelDevice(tap0) error = %v", err)
	}
	if err := validateTunnelDevice(TunnelModePointToPoint, "utun3"); err != nil {
		t.Fatalf("validateTunnelDevice(utun3) error = %v", err)
	}
	if err := validateTunnelDevice(TunnelModePointToPoint, "tap0"); err == nil {
		t.Fatal("validateTunnelDevice(tap0 point-to-point) error = nil, want non-nil")
	}
}

func TestDescribeTunnelUnitAndBuildLinuxTunnelName(t *testing.T) {
	if got := describeTunnelUnit(TunnelDeviceAny); got != "any" {
		t.Fatalf("describeTunnelUnit() = %q, want %q", got, "any")
	}
	if got := buildLinuxTunnelName(3, TunnelModeEthernet); got != "tap3" {
		t.Fatalf("buildLinuxTunnelName() = %q, want %q", got, "tap3")
	}
	if got := buildLinuxTunnelName(-1, TunnelModePointToPoint); got != "" {
		t.Fatalf("buildLinuxTunnelName() = %q, want empty", got)
	}
}

func TestTunnelLocalInterfaceMissing(t *testing.T) {
	tunnel := &Tunnel{LocalName: "definitely-missing-sshlib0"}
	if iface := tunnel.LocalInterface(); iface != nil {
		t.Fatalf("LocalInterface() = %#v, want nil", iface)
	}
}

func TestShouldRetryTunnelCopyErrorAdditionalCases(t *testing.T) {
	retryErrors := []error{syscall.EIO, syscall.EHOSTDOWN, syscall.EAGAIN, syscall.EWOULDBLOCK}
	for _, err := range retryErrors {
		if !shouldRetryTunnelCopyError(err) {
			t.Fatalf("shouldRetryTunnelCopyError(%v) = false, want true", err)
		}
	}
	if shouldRetryTunnelCopyError(io.EOF) {
		t.Fatal("shouldRetryTunnelCopyError(io.EOF) = true, want false")
	}
}
