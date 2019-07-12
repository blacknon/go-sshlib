// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/proxy"
)

// Connect structure to store contents about ssh connection.
type Connect struct {
	// Client *ssh.Client
	Client *ssh.Client

	// ProxyDialer
	ProxyDialer proxy.Dialer

	// Session use tty flag.
	TTY bool

	// Forward ssh agent flag.
	ForwardAgent bool

	// Forward x11 flag.
	ForwardX11 bool

	// shell terminal log flag
	Logging bool

	// terminal log add timestamp flag
	logTimestamp bool

	// terminal log path
	logFile string
}

// CreateClient
func (c *Connect) CreateClient(host, user string, port int, signers []ssh.Signer) (err error) {
	uri := net.JoinHostPort(host, string(port))

	// Create new ssh.ClientConfig{}
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            ssh.PublicKeys(signers...),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// check Dialer
	if c.Dialer == nil {
		c.Dialer = proxy.Direct
	}

	// Dial to host:port
	netConn, err := c.Dialer.Dial("tcp", uri)
	if err != nil {
		return
	}

	// Create new ssh connect
	sshCon, channel, req, err := ssh.NewClientConn(netConn, uri, config)
	if err != nil {
		return
	}

	// Create *ssh.Client
	c.Client = ssh.NewClient(sshCon, channel, req)

	return
}

// CreateSession
func (c *Connect) CreateSession() (session ssh.Session, err error) {
	// Create session
	session, err = c.Client.NewSession()

	if c.TTY {
		session, err = RequestTty(session)
		if err != nil {
			return
		}
	}

	return
}

// SendKeepAlive send packet to session.
func (c *Connect) SendKeepAlive(session *ssh.Session) {
	for {
		_, _ = session.SendRequest("keepalive", true, nil)
		time.Sleep(15 * time.Second)
	}
}

// CheckClientAlive check alive ssh.Client.
func (c *Connect) CheckClientAlive() error {
	_, _, err := c.Client.SendRequest("keepalive", true, nil)
	if err == nil || err.Error() == "request failed" {
		return nil
	}
	return err
}

// RequestTty requests the association of a pty with the session on the remote
// host. Terminal size is obtained from the currently connected terminal
//
func RequestTty(session *ssh.Session) (*ssh.Session, error) {
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Get terminal window size
	fd := int(os.Stdin.Fd())
	width, hight, err := terminal.GetSize(fd)
	if err != nil {
		return session, err
	}

	// TODO(blacknon): 環境変数から取得する方式だと、Windowsでうまく動作するか不明なので確認
	term := os.Getenv("TERM")
	if err = session.RequestPty(term, hight, width, modes); err != nil {
		session.Close()
		return session, err
	}

	// Terminal resize goroutine.
	// It is not work Windows os.
	if runtime.GOOS != "windows" {
		winch := syscall.Signal(0x1c)

		signal_chan := make(chan os.Signal, 1)
		signal.Notify(signal_chan, winch)
		go func() {
			for {
				s := <-signal_chan
				switch s {
				case winch:
					fd := int(os.Stdout.Fd())
					width, height, _ = terminal.GetSize(fd)
					session.WindowChange(height, width)
				}
			}
		}()
	}

	return
}
