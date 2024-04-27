//go:build windows
// +build windows

package sshlib

import (
	"os"

	termm "github.com/abakum/term"
	"golang.org/x/crypto/ssh"
)

// ShellAnsi connect login shell over ssh for Windows without VTP
// If session is nil then session will be created.
func (c *Connect) ShellAnsi(session *ssh.Session, emulate bool) (err error) {
	// create session
	if session == nil && c.Session == nil {
		session, err = c.CreateSession()
		if err != nil {
			return
		}
		c.Session = session
	}
	defer func() { c.Session = nil }()

	// Set Stdin, Stdout, Stderr...
	std := termm.NewIOE()
	defer std.Close()
	session.Stdin = std.ReadCloser()

	session.Stdout = os.Stdout
	session.Stdout = os.Stderr
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

	// setup
	err = c.setupShell(c.Session)
	if err != nil {
		return
	}

	// Start shell
	err = c.Session.Shell()
	if err != nil {
		return
	}

	// keep alive packet
	go c.SendKeepAlive(c.Session)

	err = c.Session.Wait()
	return
}
