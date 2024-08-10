// Copyright (c) 2024 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"net"

	osfs "github.com/go-git/go-billy/v5/osfs"
	"github.com/pkg/sftp"
	nfs "github.com/willscott/go-nfs"
	nfshelper "github.com/willscott/go-nfs/helpers"
)

func (c *Connect) NFSForward(address, port, basepoint string) (err error) {
	// create listener
	listener, err := net.Listen("tcp", net.JoinHostPort(address, port))
	if err != nil {
		return
	}
	defer listener.Close()

	client, err := sftp.NewClient(c.Client)
	if err != nil {
		return
	}

	sftpfsPlusChange := NewChangeSFTPFS(client, basepoint)

	handler := nfshelper.NewNullAuthHandler(sftpfsPlusChange)
	cacheHelper := nfshelper.NewCachingHandler(handler, 1024)

	// listen
	err = nfs.Serve(listener, cacheHelper)

	return
}

func (c *Connect) NFSReverseSocket(socket, sharepoint string) (err error) {
	// create listener
	listener, err := c.Client.Listen("unix", socket)
	if err != nil {
		return
	}
	defer listener.Close()

	bfs := osfs.New(sharepoint)
	bfsPlusChange := NewChangeOSFS(bfs)

	handler := nfshelper.NewNullAuthHandler(bfsPlusChange)
	cacheHelper := nfshelper.NewCachingHandler(handler, 2048)

	// listen
	err = nfs.Serve(listener, cacheHelper)

	return
}

// NFSReverseForward is Start NFS Server and forward port to remote server.
// This port is forawrd GO-NFS Server.
func (c *Connect) NFSReverseForward(address, port, sharepoint string) (err error) {
	// create listener
	listener, err := c.Client.Listen("tcp", net.JoinHostPort(address, port))
	if err != nil {
		return
	}
	defer listener.Close()

	bfs := osfs.New(sharepoint)
	bfsPlusChange := NewChangeOSFS(bfs)

	handler := nfshelper.NewNullAuthHandler(bfsPlusChange)
	cacheHelper := nfshelper.NewCachingHandler(handler, 1024)

	// listen
	err = nfs.Serve(listener, cacheHelper)

	return
}
