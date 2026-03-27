package sshlib

import (
	"bytes"
	"net"
	"testing"
)

type nopTunnelLocal struct {
	bytes.Buffer
}

func (n *nopTunnelLocal) Close() error {
	return nil
}

func (n *nopTunnelLocal) Read(p []byte) (int, error) {
	return 0, nil
}

func TestTunnelCopyControlOutputReadsFrames(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		if err := writeStreamFrame(server, streamFrameStdout, []byte("packet")); err != nil {
			t.Errorf("write stdout frame: %v", err)
			return
		}
		if err := writeStreamFrame(server, streamFrameExit, encodeExitStatus(0)); err != nil {
			t.Errorf("write exit frame: %v", err)
		}
	}()

	local := &nopTunnelLocal{}
	tunnel := &Tunnel{
		local: local,
		done:  make(chan error, 1),
	}

	go tunnel.copyControlOutput(local, client)

	if err := tunnel.Wait(); err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}

	if got := local.String(); got != "packet" {
		t.Fatalf("local payload = %q, want %q", got, "packet")
	}
}

func TestTunnelCopyControlOutputReturnsErrorFrame(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		if err := writeStreamFrame(server, streamFrameError, []byte("boom")); err != nil {
			t.Errorf("write error frame: %v", err)
		}
	}()

	local := &nopTunnelLocal{}
	tunnel := &Tunnel{
		local: local,
		done:  make(chan error, 1),
	}

	go tunnel.copyControlOutput(local, client)

	if err := tunnel.Wait(); err == nil || err.Error() != "boom" {
		t.Fatalf("Wait() error = %v, want boom", err)
	}
}
