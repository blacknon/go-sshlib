// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build !windows && !plan9 && !nacl
// +build !windows,!plan9,!nacl

package sshlib

import (
	"net"
	"os"
)

func NewConn() (sock net.Conn, err error) {
	// Get env "SSH_AUTH_SOCK" and connect.
	IdentityAgent := os.Getenv("SSH_AUTH_SOCK")
	sock, err := net.Dial("unix", IdentityAgent)

	return
}
