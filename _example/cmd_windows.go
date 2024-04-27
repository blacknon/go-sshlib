//go:build windows
// +build windows

package sshlib

import (
	"os"

	"github.com/abakum/go-ansiterm"
	termm "github.com/abakum/term"
)

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
