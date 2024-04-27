// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection and port forwarding Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file uses password authentication. Please replace as appropriate.

package main

import (
	"fmt"
	"os"

	sshlib "github.com/abakum/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	host       = "target.com"
	port       = "22"
	user       = "user"
	password   = "password"
	localAddr  = "localhost:80"
	remoteAddr = "localhost:8080"

	termlog = "./test_termlog"
)

func main() {
	// Create sshlib.Connect
	con := &sshlib.Connect{}

	// Create ssh.AuthMethod
	authMethod := sshlib.CreateAuthMethodPassword(password)

	// Connect ssh server
	err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// PortForward
	con.TCPReverseForward(localAddr, remoteAddr)

	// Set terminal log
	con.SetLog(termlog, false)

	// Create session
	session, err := con.CreateSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start ssh shell
	con.Shell(session)
}
