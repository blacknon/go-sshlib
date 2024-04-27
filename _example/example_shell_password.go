// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file uses password authentication. Please replace as appropriate.

package main

import (
	"fmt"
	"os"

	"github.com/abakum/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	// host     = "10.161.115.160"
	// port     = "22"
	// user     = "root"
	host     = "10.161.115.189"
	port     = "2222"
	user     = "user_"
	password = "password"

	termlog = "./test_termlog"
)

func main() {
	// Create sshlib.Connect
	con := &sshlib.Connect{
		// If you use x11 forwarding, please set to true.
		ForwardX11: false,

		// If you use ssh-agent forwarding, please set to true.
		// And after, run `con.ConnectSshAgent()`.
		ForwardAgent: false,
	}

	// Create ssh.AuthMethod
	authMethod := sshlib.CreateAuthMethodPassword(password)

	// If you use ssh-agent, uncomment it.
	con.ConnectSshAgent()

	// Connect ssh server
	err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set terminal log
	// con.SetLog(termlog, true)
	con.SetLogWithRemoveAnsiCode(termlog, false)

	// Start ssh shell
	con.Shell(nil)
}
