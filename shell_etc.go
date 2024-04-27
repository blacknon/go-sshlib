//go:build !windows
// +build !windows

package sshlib

import "golang.org/x/crypto/ssh"

func (c *Connect) ShellAnsi(session *ssh.Session, _ bool) (err error) {
	return c.Shell(session)
}
