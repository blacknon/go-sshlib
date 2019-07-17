// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"
)

// CmdWriter connect and run command over ssh.
// In order to be able to send in parallel from io.MultiWriter, it is made to receive Writer by channel.
func (c *Connect) CmdWriter(command string, output chan []byte, input chan io.Writer) (err error) {
	// create session
	session, err := c.CreateSession()
	if err != nil {
		close(output)
		return
	}

	// ssh agent forwarding
	if c.ForwardAgent {
		c.ForwardSshAgent(session)
	}

	// x11 forwarding
	if c.ForwardX11 {
		err = c.X11Forward(session)
		if err != nil {
			log.Fatal(err)
		}
	}

	// if set Stdin,
	writer, _ := session.StdinPipe()
	input <- writer
	defer writer.Close()

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
	sendCmdOutput(buf, output, isExit)

	return
}

// Cmd connect and run command over ssh.
// Output data is processed by channel because it is executed in parallel. If specification is troublesome, it is good to generate and process session from ssh package.
func (c *Connect) Cmd(command string, output chan []byte) (err error) {
	// create session
	session, err := c.CreateSession()
	if err != nil {
		close(output)
		return
	}

	// ssh agent forwarding
	if c.ForwardAgent {
		c.ForwardSshAgent(session)
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
		session.Stdin = os.Stdin
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
	sendCmdOutput(buf, output, isExit)

	return
}

// sendCmdOutput send to output channel.
func sendCmdOutput(buf *bytes.Buffer, output chan []byte, isExit <-chan bool) {
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
