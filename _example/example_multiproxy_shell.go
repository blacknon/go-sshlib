// Copyright (c) 2021 Blacknon. All rights reserved.
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
	// http proxy server
	httpProxyHost = "http-proxy.com"
	httpProxyPort = "4444"

	// socks5 proxy server
	socks5ProxyHost = "socks5-proxy.com"
	socks5ProxyPort = "5555"

	// ssh target server
	host     = "target.com"
	port     = "22"
	user     = "user"
	password = "password"

	termlog = "./test_termlog"
)

func main() {
	// ==========
	// http proxy connect
	// ==========

	httpProxy := &sshlib.Proxy{
		Type: "http",
		Addr: httpProxyHost,
		Port: httpProxyPort,
	}
	httpProxyDialer, err := httpProxy.CreateProxyDialer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ==========
	// socks5 proxy connect
	// ==========

	socks5Proxy := &sshlib.Proxy{
		Type:      "socks5",
		Addr:      httpProxyHost,
		Port:      httpProxyPort,
		Forwarder: httpProxyDialer,
	}
	socks5ProxyDialer, err := httpProxy.CreateProxyDialer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ==========
	// target connect
	// ==========

	// Create target sshlib.Connect
	targetCon := &sshlib.Connect{
		ProxyDialer: socks5ProxyDialer,
	}

	// Create target ssh.AuthMethod
	authMethod := sshlib.CreateAuthMethodPassword(password)

	// Connect target server
	err = targetCon.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
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
