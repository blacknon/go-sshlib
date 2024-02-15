// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// AgentInterface Interface for storing agent.Agent or agent.ExtendedAgent.
type AgentInterface interface{}

// AddKeySshAgent is wrapper agent.Add().
// key must be a *rsa.PrivateKey, *dsa.PrivateKey or
// *ecdsa.PrivateKey, which will be inserted into the agent.
//
// Should use `ssh.ParseRawPrivateKey()` or `ssh.ParseRawPrivateKeyWithPassphrase()`.
func (c *Connect) AddKeySshAgent(sshAgent interface{}, key interface{}) {
	addedKey := agent.AddedKey{
		PrivateKey:       key,
		ConfirmBeforeUse: true,
		LifetimeSecs:     3000,
	}

	switch ag := sshAgent.(type) {
	case agent.ExtendedAgent:
		ag.Add(addedKey)
	case agent.Agent:
		ag.Add(addedKey)
	}
}

// ForwardAgent forward ssh-agent in session.
func (c *Connect) ForwardSshAgent(session *ssh.Session) {
	// forward ssh-agent
	switch ag := c.Agent.(type) {
	case agent.ExtendedAgent:
		agent.ForwardToAgent(c.Client, ag)
	case agent.Agent:
		agent.ForwardToAgent(c.Client, ag)
	}

	agent.RequestAgentForwarding(session)
}

func (c *Connect) ConnectSshAgent() {
	sock, err := NewConn()

	if err != nil {
		c.Agent = agent.NewKeyring()
	} else {
		defer sock.Close()
		c.Agent = agent.NewClient(sock)
	}
}

/*
IdentityAgent
         Specifies the UNIX-domain socket used to communicate with the
         authentication agent.

         This option overrides the SSH_AUTH_SOCK environment variable and
         can be used to select a specific agent.  Setting the socket name
         to none disables the use of an authentication agent.  If the
         string "SSH_AUTH_SOCK" is specified, the location of the socket
         will be read from the SSH_AUTH_SOCK environment variable.
         Otherwise if the specified value begins with a ‘$’ character,
         then it will be treated as an environment variable containing the
         location of the socket.

         Arguments to IdentityAgent may use the tilde syntax to refer to a
         user's home directory or the tokens described in the TOKENS
         section.
*/
