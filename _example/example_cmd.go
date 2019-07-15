// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// exec command connection Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file uses password authentication. Please replace as appropriate.

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	host     = "target.com"
	port     = "22"
	user     = "user"
	password = "password"

	command1 = "ls -la"
	command2 = "echo command2;timeout 10 cat"
	command3 = "echo command3;timeout 10 cat"
)

func pushInput(isExit <-chan bool, writer io.Writer) {
	rd := bufio.NewReader(os.Stdin)
loop:
	for {
		data, _ := rd.ReadBytes('\n')
		if len(data) > 0 {
			writer.Write(data)
		}

		select {
		case <-isExit:
			break loop
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
}

func main() {
	// Create sshlib.Connect
	con := &sshlib.Connect{}

	// Create ssh.AuthMethod
	authMethod := sshlib.CreateAuthMethodPassword(password)

	// Connect ssh server
	err := con.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// output channel
	output := make(chan []byte)
	defer close(output)

	// print output channel goroutine
	go func() {
		for data := range output {
			str := strings.TrimRight(string(data), "\n")
			fmt.Printf("%s\n", str)
		}
	}()

	// ------------------------------
	// Start command1 (simple run)
	// ------------------------------
	fmt.Printf("\nRun Command Start: %s\n", command1)
	con.Cmd(command1, output)
	fmt.Printf("\nRun Command Exit : %s\n", command1)

	// ------------------------------
	// Start command2 (use key input)
	// ------------------------------
	fmt.Printf("\nRun Command Start: %s\n", command2)
	con.Cmd(command2, output)
	fmt.Printf("\nRun Command Exit : %s\n", command2)

	// ------------------------------
	// Start command3 (use key input)
	//
	// Passing writer by channel so that you can send key input in parallel using io.MultiWriter
	// ------------------------------

	// input channel
	input := make(chan io.Writer)
	defer close(input)

	// inputExit channel
	inputExit := make(chan bool)

	go func(inputExit chan bool) {
		stdin := <-input
		writer := io.MultiWriter(stdin)
		pushInput(inputExit, writer)
	}(inputExit)

	fmt.Printf("\nRun Command Start: %s\n", command3)
	con.Cmd(command3, output)
	fmt.Printf("\nRun Command Exit : %s\n", command3)

	fmt.Println("send")
	close(inputExit)
	fmt.Println("send2")

	fmt.Println("Run Command Exit")
