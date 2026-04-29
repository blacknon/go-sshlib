package sshlib

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/net/proxy"
)

func TestCreateClientWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")

	con := &Connect{}

	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{CreateAuthMethodPassword(password)}); err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}
	defer con.Client.Close()

	session, err := con.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer session.Close()

	output, err := session.Output("printf connected")
	if err != nil {
		t.Fatalf("session.Output() error = %v", err)
	}

	if string(output) != "connected" {
		t.Fatalf("unexpected output: got %q want %q", string(output), "connected")
	}
}

func TestAutoReconnectCreateSessionWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")

	con := &Connect{
		AutoReconnect: true,
	}

	authMethod := CreateAuthMethodPassword(password)
	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}
	defer con.Close()

	if err := con.Client.Close(); err != nil {
		t.Fatalf("Client.Close() error = %v", err)
	}
	con.Client = nil

	session, err := con.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession() after reconnect error = %v", err)
	}
	defer session.Close()

	output, err := session.Output("printf reconnected")
	if err != nil {
		t.Fatalf("session.Output() error = %v", err)
	}

	if string(output) != "reconnected" {
		t.Fatalf("unexpected output: got %q want %q", string(output), "reconnected")
	}
}

func TestControlMasterCommandWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := shortControlPath(t, "control.sock")

	master := &Connect{
		ControlMaster:  "auto",
		ControlPath:    controlPath,
		ControlPersist: time.Minute,
	}
	masterAuthMethod := CreateAuthMethodPassword(password)
	master.ControlPersistAuth = &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{masterAuthMethod}}
	if err := master.CreateClient(host, port, user, []ssh.AuthMethod{masterAuthMethod}); err != nil {
		t.Fatalf("master CreateClient() error = %v", err)
	}
	defer master.Close()

	slave := &Connect{
		ControlMaster:  "auto",
		ControlPath:    controlPath,
		ControlPersist: time.Minute,
	}
	slaveAuthMethod := CreateAuthMethodPassword(password)
	slave.ControlPersistAuth = &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{slaveAuthMethod}}
	if err := slave.CreateClient(host, port, user, []ssh.AuthMethod{slaveAuthMethod}); err != nil {
		t.Fatalf("slave CreateClient() error = %v", err)
	}

	var stdout bytes.Buffer
	slave.Stdout = &stdout
	if err := slave.Command("printf shared"); err != nil {
		t.Fatalf("slave Command() error = %v", err)
	}

	if got := stdout.String(); got != "shared" {
		t.Fatalf("unexpected output: got %q want %q", got, "shared")
	}
}

func TestControlPersistReplacesExpiredMaster(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := shortControlPath(t, "persist.sock")
	persist := 1200 * time.Millisecond

	runCommand := func(expected string) {
		authMethod := CreateAuthMethodPassword(password)
		con := &Connect{
			ControlMaster:      "auto",
			ControlPath:        controlPath,
			ControlPersist:     persist,
			ControlPersistAuth: &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}},
		}
		if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
			t.Fatalf("CreateClient() error = %v", err)
		}

		var stdout bytes.Buffer
		con.Stdout = &stdout
		if err := con.Command("printf " + expected); err != nil {
			t.Fatalf("Command() error = %v", err)
		}

		if got := stdout.String(); got != expected {
			t.Fatalf("unexpected output: got %q want %q", got, expected)
		}
	}

	runCommand("first")
	time.Sleep(persist + 800*time.Millisecond)
	runCommand("second")
}

func TestControlMasterForwardAgentWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := shortControlPath(t, "agent.sock")

	authMethod := CreateAuthMethodPassword(password)
	con := &Connect{
		ControlMaster:      "auto",
		ControlPath:        controlPath,
		ControlPersist:     time.Minute,
		ControlPersistAuth: &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}},
		ForwardAgent:       true,
		Agent:              agent.NewKeyring(),
	}
	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}

	var stdout bytes.Buffer
	con.Stdout = &stdout
	if err := con.Command(`if [ -S "$SSH_AUTH_SOCK" ]; then printf forwarded; else printf missing; fi`); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	if got := stdout.String(); got != "forwarded" {
		t.Fatalf("unexpected output: got %q want %q", got, "forwarded")
	}
}

