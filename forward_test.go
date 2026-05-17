package sshlib

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestGetDisplay(t *testing.T) {

	for _, tc := range []struct {
		expect int
		input  string
	}{
		{0, ":0.0"},
		{123, ":123.0"},
		{123, ":123"},
		{0, "xxx"},
		{11, "localhost:11.0"},
		{123, "randomhost:123.0"},
	} {
		if act := getX11DisplayNumber(tc.input); act != tc.expect {
			t.Errorf(`unexpected result for getX11Display("%s"), act="%v", exp="%v"`, tc.input, act, tc.expect)
		}
	}
}

func TestX11ConnectForwardedDisplayUsesTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:6011")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{}, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- struct{}{}
			_ = conn.Close()
		}
	}()

	conn, err := x11Connect("127.0.0.1:11.0")
	if err != nil {
		t.Fatalf("x11Connect() error = %v", err)
	}
	_ = conn.Close()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("x11Connect() did not connect to forwarded TCP display")
	}
}

func TestX11ConnectPathDisplayUsesUnixSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix domain sockets are not used for X11 paths on Windows")
	}

	socketPath := filepath.Join("/tmp", "sshlib-x11-path.sock")
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	accepted := make(chan struct{}, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- struct{}{}
			_ = conn.Close()
		}
	}()

	conn, err := x11Connect(socketPath)
	if err != nil {
		t.Fatalf("x11Connect() error = %v", err)
	}
	_ = conn.Close()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("x11Connect() did not connect to Unix socket display path")
	}
}

func TestX11ConnectUnixDisplayUsesDefaultSocketPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix display sockets are not available on Windows")
	}

	dir := "/tmp/.X11-unix"
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	socketPath := filepath.Join(dir, "X77")
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	accepted := make(chan struct{}, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- struct{}{}
			_ = conn.Close()
		}
	}()

	conn, err := x11Connect(":77.0")
	if err != nil {
		t.Fatalf("x11Connect() error = %v", err)
	}
	_ = conn.Close()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("x11Connect() did not connect to default Unix display socket")
	}
}

func ExampleConnect_TCPLocalForward() {
	// host
	host := "target.com"
	port := "22"
	user := "user"
	key := "~/.ssh/id_rsa"

	// port forwarding
	localAddr := "localhost:10022"
	remoteAddr := "localhost:22"

	// Create ssh.AuthMethod
	authMethod, _ := CreateAuthMethodPublicKey(key, "")

	// Create sshlib.Connect
	con := &Connect{}

	// PortForward
	con.TCPLocalForward(localAddr, remoteAddr)

	// Connect ssh server
	con.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
}

func ExampleConnect_UnixLocalForward() {
	host := "target.com"
	port := "22"
	user := "user"
	key := "~/.ssh/id_rsa"

	localPath := "/tmp/local.sock"
	remotePath := "/tmp/remote.sock"

	authMethod, _ := CreateAuthMethodPublicKey(key, "")

	con := &Connect{}

	con.UnixLocalForward(localPath, remotePath)

	con.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
}

func TestConnect_X11Forward(t *testing.T) {
	t.Skip("requires a live SSH session and X11 environment")
}
