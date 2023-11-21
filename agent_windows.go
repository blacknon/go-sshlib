// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

package sshlib

import (
	"net"
	"os"

	"github.com/davidmz/go-pageant"

	"golang.org/x/crypto/ssh/agent"
)

// ConnectSshAgent
func ConnectSshAgent() (ag AgentInterface) {
	// Get env "SSH_AUTH_SOCK" and connect.
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	sock, err := net.Dial("unix", sockPath)

	if err != nil {
		ag = pageant.New()

		if ag == nil {
			ag = agent.NewKeyring()
		}
	} else {
		// connect SSH_AUTH_SOCK
		ag = agent.NewClient(sock)
	}

	return
}
