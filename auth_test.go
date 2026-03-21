package sshlib

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestControlPersistAuthResolvedFromPasswordAuthMethod(t *testing.T) {
	authMethod := CreateAuthMethodPassword("secret")

	resolved, err := (&ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}}).resolved()
	if err != nil {
		t.Fatalf("resolved() error = %v", err)
	}

	if len(resolved) != 1 {
		t.Fatalf("len(resolved) = %d, want %d", len(resolved), 1)
	}

	if resolved[0].Type != "password" {
		t.Fatalf("resolved[0].Type = %q, want %q", resolved[0].Type, "password")
	}

	if resolved[0].Password != "secret" {
		t.Fatalf("resolved[0].Password = %q, want %q", resolved[0].Password, "secret")
	}
}

func TestControlPersistAuthResolvedFromAuthMethods(t *testing.T) {
	authMethods := []ssh.AuthMethod{
		CreateAuthMethodPassword("first"),
		CreateAuthMethodPassword("second"),
	}

	resolved, err := (&ControlPersistAuth{AuthMethods: authMethods}).resolved()
	if err != nil {
		t.Fatalf("resolved() error = %v", err)
	}

	if len(resolved) != 2 {
		t.Fatalf("len(resolved) = %d, want %d", len(resolved), 2)
	}

	if resolved[0].Password != "first" {
		t.Fatalf("resolved[0].Password = %q, want %q", resolved[0].Password, "first")
	}

	if resolved[1].Password != "second" {
		t.Fatalf("resolved[1].Password = %q, want %q", resolved[1].Password, "second")
	}
}

func TestControlPersistAuthResolvedRejectsUnknownAuthMethod(t *testing.T) {
	_, err := (&ControlPersistAuth{AuthMethods: []ssh.AuthMethod{ssh.Password("secret")}}).resolved()
	if err == nil {
		t.Fatal("resolved() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unsupported authMethod") {
		t.Fatalf("resolved() error = %v, want unsupported authMethod", err)
	}
}

func TestControlPersistAuthResolvedRejectsEmptyAuthMethods(t *testing.T) {
	_, err := (&ControlPersistAuth{}).resolved()
	if err == nil {
		t.Fatal("resolved() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "AuthMethods is required") {
		t.Fatalf("resolved() error = %v, want AuthMethods is required", err)
	}
}

func TestCreateControlPersistAuthMethodsRejectsUnsupportedType(t *testing.T) {
	_, err := createControlPersistAuthMethods([]controlPersistAuthMethodDefinition{{Type: "unknown"}})
	if err == nil {
		t.Fatal("createControlPersistAuthMethods() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unsupported ControlPersistAuth type") {
		t.Fatalf("createControlPersistAuthMethods() error = %v, want unsupported ControlPersistAuth type", err)
	}
}
