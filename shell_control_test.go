package sshlib

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestCopyControlOutputReturnsUnexpectedEOF(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := (&Connect{}).copyControlOutput(client, &stdout, &stderr, nil)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("copyControlOutput() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestCopyControlOutputReadsExitFrame(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		if err := writeStreamFrame(server, streamFrameStdout, []byte("shared")); err != nil {
			t.Errorf("write stdout frame: %v", err)
			return
		}
		if err := writeStreamFrame(server, streamFrameExit, encodeExitStatus(0)); err != nil {
			t.Errorf("write exit frame: %v", err)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := (&Connect{}).copyControlOutput(client, &stdout, &stderr, nil); err != nil {
		t.Fatalf("copyControlOutput() error = %v", err)
	}

	if got := stdout.String(); got != "shared" {
		t.Fatalf("stdout = %q, want %q", got, "shared")
	}
}

func TestShouldSuppressControlStreamErrorForInteractiveRequests(t *testing.T) {
	exitErr := &ssh.ExitError{}
	exitMissingErr := &ssh.ExitMissingError{}

	if !shouldSuppressControlStreamError(controlRequest{Type: controlRequestShell}, exitErr) {
		t.Fatal("shell exit error should be suppressed")
	}

	if !shouldSuppressControlStreamError(controlRequest{Type: controlRequestCmdShell}, exitErr) {
		t.Fatal("cmdshell exit error should be suppressed")
	}

	if shouldSuppressControlStreamError(controlRequest{Type: controlRequestCommand}, exitErr) {
		t.Fatal("command exit error should not be suppressed")
	}

	if !shouldSuppressControlStreamError(controlRequest{Type: controlRequestCmdShell}, exitMissingErr) {
		t.Fatal("cmdshell exit missing error should be suppressed")
	}
}
