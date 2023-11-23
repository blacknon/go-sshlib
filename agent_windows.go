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
	sshagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh/agent"
)

// ConnectSshAgent
func ConnectSshAgent() (ag AgentInterface) {
	// Get env "SSH_AUTH_SOCK" and connect.
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	sock, err := net.Dial("unix", sockPath) // for some versions of Windows

	if err != nil {
		const (
			PIPE         = `\\.\pipe\`
			sshAgentPipe = PIPE + "openssh-ssh-agent"
		)
		if len(sockPath) == 0 {
			sockPath = sshAgentPipe
		}
		if !strings.HasPrefix(sockPath, PIPE) {
			sockPath = PIPE + sockPath
		}
		sock, err = winio.DialPipe(sockPath, nil)
		if err != nil {
			if sshagent.Available() {
				ag, _, err = sshagent.New()
				if err == nil {
					return
				}
			}
		}
	}

	if err != nil {
		ag = agent.NewKeyring()
	} else {
		// connect SSH_AUTH_SOCK
		ag = agent.NewClient(sock)
	}

	return
}
