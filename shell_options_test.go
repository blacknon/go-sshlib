package sshlib

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestSetLogConfiguresLogging(t *testing.T) {
	c := &Connect{}

	c.SetLog("/tmp/sshlib.log", true)

	if !c.logging {
		t.Fatal("SetLog() should enable logging")
	}
	if c.logFile != "/tmp/sshlib.log" {
		t.Fatalf("SetLog() logFile = %q, want %q", c.logFile, "/tmp/sshlib.log")
	}
	if !c.logTimestamp {
		t.Fatal("SetLog() should enable timestamps")
	}
	if c.logRemoveAnsiCode {
		t.Fatal("SetLog() should not enable ANSI stripping")
	}
}

func TestSetLogWithRemoveAnsiCodeConfiguresLogging(t *testing.T) {
	c := &Connect{}

	c.SetLogWithRemoveAnsiCode("/tmp/sshlib.log", false)

	if !c.logging {
		t.Fatal("SetLogWithRemoveAnsiCode() should enable logging")
	}
	if c.logFile != "/tmp/sshlib.log" {
		t.Fatalf("SetLogWithRemoveAnsiCode() logFile = %q, want %q", c.logFile, "/tmp/sshlib.log")
	}
	if c.logTimestamp {
		t.Fatal("SetLogWithRemoveAnsiCode() should preserve timestamp flag")
	}
	if !c.logRemoveAnsiCode {
		t.Fatal("SetLogWithRemoveAnsiCode() should enable ANSI stripping")
	}
}

func TestLogWritersWritesToOriginalOutputsAndLogFile(t *testing.T) {
	logPath := t.TempDir() + "/ssh.log"
	c := &Connect{logFile: logPath}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	outWriter, errWriter, err := c.logWriters(&stdout, &stderr)
	if err != nil {
		t.Fatalf("logWriters() error = %v", err)
	}

	if _, err := io.WriteString(outWriter, "hello stdout\n"); err != nil {
		t.Fatalf("stdout write error = %v", err)
	}
	if _, err := io.WriteString(errWriter, "hello stderr\n"); err != nil {
		t.Fatalf("stderr write error = %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if got := stdout.String(); got != "hello stdout\n" {
		t.Fatalf("stdout = %q, want %q", got, "hello stdout\n")
	}
	if got := stderr.String(); got != "hello stderr\n" {
		t.Fatalf("stderr = %q, want %q", got, "hello stderr\n")
	}
	logText := string(data)
	if !strings.Contains(logText, "hello stdout\n") || !strings.Contains(logText, "hello stderr\n") {
		t.Fatalf("log file missing expected output: %q", logText)
	}
}

func TestControlSessionOptionsCarriesFlags(t *testing.T) {
	oldTerm, hadTerm := os.LookupEnv("TERM")
	if err := os.Setenv("TERM", "xterm-256color"); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}
	defer func() {
		if hadTerm {
			_ = os.Setenv("TERM", oldTerm)
		} else {
			_ = os.Unsetenv("TERM")
		}
	}()

	c := &Connect{
		ForwardX11:        true,
		ForwardX11Trusted: true,
		ForwardAgent:      true,
	}

	opts := c.controlSessionOptions(true)
	if !opts.TTY {
		t.Fatal("controlSessionOptions() should preserve forced TTY")
	}
	if opts.Term != "xterm-256color" {
		t.Fatalf("controlSessionOptions() term = %q, want %q", opts.Term, "xterm-256color")
	}
	if !opts.ForwardX11 || !opts.ForwardX11Trusted || !opts.ForwardAgent {
		t.Fatalf("controlSessionOptions() flags = %+v", opts)
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		t.Fatalf("controlSessionOptions() size = %dx%d, want positive values", opts.Width, opts.Height)
	}
}

func TestGetDynamicForwardLoggerDefaultsToDiscard(t *testing.T) {
	c := &Connect{}

	logger := c.getDynamicForwardLogger()
	if logger == nil {
		t.Fatal("getDynamicForwardLogger() returned nil")
	}

	var buf bytes.Buffer
	custom := log.New(&buf, "sshlib ", 0)
	c.DynamicForwardLogger = custom
	if got := c.getDynamicForwardLogger(); got != custom {
		t.Fatal("getDynamicForwardLogger() did not return configured logger")
	}
}
