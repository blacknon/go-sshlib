package sshlib

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestAutoReconnectConfigDefaults(t *testing.T) {
	c := &Connect{}

	interval, max := c.autoReconnectConfig()
	if interval != time.Second {
		t.Fatalf("autoReconnectConfig() interval = %v, want %v", interval, time.Second)
	}
	if max != 1 {
		t.Fatalf("autoReconnectConfig() max = %d, want %d", max, 1)
	}
}

func TestAutoReconnectConfigOverrides(t *testing.T) {
	c := &Connect{
		AutoReconnectInterval: 3,
		AutoReconnectMax:      5,
	}

	interval, max := c.autoReconnectConfig()
	if interval != 3*time.Second {
		t.Fatalf("autoReconnectConfig() interval = %v, want %v", interval, 3*time.Second)
	}
	if max != 5 {
		t.Fatalf("autoReconnectConfig() max = %d, want %d", max, 5)
	}
}

func TestCanAutoReconnectRequiresConnectionDetails(t *testing.T) {
	c := &Connect{AutoReconnect: true}
	if c.canAutoReconnect() {
		t.Fatal("canAutoReconnect() = true, want false without connection details")
	}

	c.controlHost = "127.0.0.1"
	c.controlPort = "22"
	c.controlUser = "tester"
	c.reconnectAuths = []ssh.AuthMethod{CreateAuthMethodPassword("secret")}
	if !c.canAutoReconnect() {
		t.Fatal("canAutoReconnect() = false, want true with reconnect settings")
	}
}

func TestCanAutoReconnectDisabledForControlClient(t *testing.T) {
	c := &Connect{
		AutoReconnect: true,
		ControlMaster: "auto",
		controlClient: &controlClient{},
		controlHost:   "127.0.0.1",
		controlPort:   "22",
		controlUser:   "tester",
		reconnectAuths: []ssh.AuthMethod{
			CreateAuthMethodPassword("secret"),
		},
	}

	if c.canAutoReconnect() {
		t.Fatal("canAutoReconnect() = true, want false for control client mode")
	}
}

func TestRememberReconnectConfigCopiesAuthMethods(t *testing.T) {
	auths := []ssh.AuthMethod{
		CreateAuthMethodPassword("first"),
		CreateAuthMethodPassword("second"),
	}

	c := &Connect{}
	c.rememberReconnectConfig("host", "2222", "user", auths)

	auths = auths[:1]
	if c.controlHost != "host" || c.controlPort != "2222" || c.controlUser != "user" {
		t.Fatalf("rememberReconnectConfig() stored unexpected target: %#v", c)
	}
	if len(c.reconnectAuths) != 2 {
		t.Fatalf("rememberReconnectConfig() auth count = %d, want %d", len(c.reconnectAuths), 2)
	}
}

func TestEnsureActiveConnectionReturnsNilClientWithoutAutoReconnect(t *testing.T) {
	c := &Connect{}

	err := c.ensureActiveConnection()
	if err == nil {
		t.Fatal("ensureActiveConnection() error = nil, want non-nil")
	}
	if err.Error() != "ssh client is nil" {
		t.Fatalf("ensureActiveConnection() error = %q, want %q", err.Error(), "ssh client is nil")
	}
}

func TestReconnectReturnsNilClientWithoutAutoReconnect(t *testing.T) {
	c := &Connect{}

	err := c.reconnect()
	if err == nil {
		t.Fatal("reconnect() error = nil, want non-nil")
	}
	if err.Error() != "ssh client is nil" {
		t.Fatalf("reconnect() error = %q, want %q", err.Error(), "ssh client is nil")
	}
}

func TestShouldReconnectSession(t *testing.T) {
	c := &Connect{
		AutoReconnect: true,
		controlHost:   "127.0.0.1",
		controlPort:   "22",
		controlUser:   "tester",
		reconnectAuths: []ssh.AuthMethod{
			CreateAuthMethodPassword("secret"),
		},
	}

	if c.shouldReconnectSession(nil) {
		t.Fatal("shouldReconnectSession(nil) = true, want false")
	}

	if !c.shouldReconnectSession(errors.New("connection lost")) {
		t.Fatal("shouldReconnectSession(connection lost) = false, want true when client is nil")
	}
}

func TestShouldReconnectSessionIgnoresExitError(t *testing.T) {
	c := &Connect{
		AutoReconnect: true,
		controlHost:   "127.0.0.1",
		controlPort:   "22",
		controlUser:   "tester",
		reconnectAuths: []ssh.AuthMethod{
			CreateAuthMethodPassword("secret"),
		},
	}

	if c.shouldReconnectSession(&ssh.ExitError{}) {
		t.Fatal("shouldReconnectSession(exit error) = true, want false")
	}
}
