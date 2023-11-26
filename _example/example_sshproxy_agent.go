// Copyright (c) 2020 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file has a simple ssh proxy connection.
// Also, the authentication method is agent authentication.
// Please replace as appropriate.

package main

import (
	"fmt"
	"os"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	// ssh -J user@proxy.com user@target.com
	// Proxy ssh server
	// host1     = "proxy.com"
	// port1     = "22"
	// user1     = "user"

	// dropbear on linux
	// host1 = "10.161.115.160"
	// port1 = "22"
	// user1 = "root"

	// sshd of OpenSSH on Windows
	host1 = "10.161.115.189"
	port1 = "22"
	user1 = "user_"

	// Target ssh server
	// host2     = "target.com"
	// port2     = "22"
	// user2     = "user"

	// dropbear on linux
	host2 = "10.161.115.160"
	port2 = "22"
	user2 = "root"

	// sshd of OpenSSH on Windows
	// host2 = "10.161.115.189"
	// port2 = "22"
	// user2 = "user_"

	termlog = "./test_termlog"
)

func main() {
	// ==========
	// proxy connect
	// ==========

	// Create proxy sshlib.Connect
	proxyCon := &sshlib.Connect{}

	// Create ssh.AuthMethod
	AuthMethod, err := sshlib.CreateAuthMethodAgent(proxyCon)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect proxy server
	err = proxyCon.CreateClient(host1, port1, user1, []ssh.AuthMethod{AuthMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ==========
	// target connect
	// ==========

	// Create target sshlib.Connect
	targetCon := &sshlib.Connect{
		ProxyDialer: proxyCon.Client,
	}

	// Connect target server
	err = targetCon.CreateClient(host2, port2, user2, []ssh.AuthMethod{AuthMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set terminal log
	// targetCon.SetLog(termlog, false)

	// Create Session
	session, err := targetCon.CreateSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start ssh shell
	targetCon.Shell(session)
}
