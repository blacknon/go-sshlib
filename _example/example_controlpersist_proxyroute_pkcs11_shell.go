// Copyright (c) 2026 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// ControlPersist + ProxyRoute + PKCS11 shell Example file.
// The detached helper asks for the PKCS11 PIN through the parent process TTY.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	sshlib "github.com/blacknon/go-sshlib"
)

var (
	httpProxyHost  = "127.0.0.1"
	httpProxyPort  = "8080"
	pkcs11Provider = "/usr/local/lib/opensc-pkcs11.so"
	bastionHost    = "bastion.example.com"
	bastionPort    = "22"
	bastionUser    = "jump-user"
	targetHost     = "target.example.com"
	targetPort     = "22"
	targetUser     = "target-user"
)

func main() {
	controlPath := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("go-sshlib-proxyroute-pkcs11-%s-%s-%s.sock", targetUser, targetHost, targetPort),
	)

	con := &sshlib.Connect{
		ControlMaster:  "auto",
		ControlPath:    controlPath,
		ControlPersist: 10 * time.Minute,
		ControlPersistAuth: &sshlib.ControlPersistAuth{
			Methods: []sshlib.ControlPersistAuthMethod{
				{
					Type:           "pkcs11",
					PKCS11Provider: pkcs11Provider,
				},
			},
		},
		ProxyRoute: []sshlib.ProxyRoute{
			{
				Type: "http",
				Addr: httpProxyHost,
				Port: httpProxyPort,
			},
			{
				Type: "ssh",
				Addr: bastionHost,
				Port: bastionPort,
				User: bastionUser,
				Auth: &sshlib.ControlPersistAuth{
					Methods: []sshlib.ControlPersistAuthMethod{
						{
							Type:           "pkcs11",
							PKCS11Provider: pkcs11Provider,
						},
					},
				},
			},
		},
	}

	if err := con.CreateClient(targetHost, targetPort, targetUser, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer con.Close()

	if err := con.Shell(nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
