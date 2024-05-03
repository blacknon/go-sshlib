//go:build !windows
// +build !windows

package sshlib

import "github.com/abakum/go-ansiterm"

func (c *Connect) CommandAnsi(command string, _, _ bool) (err error) {
	return c.Command(command)
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
