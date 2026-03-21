// Copyright (c) 2026 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection and tunnel forwarding Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file uses password authentication. Please replace as appropriate.
// The local tunnel device is currently supported on Linux and macOS.

package main

import (
	"fmt"
	"os"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	host     = "target.com"
	port     = "22"
	user     = "user"
	password = "password"

	localTun  = 0
	remoteTun = 0

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

	// Open SSH tunnel forwarding. This is similar to `ssh -w 0:0 user@host`.
	tunnel, err := con.Tunnel(localTun, remoteTun)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tunnel.Close()

	fmt.Printf("local tunnel interface: %s\n", tunnel.LocalName)
	fmt.Println("Please configure interface addresses and routes with OS tools.")

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