func TestControlMasterForwardX11WithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY is required for X11 forwarding integration test")
	}
	if os.Getenv("XAUTHORITY") == "" && os.Getenv("HOME") == "" {
		t.Skip("XAUTHORITY or HOME is required for X11 forwarding integration test")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := shortControlPath(t, "x11.sock")

	authMethod := CreateAuthMethodPassword(password)
	con := &Connect{
		ControlMaster:      "auto",
		ControlPath:        controlPath,
		ControlPersist:     time.Minute,
		ControlPersistAuth: &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}},
		ForwardX11:         true,
	}
	if err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod}); err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}

	var stdout bytes.Buffer
	con.Stdout = &stdout
	if err := con.Command(`if [ -n "$DISPLAY" ]; then printf forwarded; else printf missing; fi`); err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	if got := stdout.String(); got != "forwarded" {
		t.Fatalf("unexpected output: got %q want %q", got, "forwarded")
	}
}

func TestControlMasterTCPLocalForwardWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, port := newControlMasterPair(t)

	localAddr := "127.0.0.1:" + port
	remoteAddr := testSSHServiceAddr()

	if err := slave.TCPLocalForward(localAddr, remoteAddr); err != nil {
		t.Fatalf("TCPLocalForward() error = %v", err)
	}

	waitForTCPReady(t, localAddr)

	conn, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		t.Fatalf("DialTimeout() error = %v", err)
	}
	defer conn.Close()

	assertSSHBanner(t, conn)
}

func TestControlMasterUnixLocalForwardWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, _ := newControlMasterPair(t)

	localDir := shortControlPath(t, "unix-forward")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	localPath := filepath.Join(localDir, "local.sock")
	t.Cleanup(func() {
		_ = os.Remove(localPath)
	})

	remotePath := fmt.Sprintf("/tmp/go-sshlib-controlmaster-unix-%d.sock", time.Now().UnixNano())

	listener, err := slave.Listen("unix", remotePath)
	if err != nil {
		t.Fatalf("Listen(unix) error = %v", err)
	}
	defer listener.Close()

	acceptErrCh := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			acceptErrCh <- err
			return
		}
		defer conn.Close()

		buf := make([]byte, len("ping"))
		if _, err := io.ReadFull(conn, buf); err != nil {
			acceptErrCh <- err
			return
		}
		if string(buf) != "ping" {
			acceptErrCh <- fmt.Errorf("unexpected payload: got %q want %q", string(buf), "ping")
			return
		}
		if _, err := conn.Write([]byte("pong")); err != nil {
			acceptErrCh <- err
			return
		}

		acceptErrCh <- nil
	}()

	if err := slave.UnixLocalForward(localPath, remotePath); err != nil {
		t.Fatalf("UnixLocalForward() error = %v", err)
	}

	waitForUnixReady(t, localPath)

	conn, err := net.DialTimeout("unix", localPath, 5*time.Second)
	if err != nil {
		t.Fatalf("DialTimeout(unix) error = %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}

	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	reply := make([]byte, len("pong"))
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatalf("ReadFull() error = %v", err)
	}
	if string(reply) != "pong" {
		t.Fatalf("unexpected reply: got %q want %q", string(reply), "pong")
	}

	if err := <-acceptErrCh; err != nil {
		t.Fatalf("remote unix listener error = %v", err)
	}
}

func TestControlMasterTCPDynamicForwardWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, port := newControlMasterPair(t)

	localAddr := "127.0.0.1:" + port
	errCh := make(chan error, 1)
	go func() {
		errCh <- slave.TCPDynamicForward("127.0.0.1", port)
	}()

	waitForTCPReady(t, localAddr)

	dialer, err := proxy.SOCKS5("tcp", localAddr, nil, proxy.Direct)
	if err != nil {
		t.Fatalf("proxy.SOCKS5() error = %v", err)
	}

	conn, err := dialer.Dial("tcp", testSSHServiceAddr())
	if err != nil {
		t.Fatalf("SOCKS5 dial error = %v", err)
	}
	defer conn.Close()

	assertSSHBanner(t, conn)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("TCPDynamicForward() error = %v", err)
		}
	default:
	}
}

