package sshlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
)

type stubWriteCloser struct {
	closeCount int
}

func (s *stubWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (s *stubWriteCloser) Close() error {
	s.closeCount++
	return nil
}

type stubConn struct {
	net.Conn
	closeCount int
}

func (c *stubConn) Close() error {
	c.closeCount++
	return nil
}

func TestControlTerminalStdinWriteAndClose(t *testing.T) {
	var buf bytes.Buffer
	stdin := &controlTerminalStdin{writer: &lockedFrameWriter{w: &buf}}

	n, err := stdin.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 5 {
		t.Fatalf("Write() = %d, want %d", n, 5)
	}
	if err := stdin.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	frameType, payload, err := readStreamFrame(&buf)
	if err != nil {
		t.Fatalf("readStreamFrame(stdout) error = %v", err)
	}
	if frameType != streamFrameStdin || string(payload) != "hello" {
		t.Fatalf("stdin frame = type:%d payload:%q", frameType, payload)
	}

	frameType, payload, err = readStreamFrame(&buf)
	if err != nil {
		t.Fatalf("readStreamFrame(close) error = %v", err)
	}
	if frameType != streamFrameCloseStdin || len(payload) != 0 {
		t.Fatalf("close frame = type:%d payload:%q", frameType, payload)
	}
}

func TestControlTerminalStdinWriteClosedPipe(t *testing.T) {
	stdin := &controlTerminalStdin{}
	if _, err := stdin.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Write() error = %v, want %v", err, io.ErrClosedPipe)
	}
}

func TestTerminalResizeWritesWindowChangeFrame(t *testing.T) {
	var buf bytes.Buffer
	tm := &Terminal{writer: &lockedFrameWriter{w: &buf}}

	if err := tm.Resize(120, 40); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}

	frameType, payload, err := readStreamFrame(&buf)
	if err != nil {
		t.Fatalf("readStreamFrame() error = %v", err)
	}
	if frameType != streamFrameWindowChange {
		t.Fatalf("frameType = %d, want %d", frameType, streamFrameWindowChange)
	}
	if got := binary.BigEndian.Uint32(payload[:4]); got != 120 {
		t.Fatalf("cols = %d, want 120", got)
	}
	if got := binary.BigEndian.Uint32(payload[4:]); got != 40 {
		t.Fatalf("rows = %d, want 40", got)
	}
}

func TestTerminalResizeUsesDefaultDimensionsForNonPositiveValues(t *testing.T) {
	var buf bytes.Buffer
	tm := &Terminal{writer: &lockedFrameWriter{w: &buf}}

	if err := tm.Resize(0, -1); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}

	frameType, payload, err := readStreamFrame(&buf)
	if err != nil {
		t.Fatalf("readStreamFrame() error = %v", err)
	}
	if frameType != streamFrameWindowChange {
		t.Fatalf("frameType = %d, want %d", frameType, streamFrameWindowChange)
	}
	if got := binary.BigEndian.Uint32(payload[:4]); got != 80 {
		t.Fatalf("cols = %d, want 80", got)
	}
	if got := binary.BigEndian.Uint32(payload[4:]); got != 24 {
		t.Fatalf("rows = %d, want 24", got)
	}
}

func TestTerminalWaitNil(t *testing.T) {
	var tm *Terminal
	if err := tm.Wait(); err == nil {
		t.Fatal("Wait() error = nil, want non-nil")
	}
}

func TestOpenTerminalRejectsConflictingOptions(t *testing.T) {
	c := &Connect{}

	_, err := c.OpenTerminal(TerminalOptions{
		StartShell: true,
		Command:    "echo hi",
	})
	if err == nil {
		t.Fatal("OpenTerminal() error = nil, want non-nil")
	}
}

func TestOpenControlTerminalRequiresShellOrCommand(t *testing.T) {
	c := &Connect{
		ControlMaster: "auto",
		controlClient: &controlClient{},
	}

	_, err := c.openControlTerminal(TerminalOptions{})
	if err == nil {
		t.Fatal("openControlTerminal() error = nil, want non-nil")
	}
}

func TestTerminalCopyControlOutput(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	defer stdoutReader.Close()
	defer stderrReader.Close()

	tm := &Terminal{
		conn:   client,
		waitCh: make(chan error, 1),
	}

	go tm.copyControlOutput(stdoutWriter, stderrWriter)
	go func() {
		defer server.Close()
		_ = writeStreamFrame(server, streamFrameStdout, []byte("out"))
		_ = writeStreamFrame(server, streamFrameStderr, []byte("err"))
		_ = writeStreamFrame(server, streamFrameError, []byte("boom"))
		_ = writeStreamFrame(server, streamFrameExit, encodeExitStatus(0))
	}()

	var stdoutData []byte
	var stderrData []byte
	var stdoutErr error
	var stderrErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdoutData, stdoutErr = io.ReadAll(stdoutReader)
	}()
	go func() {
		defer wg.Done()
		stderrData, stderrErr = io.ReadAll(stderrReader)
	}()
	wg.Wait()
	if stdoutErr != nil {
		t.Fatalf("ReadAll(stdout) error = %v", stdoutErr)
	}
	if stderrErr != nil {
		t.Fatalf("ReadAll(stderr) error = %v", stderrErr)
	}
	if string(stdoutData) != "out" {
		t.Fatalf("stdout = %q, want %q", stdoutData, "out")
	}
	if string(stderrData) != "errboom\n" {
		t.Fatalf("stderr = %q, want %q", stderrData, "errboom\n")
	}
	if err := tm.Wait(); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestTerminalCopyControlOutputNonZeroExit(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	defer stdoutReader.Close()
	defer stderrReader.Close()

	tm := &Terminal{
		conn:   client,
		waitCh: make(chan error, 1),
	}

	go tm.copyControlOutput(stdoutWriter, stderrWriter)
	go func() {
		defer server.Close()
		_ = writeStreamFrame(server, streamFrameExit, encodeExitStatus(23))
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.ReadAll(stdoutReader)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.ReadAll(stderrReader)
	}()
	wg.Wait()
	err := tm.Wait()
	var exitErr *controlExitError
	if !errors.As(err, &exitErr) || exitErr.status != 23 {
		t.Fatalf("Wait() error = %v, want controlExitError status 23", err)
	}
}

func TestTerminalCloseIdempotent(t *testing.T) {
	stdin := &stubWriteCloser{}
	conn := &stubConn{}
	tm := &Terminal{
		Stdin: stdin,
		conn:  conn,
	}

	if err := tm.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := tm.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if stdin.closeCount != 1 {
		t.Fatalf("stdin closeCount = %d, want 1", stdin.closeCount)
	}
	if conn.closeCount != 1 {
		t.Fatalf("conn closeCount = %d, want 1", conn.closeCount)
	}
}

func TestTerminalCloseNilIsSafe(t *testing.T) {
	var tm *Terminal
	if err := tm.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}
