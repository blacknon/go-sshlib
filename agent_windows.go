// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

package sshlib

import (
	sshagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh/agent"
)

// ConnectSshAgent
func ConnectSshAgent() (ag AgentInterface) {
	// first try use pageant then ssh-agent of OpenSSH
	if sshagent.Available() {
		var err error
		ag, _, err = sshagent.New()
		if err != nil {
			ag = agent.NewKeyring()
		}
	} else {
		ag = agent.NewKeyring()
	}

	return
}
