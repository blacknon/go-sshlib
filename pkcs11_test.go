// Copyright (c) 2026 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build cgo
// +build cgo

package sshlib

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestC11GetPINUsesCustomPrompt(t *testing.T) {
	var gotPrompt string
	c := &C11{
		Label: "token-a",
		Prompt: func(prompt string) (string, error) {
			gotPrompt = prompt
			return "1234", nil
		},
	}

	if err := c.getPIN(); err != nil {
		t.Fatalf("getPIN() error = %v", err)
	}
	if gotPrompt != "token-a's PIN:" {
		t.Fatalf("getPIN() prompt = %q, want %q", gotPrompt, "token-a's PIN:")
	}
	if c.PIN != "1234" {
		t.Fatalf("getPIN() PIN = %q, want %q", c.PIN, "1234")
	}
}

func TestC11GetPINSkipsPromptWhenPINAlreadySet(t *testing.T) {
	called := false
	c := &C11{
		Label: "token-b",
		PIN:   "preset",
		Prompt: func(prompt string) (string, error) {
			called = true
			return "", nil
		},
	}

	if err := c.getPIN(); err != nil {
		t.Fatalf("getPIN() error = %v", err)
	}
	if called {
		t.Fatal("getPIN() should not call prompt when PIN is already set")
	}
	if c.PIN != "preset" {
		t.Fatalf("getPIN() PIN = %q, want %q", c.PIN, "preset")
	}
}

func TestCreateSignerPKCS11WithPromptMissingProvider(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing-provider.so")

	signers, err := CreateSignerPKCS11WithPrompt(missingPath, "", func(prompt string) (string, error) {
		t.Fatalf("prompt should not be called for missing provider, got %q", prompt)
		return "", nil
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("CreateSignerPKCS11WithPrompt() error = %v, want %v", err, os.ErrNotExist)
	}
	if len(signers) != 0 {
		t.Fatalf("CreateSignerPKCS11WithPrompt() signers len = %d, want %d", len(signers), 0)
	}
}

func TestCreateAuthMethodPKCS11WithPromptMissingProvider(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing-provider.so")

	auth, err := CreateAuthMethodPKCS11WithPrompt(missingPath, "", func(prompt string) (string, error) {
		t.Fatalf("prompt should not be called for missing provider, got %q", prompt)
		return "", nil
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("CreateAuthMethodPKCS11WithPrompt() error = %v, want %v", err, os.ErrNotExist)
	}
	if len(auth) != 0 {
		t.Fatalf("CreateAuthMethodPKCS11WithPrompt() auth len = %d, want %d", len(auth), 0)
	}
}
