package sshlib

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestCreateClientWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")

	con := &Connect{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

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

func TestControlMasterCommandWithDockerSSHD(t *testing.T) {
	if os.Getenv("SSHLIB_INTEGRATION") == "" {
		t.Skip("set SSHLIB_INTEGRATION=1 to run integration tests")
	}

	host := getenvDefault("SSHLIB_TEST_HOST", "127.0.0.1")
	port := getenvDefault("SSHLIB_TEST_PORT", "2222")
	user := getenvDefault("SSHLIB_TEST_USER", "testuser")
	password := getenvDefault("SSHLIB_TEST_PASSWORD", "testpass")
	controlPath := filepath.Join(t.TempDir(), "control.sock")

	master := &Connect{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ControlMaster:   "auto",
		ControlPath:     controlPath,
	}
	if err := master.CreateClient(host, port, user, []ssh.AuthMethod{CreateAuthMethodPassword(password)}); err != nil {
		t.Fatalf("master CreateClient() error = %v", err)
	}
	defer master.Close()

	slave := &Connect{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ControlMaster:   "auto",
		ControlPath:     controlPath,
	}
	if err := slave.CreateClient(host, port, user, []ssh.AuthMethod{CreateAuthMethodPassword(password)}); err != nil {
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

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
