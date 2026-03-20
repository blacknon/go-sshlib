// Copyright (c) 2026 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// ControlMaster-like connection sharing shell Example file.
// Start one process first, then run another process with the same ControlPath.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	host = "127.0.0.1"
	port = "22"
	user = "user"
	key  = "~/.ssh/id_rsa"
)

func main() {
	authMethod, err := sshlib.CreateAuthMethodPublicKey(key, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	controlPath := filepath.Join(os.TempDir(), "go-sshlib-control.sock")
	con := &sshlib.Connect{
		ControlMaster: "auto",
		ControlPath:   controlPath,
	}

	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer con.Close()

	if con.Client != nil {
		session, err := con.CreateSession()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer session.Close()

		fmt.Println("mode: master")
		if err := con.Shell(session); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("master shell exited; keeping the shared connection alive until SIGINT/SIGTERM")
		waitForTerminateSignal()
		return
	}

	fmt.Println("mode: slave")
	if err := con.Shell(nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func waitForTerminateSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	<-sigCh
}
