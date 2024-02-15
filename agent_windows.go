// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

package sshlib

import (
	"net"
	"os"
	"strings"

	"github.com/Microsoft/go-winio"
	"github.com/abakum/pageant"
)

func NewConn() (sock net.Conn, err error) {
	const (
		PIPE         = `\\.\pipe\`
		sshAgentPipe = "openssh-ssh-agent"
	)
	// Get env "SSH_AUTH_SOCK" and connect.
	IdentityAgent := os.Getenv("SSH_AUTH_SOCK")
	emptySockPath := IdentityAgent == ""

	if emptySockPath {
		sock, err = pageant.NewConn()
	}

	if err != nil && !emptySockPath {
		// `sc query afunix` for some versions of Windows
		sock, err = net.Dial("unix", IdentityAgent)
	}

	if err != nil {
		if emptySockPath {
			IdentityAgent = sshAgentPipe
		}
		if !strings.HasPrefix(IdentityAgent, PIPE) {
			IdentityAgent = PIPE + IdentityAgent
		}
		sock, err = winio.DialPipe(IdentityAgent, nil)
	}
	return sock, err

}
