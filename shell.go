// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/abakum/go-ansiterm"
	termm "github.com/abakum/term"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Shell connect login shell over ssh.
// If session is nil then session will be created.
func (c *Connect) Shell(session *ssh.Session) (err error) {
	if session == nil && c.Session == nil {
		session, err = c.CreateSession()
		if err != nil {
			return
		}
		c.Session = session
	}
	defer func() { c.Session = nil }()

	// Input terminal Make raw
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return
	}
	defer term.Restore(fd, state)

	// set FD
	session.Stdin, session.Stdout, session.Stderr = GetStdin(), os.Stdout, os.Stderr

	// setup
	err = c.setupShell(session)
	if err != nil {
		return
	}

	// Start shell
	err = session.Shell()
	if err != nil {
		return
	}

	// keep alive packet
	go c.SendKeepAlive(session)

	err = session.Wait()
	return
}

// Shell connect command shell over ssh.
// Used to start a shell with a specified command.
// If session is nil then session will be created.
func (c *Connect) CmdShell(session *ssh.Session, command string) (err error) {
	if session == nil && c.Session == nil {
		session, err = c.CreateSession()
		if err != nil {
			return
		}
		c.Session = session
	}
	defer func() { c.Session = nil }()

	// Input terminal Make raw
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return
	}
	defer term.Restore(fd, state)

	// set FD
	session.Stdin, session.Stdout, session.Stderr = GetStdin(), os.Stdout, os.Stderr

	// setup
	err = c.setupShell(session)
	if err != nil {
		return
	}

	// Start shell
	err = session.Start(command)
	if err != nil {
		return
	}

	// keep alive packet
	go c.SendKeepAlive(session)

	err = session.Wait()
	if err != nil {
		return
	}

	return
}

// SetLog set up terminal log logging.
// This only happens in Connect.Shell().
func (c *Connect) SetLog(path string, timestamp bool) {
	c.logging = true
	c.logFile = path
	c.logTimestamp = timestamp
}

func (c *Connect) SetLogWithRemoveAnsiCode(path string, timestamp bool) {
	c.logging = true
	c.logFile = path
	c.logTimestamp = timestamp
	c.logRemoveAnsiCode = true
}

// logger is logging terminal log to c.logFile
func (c *Connect) logger(session *ssh.Session) (err error) {
	logfile, err := os.OpenFile(c.logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return
	}

	if !c.logTimestamp && !c.logRemoveAnsiCode {
		session.Stdout = io.MultiWriter(session.Stdout, logfile)
		session.Stderr = io.MultiWriter(session.Stderr, logfile)
	} else {
		buf := new(bytes.Buffer)
		session.Stdout = io.MultiWriter(session.Stdout, buf)
		session.Stderr = io.MultiWriter(session.Stderr, buf)

		go func() {
			preLine := []byte{}
			for {
				if buf.Len() > 0 {
					// get line
					line, err := buf.ReadBytes('\n')

					if err == io.EOF {
						preLine = append(preLine, line...)
						continue
					} else {
						printLine := string(append(preLine, line...))

						if c.logTimestamp {
							timestamp := time.Now().Format("2006/01/02 15:04:05 ") // yyyy/mm/dd HH:MM:SS
							printLine = timestamp + printLine
						}

						// remove ansi code.
						if c.logRemoveAnsiCode {
							printLine, _ = ansiterm.StripBytes([]byte(printLine), ansiterm.WithFe(true))
							printLine += "\n"
						}

						fmt.Fprint(logfile, printLine)
						preLine = []byte{}
					}
				} else {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	return err
}

func (c *Connect) setupShell(session *ssh.Session) (err error) {
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

// ShellAnsi connect login shell over ssh for Windows without VTP
// If session is nil then session will be created.
func (c *Connect) ShellAnsi(session *ssh.Session, emulate bool) (err error) {
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
