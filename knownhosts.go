// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// verifyAndAppendNew checks knownhosts from the files stored in c.KnownHostsFiles.
// If there is a problem with the known hosts check, it returns an error and the check content.
// If is no problem, error returns Nil.
//
// 【参考】: https://github.com/tatsushid/minssh/blob/57eae8c5bcf5d94639891f3267f05251f05face4/pkg/minssh/minssh.go#L190-L237
func (c *Connect) verifyAndAppendNew(hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
	// check count KnownHostsFiles
	if len(c.KnownHostsFiles) == 0 {
		return fmt.Errorf("there is no knownhosts file")
	}

	knownHostsFiles := c.KnownHostsFiles

	// abspath
	for i, file := range knownHostsFiles {
		file = getAbsPath(file)
		knownHostsFiles[i] = file
	}

	// get hostKeyCallback
	hostKeyCallback, err := knownhosts.New(knownHostsFiles...)
	if err != nil {
		return
	}

	// check hostkey
	err = hostKeyCallback(hostname, remote, key)
	if err == nil {
		return nil
	}

	// check error
	keyErr, ok := err.(*knownhosts.KeyError)
	if !ok || len(keyErr.Want) > 0 {
		return err
	}

	//
	if answer, err := askAddingUnknownHostKey(hostname, remote, key); err != nil || !answer {
		msg := "host key verification failed"
		if err != nil {
			msg += ": " + err.Error()
		}
		return fmt.Errorf(msg)
	}

	//
	f, err := os.OpenFile(knownHostsFiles[0], os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to add new host key: %s", err)
	}
	defer f.Close()

	//
	var addrs []string
	if remote.String() == hostname {
		addrs = []string{hostname}
	} else {
		addrs = []string{hostname, remote.String()}
	}

	//
	entry := knownhosts.Line(addrs, key)
	if _, err = f.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("failed to add new host key: %s", err)
	}

	return nil
}

// verifyAndAppendNew checks knownhosts from the files stored in c.KnownHostsFiles.
// If there is a problem with the known hosts check, it returns an error and the check content.
// If is no problem, error returns Nil.
//
// 【参考】: https://github.com/tatsushid/minssh/blob/57eae8c5bcf5d94639891f3267f05251f05face4/pkg/minssh/minssh.go#L93-L128
func askAddingUnknownHostKey(address string, remote net.Addr, key ssh.PublicKey) (bool, error) {
	//
	stopC := make(chan struct{})
	defer func() {
		close(stopC)
	}()

	go func() {
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-sigC:
			os.Exit(1)
		case <-stopC:
		}
	}()

	fmt.Printf("The authenticity of host '%s (%s)' can't be established.\n", address, remote.String())
	fmt.Printf("RSA key fingerprint is %s\n", ssh.FingerprintSHA256(key))
	fmt.Printf("Are you sure you want to continue connecting (yes/no)? ")

	b := bufio.NewReader(os.Stdin)
	for {
		answer, err := b.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read answer: %s", err)
		}
		answer = string(strings.ToLower(strings.TrimSpace(answer)))
		if answer == "yes" {
			return true, nil
		} else if answer == "no" {
			return false, nil
		}
		fmt.Print("Please type 'yes' or 'no': ")
	}
}
