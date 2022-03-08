// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

// 【参考】
//   - https://github.com/tatsushid/minssh/commit/57eae8c5bcf5d94639891f3267f05251f05face4

package sshlib

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func (c *Connect) setupShell(session *ssh.Session) (err error) {
	// set FD
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Logging
	if c.logging {
		err = c.logger(session)
		if err != nil {
			log.Println(err)
		}
	}
	err = nil

	// Request tty
	err = RequestTty(session)
	if err != nil {
		return err
	}

	// x11 forwarding
	if c.ForwardX11 {
		err = c.X11Forward(session)
		if err != nil {
			log.Println(err)
		}
	}
	err = nil

	// ssh agent forwarding
	if c.ForwardAgent {
		c.ForwardSshAgent(session)
	}

	return
}
