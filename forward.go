// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"io"
	"net"
	"sync"
)

// TCPForward
//
func (c *Connect) TCPForward(localAddr, remoteAddr addr) (err error) {
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return
	}

	go func() {
		for {
			//  (type net.Conn)
			conn, err := listner.Accept()
			if err != nil {
				return
			}

			go c.forwarder(conn, "tcp", remoteAddr)
		}
	}()
}

// forwarder tcp/udp port forward.
// NOTE: addr ... remote port address
// dialType ... tcp
func (c *Connect) forwarder(local net.Conn, dialType string, addr string) {
	// Create ssh connect
	remote, err := c.Client.Dial(dialType, addr)

	var wg sync.WaitGroup
	wg.Add(2)

	// Copy local to remote
	go func() {
		io.Copy(remote, local)
		wg.Done()
	}()

	// Copy remote to local
	go func() {
		io.Copy(local, remote)
		wg.Done()
	}()

	wg.Wait()
	conn.Close()
	local.Close()
}
