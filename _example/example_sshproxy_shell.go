// Copyright (c) 2020 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Shell connection Example file.
// Change the value of the variable and compile to make sure that you can actually connect.
//
// This file has a simple ssh proxy connection.
// Also, the authentication method is password authentication.
// Please replace as appropriate.

package main

import (
	"fmt"
	"os"

	sshlib "github.com/abakum/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	// Proxy ssh server
	host1     = "proxy.com"
	port1     = "22"
	user1     = "user"
	password1 = "password"

	// Target ssh server
	host2     = "target.com"
	port2     = "22"
	user2     = "user"
	password2 = "password"

	termlog = "./test_termlog"
)

func main() {
	// ==========
	// proxy connect
	// ==========

	// Create proxy sshlib.Connect
	proxyCon := &sshlib.Connect{}

	// Create proxy ssh.AuthMethod
	proxyAuthMethod := sshlib.CreateAuthMethodPassword(password1)

	// Connect proxy server
	err := proxyCon.CreateClient(host1, port1, user1, []ssh.AuthMethod{proxyAuthMethod})
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

	// Create target ssh.AuthMethod
	targetAuthMethod := sshlib.CreateAuthMethodPassword(password2)

	// Connect target server
	err = targetCon.CreateClient(host2, port2, user2, []ssh.AuthMethod{targetAuthMethod})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set terminal log
	targetCon.SetLog(termlog, false)

	// Create Session
	session, err := targetCon.CreateSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start ssh shell
	targetCon.Shell(session)
}
