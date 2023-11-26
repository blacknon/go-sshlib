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
	"golang.org/x/crypto/ssh/agent"
)

// ConnectSshAgent
func ConnectSshAgent() (ag AgentInterface) {
	const (
		PIPE         = `\\.\pipe\`
		sshAgentPipe = "openssh-ssh-agent"
	)
	var (
		sock net.Conn
		err  error
	)
	// Get env "SSH_AUTH_SOCK" and connect.
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	emptySockPath := len(sockPath) == 0

	if emptySockPath {
		sock, err = pageant.NewConn()
	}

	if err != nil && !emptySockPath {
		// `sc query afunix` for some versions of Windows
		sock, err = net.Dial("unix", sockPath)
	}

	if err != nil {
		if emptySockPath {
			sockPath = sshAgentPipe
		}
		if !strings.HasPrefix(sockPath, PIPE) {
			sockPath = PIPE + sockPath
		}
		sock, err = winio.DialPipe(sockPath, nil)
	}

	if err != nil {
		ag = agent.NewKeyring()
	} else {
		// connect SSH_AUTH_SOCK
		ag = agent.NewClient(sock)
	}

	return
}
