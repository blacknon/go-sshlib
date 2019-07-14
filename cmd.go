// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"bytes"
	"io"
	"log"
	"time"
)

// Cmd connect and run command over ssh.
//
//
func (c *Connect) Cmd(command string, input chan io.Writer, output chan []byte) {
	// create session
	session, err := c.CreateSession()
	if err != nil {
		close(output)
		return
	}

	// Request tty
	err = RequestTty(session)
	if err != nil {
		return
	}

	// ssh agent forwarding
	if c.ForwardAgent {
		session = c.ForwardSshAgent(session)
	}

	// x11 forwarding
	if c.ForwardX11 {
		err = c.X11Forward(session)
		if err != nil {
			log.Fatal(err)
		}
	}

	// if set Stdin,
	if len(c.Stdin) > 0 {
		session.Stdin = bytes.NewReader(c.Stdin)
	} else {
		writer, _ := session.StdinPipe()
		input <- writer
	}

	// Set output buffer
	buf := new(bytes.Buffer)
	session.Stdout = io.MultiWriter(buf)
	session.Stderr = io.MultiWriter(buf)

	// Run Command
	isExit := make(chan bool)
	go func() {
		session.Run(command)
		isExit <- true
	}()

	// Send output channel
GetOutputLoop:
	for {
		if buf.Len() > 0 {
			line, _ := buf.ReadBytes('\n')
			output <- line
		} else {
			select {
			case <-isExit:
				break GetOutputLoop
			case <-time.After(10 * time.Millisecond):
				continue GetOutputLoop
			}
		}
	}

	// last check
	if buf.Len() > 0 {
		for {
			line, err := buf.ReadBytes('\n')
			if err != io.EOF {
				output <- line
			} else {
				break
			}
		}
	}

}