func TestControlMasterDialReturnsRemoteErrorWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, _ := newControlMasterPair(t)

	conn, err := slave.Dial("tcp", "127.0.0.1:1")
	if err == nil {
		conn.Close()
		t.Fatal("Dial() error = nil, want remote dial error")
	}
}

func TestControlMasterTCPDynamicForwardReturnsDialErrorWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, port := newControlMasterPair(t)

	localAddr := "127.0.0.1:" + port
	errCh := make(chan error, 1)
	go func() {
		errCh <- slave.TCPDynamicForward("127.0.0.1", port)
	}()

	waitForTCPReady(t, localAddr)

	dialer, err := proxy.SOCKS5("tcp", localAddr, nil, proxy.Direct)
	if err != nil {
		t.Fatalf("proxy.SOCKS5() error = %v", err)
	}

	conn, err := dialer.Dial("tcp", "127.0.0.1:1")
	if err == nil {
		conn.Close()
		t.Fatal("SOCKS5 dial error = nil, want remote dial error")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("TCPDynamicForward() error = %v", err)
		}
	default:
	}
}

func TestControlMasterSFTPClientWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	_, slave, _ := newControlMasterPair(t)

	client, err := slave.newSFTPClient()
	if err != nil {
		t.Fatalf("newSFTPClient() error = %v", err)
	}
	defer client.Close()

	path, err := client.RealPath(".")
	if err != nil {
		t.Fatalf("RealPath(.) error = %v", err)
	}
	if path == "" {
		t.Fatal("RealPath(.) = empty path, want non-empty path")
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func shortControlPath(t *testing.T, name string) string {
	t.Helper()

	dir, err := os.MkdirTemp("/tmp", "go-sshlib-")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return filepath.Join(dir, name)
}

func newControlMasterPair(t *testing.T) (*Connect, *Connect, string) {
	t.Helper()

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := shortControlPath(t, "forward.sock")

	masterAuthMethod := CreateAuthMethodPassword(password)
	master := &Connect{
		ControlMaster:      "auto",
		ControlPath:        controlPath,
		ControlPersist:     time.Minute,
		ControlPersistAuth: &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{masterAuthMethod}},
	}
	if err := master.CreateClient(host, port, user, []ssh.AuthMethod{masterAuthMethod}); err != nil {
		t.Fatalf("master CreateClient() error = %v", err)
	}
	t.Cleanup(func() {
		_ = master.Close()
	})

	slaveAuthMethod := CreateAuthMethodPassword(password)
	slave := &Connect{
		ControlMaster:      "auto",
		ControlPath:        controlPath,
		ControlPersist:     time.Minute,
		ControlPersistAuth: &ControlPersistAuth{AuthMethods: []ssh.AuthMethod{slaveAuthMethod}},
	}
	if err := slave.CreateClient(host, port, user, []ssh.AuthMethod{slaveAuthMethod}); err != nil {
		t.Fatalf("slave CreateClient() error = %v", err)
	}
	t.Cleanup(func() {
		_ = slave.Close()
	})

	return master, slave, freeTCPPort(t)
}

func testSSHServiceAddr() string {
	return net.JoinHostPort(
		getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1"),
		getenvDefault("SSHLIB_TEST_PORT", "2222"),
	)
}

func freeTCPPort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

func waitForTCPReady(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("listener did not become ready: %s", addr)
}

func waitForUnixReady(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		info, err := os.Stat(path)
		if err == nil && info.Mode()&os.ModeSocket != 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("listener did not become ready: %s", path)
}

func assertSSHBanner(t *testing.T, conn net.Conn) {
	t.Helper()

	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if !strings.HasPrefix(string(buf[:n]), "SSH-2.0-") {
		t.Fatalf("unexpected banner: %q", string(buf[:n]))
	}
}
