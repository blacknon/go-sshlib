package sshlib

import (
	"io"
	"syscall"
	"testing"
)

func TestShouldRetryTunnelCopyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "eio", err: syscall.EIO, want: true},
		{name: "ehostdown", err: syscall.EHOSTDOWN, want: true},
		{name: "eagain", err: syscall.EAGAIN, want: true},
		{name: "ewouldblock", err: syscall.EWOULDBLOCK, want: true},
		{name: "eof", err: io.EOF, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := shouldRetryTunnelCopyError(tt.err); got != tt.want {
				t.Fatalf("shouldRetryTunnelCopyError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
