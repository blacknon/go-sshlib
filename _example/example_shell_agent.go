// Copyright (c) 2020 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//

package main

import (
	"fmt"
	"os"

	"github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	// dropbear on linux
	// host = "10.161.115.160"
	// port = "22"
	// user = "root"

	// sshd of OpenSSH on Windows
	// host = "10.161.115.189"
	// port = "22"
	// user = "user_"

	// sshd of gliderlabs on Windows
	host = "10.161.115.189"
	port = "2222"
	user = "user_"
)

func main() {
	// Create sshlib.Connect
	con := &sshlib.Connect{
		// If you use x11 forwarding, please uncomment next line.
		// ForwardX11: true,

		// If you use ssh-agent forwarding, uncomment next line.
		ForwardAgent: true,

		// If you use ssh-agent forwarding, and not use sshlib.CreateAuthMethodAgent(con), uncomment next line.
		// Agent:        sshlib.ConnectSshAgent(),
	}

	// Create ssh.AuthMethods
	authMethod, err := sshlib.CreateAuthMethodAgent(con)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect ssh server
	err = con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod})
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
