package sshlib

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func testPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()

	keyPath := writeTempPrivateKey(t)
	signer, err := CreateSignerPublicKey(keyPath, "")
	if err != nil {
		t.Fatalf("CreateSignerPublicKey() error = %v", err)
	}
	return signer.PublicKey()
}

func TestWriteKnownHostsKeyAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	key := testPublicKey(t)
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := writeKnownHostsKey(path, 0, "example.com", remote, key); err != nil {
		t.Fatalf("writeKnownHostsKey() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := strings.TrimSpace(string(data))
	want := knownhosts.Line([]string{"example.com", remote.String()}, key)
	if text != want {
		t.Fatalf("known_hosts entry = %q, want %q", text, want)
	}
}

func TestWriteKnownHostsKeyOverwriteLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	original := "old line\nkeep line\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	key := testPublicKey(t)
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := writeKnownHostsKey(path, 1, "example.com", remote, key); err != nil {
		t.Fatalf("writeKnownHostsKey() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	wantFirst := knownhosts.Line([]string{"example.com", remote.String()}, key)
	if len(lines) != 2 || lines[0] != wantFirst || lines[1] != "keep line" {
		t.Fatalf("known_hosts lines = %#v", lines)
	}
}

func TestWriteKnownHostsKeyUsesSingleAddressWhenRemoteMatchesHostname(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	key := testPublicKey(t)
	remote := staticAddr{network: "tcp", address: "example.com"}
	if err := writeKnownHostsKey(path, 0, "example.com", remote, key); err != nil {
		t.Fatalf("writeKnownHostsKey() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := strings.TrimSpace(string(data))
	want := knownhosts.Line([]string{"example.com"}, key)
	if text != want {
		t.Fatalf("known_hosts entry = %q, want %q", text, want)
	}
}

func TestAskAddingUnknownHostKeyAcceptsShortYesAfterRetry(t *testing.T) {
	key := testPublicKey(t)
	remote := staticAddr{network: "tcp", address: "127.0.0.1:22"}

	oldInput := knownHostsPromptInput
	oldOutput := knownHostsPromptOutput
	oldWatcher := startKnownHostsSignalWatcher
	defer func() {
		knownHostsPromptInput = oldInput
		knownHostsPromptOutput = oldOutput
		startKnownHostsSignalWatcher = oldWatcher
	}()

	var output bytes.Buffer
	knownHostsPromptInput = strings.NewReader("maybe\ny\n")
	knownHostsPromptOutput = &output
	startKnownHostsSignalWatcher = func(stopC <-chan struct{}) {}

	ok, err := askAddingUnknownHostKey("Host {{.Address}} {{.Fingerprint}}? ", "example.com", remote, key)
	if err != nil {
		t.Fatalf("askAddingUnknownHostKey() error = %v", err)
	}
	if !ok {
		t.Fatal("askAddingUnknownHostKey() = false, want true")
	}
	text := output.String()
	if !strings.Contains(text, "Host example.com ") {
		t.Fatalf("prompt output = %q, want rendered template", text)
	}
	if !strings.Contains(text, "Please type 'yes' or 'no': ") {
		t.Fatalf("prompt output = %q, want retry prompt", text)
	}
}

func TestAskOverwriteKnownHostKeyRejectsNo(t *testing.T) {
	key := testPublicKey(t)
	remote := staticAddr{network: "tcp", address: "127.0.0.1:22"}

	oldInput := knownHostsPromptInput
	oldOutput := knownHostsPromptOutput
	oldWatcher := startKnownHostsSignalWatcher
	defer func() {
		knownHostsPromptInput = oldInput
		knownHostsPromptOutput = oldOutput
		startKnownHostsSignalWatcher = oldWatcher
	}()

	var output bytes.Buffer
	knownHostsPromptInput = strings.NewReader("n\n")
	knownHostsPromptOutput = &output
	startKnownHostsSignalWatcher = func(stopC <-chan struct{}) {}

	ok, err := askOverwriteKnownHostKey("Old={{.OldKeyText}} ", "example.com", remote, key, "old-key")
	if err != nil {
		t.Fatalf("askOverwriteKnownHostKey() error = %v", err)
	}
	if ok {
		t.Fatal("askOverwriteKnownHostKey() = true, want false")
	}
	if !strings.Contains(output.String(), "Old=old-key") {
		t.Fatalf("prompt output = %q, want old key text", output.String())
	}
}
