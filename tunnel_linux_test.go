//go:build linux

package sshlib

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestFormatLinuxTunnelOpenError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "not exist",
			err:      os.ErrNotExist,
			contains: []string{"/dev/net/tun", "unavailable", "tun module", "NET_ADMIN"},
		},
		{
			name:     "permission",
			err:      os.ErrPermission,
			contains: []string{"/dev/net/tun", "permission denied", "CAP_NET_ADMIN"},
		},
		{
			name:     "other",
			err:      errors.New("boom"),
			contains: []string{"/dev/net/tun", "boom"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatLinuxTunnelOpenError(tt.err).Error()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("formatLinuxTunnelOpenError(%v) = %q, want substring %q", tt.err, got, want)
				}
			}
		})
	}
}
