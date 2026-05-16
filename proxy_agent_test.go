package sshlib

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/ssh/agent"
)

type stubDialer struct {
	conn net.Conn
	err  error
}

func (d *stubDialer) Dial(network, addr string) (net.Conn, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.conn, nil
}

type stubContextDialer struct {
	conn    net.Conn
	err     error
	called  bool
	ctxSeen context.Context
}

func (d *stubContextDialer) Dial(network, addr string) (net.Conn, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.conn, nil
}

func (d *stubContextDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d.called = true
	d.ctxSeen = ctx
	if d.err != nil {
		return nil, d.err
	}
	return d.conn, nil
}

func TestContextDialerDialContextUsesUnderlyingContextDialer(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	dialer := &stubContextDialer{conn: client}
	ctxDialer := &ContextDialer{Dialer: dialer}
	ctx := context.Background()

	conn, err := ctxDialer.DialContext(ctx, "tcp", "example.com:22")
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}
	defer conn.Close()

	if !dialer.called {
		t.Fatal("DialContext() should call underlying DialContext when available")
	}
	if dialer.ctxSeen != ctx {
		t.Fatal("DialContext() did not pass through the original context")
	}
}

func TestContextDialerDialContextFallsBackToDialError(t *testing.T) {
	wantErr := errors.New("dial failed")
	ctxDialer := &ContextDialer{Dialer: &stubDialer{err: wantErr}}

	_, err := ctxDialer.DialContext(context.Background(), "tcp", "example.com:22")
	if !errors.Is(err, wantErr) {
		t.Fatalf("DialContext() error = %v, want %v", err, wantErr)
	}
}

func TestContextDialerDialContextReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ctxDialer := &ContextDialer{Dialer: &blockingDialer{}}

	_, err := ctxDialer.DialContext(ctx, "tcp", "example.com:22")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DialContext() error = %v, want %v", err, context.Canceled)
	}
}

func TestCreateProxyDialerCommandReturnsContextDialer(t *testing.T) {
	p := &Proxy{
		Type:    "command",
		Command: "cat",
	}

	dialer, err := p.CreateProxyDialer()
	if err != nil {
		t.Fatalf("CreateProxyDialer() error = %v", err)
	}

	ctxDialer, ok := dialer.(*ContextDialer)
	if !ok {
		t.Fatalf("CreateProxyDialer() type = %T, want *ContextDialer", dialer)
	}
	if _, ok := ctxDialer.GetDialer().(*NetPipe); !ok {
		t.Fatalf("CreateProxyDialer() inner type = %T, want *NetPipe", ctxDialer.GetDialer())
	}
}

func TestCreateProxyDialerUnsupportedTypeReturnsEmptyContextDialer(t *testing.T) {
	p := &Proxy{Type: "unsupported"}

	dialer, err := p.CreateProxyDialer()
	if err != nil {
		t.Fatalf("CreateProxyDialer() error = %v", err)
	}

	ctxDialer, ok := dialer.(*ContextDialer)
	if !ok {
		t.Fatalf("CreateProxyDialer() type = %T, want *ContextDialer", dialer)
	}
	if ctxDialer.GetDialer() != nil {
		t.Fatalf("CreateProxyDialer() inner dialer = %T, want nil", ctxDialer.GetDialer())
	}
}

func TestConnectSshAgentFallsBackToKeyring(t *testing.T) {
	oldSock, hadSock := os.LookupEnv("SSH_AUTH_SOCK")
	if err := os.Setenv("SSH_AUTH_SOCK", "/path/that/does/not/exist.sock"); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}
	defer func() {
		if hadSock {
			_ = os.Setenv("SSH_AUTH_SOCK", oldSock)
		} else {
			_ = os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	ag := ConnectSshAgent()
	if _, ok := ag.(agent.Agent); !ok {
		t.Fatalf("ConnectSshAgent() type = %T, want agent.Agent fallback", ag)
	}
}

func TestEnsureSshAgentInitializesAgent(t *testing.T) {
	oldSock, hadSock := os.LookupEnv("SSH_AUTH_SOCK")
	if err := os.Setenv("SSH_AUTH_SOCK", "/path/that/does/not/exist.sock"); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}
	defer func() {
		if hadSock {
			_ = os.Setenv("SSH_AUTH_SOCK", oldSock)
		} else {
			_ = os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	c := &Connect{}
	c.ensureSshAgent()

	if c.Agent == nil {
		t.Fatal("ensureSshAgent() left Agent nil")
	}
}

func TestEnsureSshAgentPreservesExistingAgent(t *testing.T) {
	existing := agent.NewKeyring()
	c := &Connect{Agent: existing}

	c.ensureSshAgent()

	if c.Agent != existing {
		t.Fatal("ensureSshAgent() replaced existing Agent")
	}
}

func TestAddKeySshAgentAddsKeyToKeyring(t *testing.T) {
	keyring := agent.NewKeyring()
	c := &Connect{}

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	c.AddKeySshAgent(keyring, key)

	keys, err := keyring.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("List() count = %d, want %d", len(keys), 1)
	}
}

type blockingDialer struct{}

func (d *blockingDialer) Dial(network, addr string) (net.Conn, error) {
	time.Sleep(100 * time.Millisecond)
	return nil, errors.New("dial should have been canceled")
}
