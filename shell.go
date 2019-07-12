// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// Shell connect login shell over ssh.
//
//
func (c Connect) Shell() (err error) {
	session, err := c.CreateSession()
	if err != nil {
		return
	}

	// input
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return
	}
	defer terminal.Restore(fd, state)

	// Request tty
	session, err = RequestTty(session)
	if err != nil {
		return
	}

	// Start shell
	err = session.Shell()
	if err != nil {
		return
	}

	// Terminal resize
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGWINCH)
	go func() {
		for {
			s := <-signal_chan
			switch s {
			case syscall.SIGWINCH:
				fd := int(os.Stdout.Fd())
				width, height, _ = terminal.GetSize(fd)
				session.WindowChange(height, width)
			}
		}
	}()

	// keep alive packet
	go c.SendKeepAlive(session)

	err = session.Wait()
	if err != nil {
		return
	}

	return
}
