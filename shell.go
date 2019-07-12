// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Shell connect login shell over ssh.
//
//
func (c *Connect) Shell() (err error) {
	session, err := c.CreateSession()
	if err != nil {
		return
	}

	// set FD
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

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

	// Logging
	if c.Logging {
		session, _ = c.logging(session)
		// TODO(blacknon): errorの出力をライブラリとしてどうするかを検討する
		// if err != nil {
		// }
	}

	// Start shell
	err = session.Shell()
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

//
func (c *Connect) logging(session *ssh.Session) (*ssh.Session, error) {
	log, err := os.OpenFile(c.logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return session, err
	}
	defer log.Close()

	if c.logTimestamp {
		buf := new(bytes.Buffer)
		session.Stdout = io.MultiWriter(session.Stdout, buf)
		session.Stderr = io.MultiWriter(session.Stderr, buf)

		go func() {
			preLine := []byte{}
			for {
				if buf.Len() > 0 {
					line, err := buf.ReadBytes('\n')

					if err == io.EOF {
						preLine = append(preLine, line...)
						continue
					} else {
						timestamp := time.Now().Format("2006/01/02 15:04:05 ") // yyyy/mm/dd HH:MM:SS
						fmt.Fprintf(log, timestamp+string(append(preLine, line...)))
						preLine = []byte{}
					}
				} else {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()

	} else {
		session.Stdout = io.MultiWriter(session.Stdout, log)
		session.Stderr = io.MultiWriter(session.Stderr, log)
	}

	return session, err
}
