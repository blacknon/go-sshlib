// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"io"
	"log"
	"os"

	"github.com/abakum/go-ansiterm"
	termm "github.com/abakum/term"
	"golang.org/x/crypto/ssh"
)

// Command connect and run command over ssh.
// Output data is processed by channel because it is executed in parallel. If specification is troublesome, it is good to generate and process session from ssh package.
func (c *Connect) Command(command string) (err error) {
	// create session
	if c.Session == nil {
		c.Session, err = c.CreateSession()
		if err != nil {
			return
		}
	}
	defer func() { c.Session = nil }()

	// setup options
	err = c.setOption(c.Session)
	if err != nil {
		return
	}

	// Set Stdin, Stdout, Stderr...
	if c.Stdin != nil {
		w, _ := c.Session.StdinPipe()
		go io.Copy(w, c.Stdin)
	} else {
		stdin := GetStdin()
		c.Session.Stdin = stdin
	}

	if c.Stdout != nil {
		or, _ := c.Session.StdoutPipe()
		go io.Copy(c.Stdout, or)
	} else {
		c.Session.Stdout = os.Stdout
	}

	if c.Stderr != nil {
		er, _ := c.Session.StderrPipe()
		go io.Copy(c.Stderr, er)
	} else {
		c.Session.Stderr = os.Stderr
	}

	// Run Command
	c.Session.Run(command)

	return
}

// RequestTty, ForwardSshAgent, X11Forward
func (c *Connect) setOption(session *ssh.Session) (err error) {
	// Request tty
	if c.TTY {
		err = RequestTty(session)
		if err != nil {
			return err
		}
	}

	// x11 forwarding
	if c.ForwardX11 {
		err = c.X11Forward(session)
		if err != nil {
			log.Println(err)
		}
		err = nil
	}

	// ssh agent forwarding
	if c.ForwardAgent {
		c.ForwardSshAgent(session)
	}

	return
}

// CommandAnsi connect and run command over ssh for Windows without VTP.
//
// Output data is processed by channel because it is executed in parallel. If specification is troublesome, it is good to generate and process session from ssh package.
func (c *Connect) CommandAnsi(command string, emulate, fixOpenSSH bool) (err error) {
	// create session
	if c.Session == nil {
		c.Session, err = c.CreateSession()
		if err != nil {
			return
		}
	}
	defer func() { c.Session = nil }()

	// setup options
	err = c.setOption(c.Session)
	if err != nil {
		return
	}

	// Set Stdin, Stdout, Stderr...
	std := termm.NewIOE()
	defer std.Close()
	c.Session.Stdin = std.ReadCloser()

	c.Session.Stdout = os.Stdout
	c.Session.Stdout = os.Stderr
	if emulate {
		wo, do, err := termm.StdOE(os.Stdout)
		if err == nil {
			//Win7
			defer do.Close()
			c.Session.Stdout = wo
		}

		we, de, err := termm.StdOE(os.Stderr)
		if err == nil {
			defer de.Close()
			c.Session.Stderr = we
		}
	}
	if fixOpenSSH {
		// fix sshd of OpenSSH
		command += "&timeout/t 1"
	}

	// Run Command
	err = c.Session.Run(command)

	return
}

// Output runs cmd on the remote host and returns its standard output.
func (c *Connect) Output(cmd string, pty bool) (bs []byte, err error) {
	// create session
	if c.Session == nil {
		c.Session, err = c.CreateSession()
		if err != nil {
			return
		}
	}
	tty := c.TTY
	c.TTY = pty

	defer func() {
		c.Session = nil
		c.TTY = tty
	}()

	// setup options
	err = c.setOption(c.Session)
	if err != nil {
		return
	}
	bs, err = c.Session.Output(cmd)
	if err != nil {
		return
	}
	if pty {
		bs, err = ansiterm.Strip(bs, ansiterm.WithFe(true))
	}
	return
}
