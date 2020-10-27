// Copyright (c) 2020 Blacknon. All rights reserved.
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

	sshlib "github.com/blacknon/go-sshlib"
)

var (
	host = "target.com"
	port = "22"
	user = "user"
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
	authMethod, err := sshlib.CreateAuthMethodPKCS11("/usr/local/opt/opensc/lib/opensc-pkcs11.so", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// If you use ssh-agent forwarding, uncomment it.
	// con.ConnectSshAgent()

	// Connect ssh server
	err = con.CreateClient(host, port, user, authMethod)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create Session
	session, err := con.CreateSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start ssh shell
	con.Shell(session)
}
