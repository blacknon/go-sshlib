// Copyright (c) 2026 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// ControlMaster-like connection sharing shell Example file.
// Start one process first, then run another process with the same ControlPath.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
	host = "target.com"
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

	controlPath := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("go-sshlib-%s-%s-%s.sock", user, host, port),
	)
	con := &sshlib.Connect{
		ControlMaster:      "auto",
		ControlPath:        controlPath,
		ControlPersist:     10 * time.Minute,
		ControlPersistAuth: &sshlib.ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}},
	}

	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer con.Close()

	switch {
	case con.SpawnedControlMaster():
		fmt.Println("mode: slave (started new control master)")
	case con.IsControlClient():
		fmt.Println("mode: slave (connected to existing control master)")
	default:
		fmt.Println("mode: direct")
	}

	if err := con.Shell(nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
