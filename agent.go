// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// ConnectSshAgent
//
func (c *Connect) ConnectSshAgent() {
	// Get env "SSH_AUTH_SOCK" and connect.
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	sock, err := net.Dial("unix", sockPath)

	if err != nil {
		c.agent = agent.NewKeyring()
	} else {
		// connect SSH_AUTH_SOCK
		c.agent = agent.NewClient(sock)
	}
}

// AddKeySshAgent is rapper agent.Add().
// key must be a *rsa.PrivateKey, *dsa.PrivateKey or
// *ecdsa.PrivateKey, which will be inserted into the agent.
//
// Should use `ssh.ParseRawPrivateKey()` or `ssh.ParseRawPrivateKeyWithPassphrase()`.
func (c *Connect) AddKeySshAgent(key interface{}) error {
	addedKey = agent.AddedKey{
		PrivateKey:       key,
		ConfirmBeforeUse: true,
		LifetimeSecs:     3000,
	}

	err = c.agent.Add(addedKey)
}

// ForwardAgent forward ssh-agent in session.
//
func (c *Connect) ForwardAgent(session *ssh.Session) *ssh.Session {
	// forward ssh-agent
	agent.ForwardToAgent(c.Client, c.agent)
	agent.RequestAgentForwarding(session)

	return session
}
