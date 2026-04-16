package sshlib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
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

func TestCreateAuthMethodPublicKeyTransientMarksDefinition(t *testing.T) {
	keyPath := writeTempPrivateKey(t)

	authMethod, err := CreateAuthMethodPublicKeyTransient(keyPath, "")
	if err != nil {
		t.Fatalf("CreateAuthMethodPublicKeyTransient() error = %v", err)
	}

	resolved, err := (&ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}}).resolved()
	if err != nil {
		t.Fatalf("resolved() error = %v", err)
	}

	if len(resolved) != 1 {
		t.Fatalf("len(resolved) = %d, want 1", len(resolved))
	}
	if !resolved[0].Transient {
		t.Fatal("resolved[0].Transient = false, want true")
	}
}

func TestCreateControlPersistAuthMethodsRemovesTransientKeyFile(t *testing.T) {
	keyPath := writeTempPrivateKey(t)

	authMethods, err := createControlPersistAuthMethods([]controlPersistAuthMethodDefinition{{
		Type:      "publickey",
		KeyPath:   keyPath,
		Transient: true,
	}})
	if err != nil {
		t.Fatalf("createControlPersistAuthMethods() error = %v", err)
	}
	if len(authMethods) != 1 {
		t.Fatalf("len(authMethods) = %d, want 1", len(authMethods))
	}

	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Fatalf("transient key file still exists: stat err = %v", err)
	}
}

func writeTempPrivateKey(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "id_rsa")
	if err := os.WriteFile(path, privateKeyPEM, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}
