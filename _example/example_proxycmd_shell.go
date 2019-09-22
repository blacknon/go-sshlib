// Copyright (c) 2019 Blacknon. All rights reserved.
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

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	// Proxy ssh server
	host         = "proxy.com"
	port         = "22"
	user         = "user"
	password     = "password"
	proxyCommand = "ssh -W %h:%p ProxyServer"

	termlog = "./test_termlog"
)

func main() {
	// ==========
	// proxy
	// ==========

	p := &sshlib.Proxy{
		Type:    command,
		Command: proxyCommand,
	}

	dialer, err := p.CreateProxyDialer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create target sshlib.Connect
	targetCon := &sshlib.Connect{
		ProxyDialer: dialer,
	}

	// Create target ssh.AuthMethod
	targetAuthMethod := sshlib.CreateAuthMethodPassword(password1)

	// Connect target server
	err = targetCon.CreateClient(host1, user1, port1, []ssh.AuthMethod{targetAuthMethod})
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
