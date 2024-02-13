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

	"github.com/abakum/go-sshlib"
)

var (
	// dropbear on linux
	// host    = "10.161.115.160"
	// port    = "22"
	// user    = "root"
	// command = "ssh user_@10.161.115.189"

	// sshd of OpenSSH on Windows
	// host    = "10.161.115.189"
	// port    = "22"
	// user    = "user_"
	// command = "ssh root@10.161.115.160"

	// sshd of gliderlabs on Windows
	host    = "10.161.115.189"
	port    = "2222"
	user    = "user_"
	command = "ssh root@10.161.115.160"
)

func main() {
	// Create sshlib.Connect
	con := &sshlib.Connect{
		// If you use x11 forwarding, please uncomment next line.
		// ForwardX11: true,

		// If you use ssh-agent forwarding, please set to true.
		// And after, run `con.ConnectSshAgent()`.
		ForwardAgent: true,
	}

	// setup con.Agent for use ssh-agent
	con.ConnectSshAgent()

	// Connect ssh server
	// set authMethods to nil for use ssh-agent
	err := con.CreateClient(host, port, user, nil)
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

	// Start ssh shell with command
	con.CmdShell(session, command)
}
