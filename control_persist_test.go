package sshlib

import (
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

func TestControlPersistAuthResolvedFromMethodsPKCS11(t *testing.T) {
	resolved, err := (&ControlPersistAuth{
		Methods: []ControlPersistAuthMethod{
			{Type: "pkcs11", PKCS11Provider: "/usr/local/lib/opensc-pkcs11.so"},
		},
	}).resolved()
	if err != nil {
		t.Fatalf("resolved() error = %v", err)
	}

	if len(resolved) != 1 || resolved[0].Type != "pkcs11" || resolved[0].PKCS11Provider != "/usr/local/lib/opensc-pkcs11.so" {
		t.Fatalf("unexpected resolved auth = %#v", resolved)
	}
}

func TestSerializeControlPersistProxyRoutesWithPKCS11(t *testing.T) {
	def, err := serializeControlPersistProxyRoutes([]ProxyRoute{
		{
			Type: "ssh",
			Addr: "bastion.example.com",
			User: "jump",
			Auth: &ControlPersistAuth{
				Methods: []ControlPersistAuthMethod{
					{Type: "pkcs11", PKCS11Provider: "/opt/pkcs11.so"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("serializeControlPersistProxyRoutes() error = %v", err)
	}

	if len(def) != 1 || len(def[0].Auth) != 1 || def[0].Auth[0].Type != "pkcs11" {
		t.Fatalf("unexpected definition = %#v", def)
	}
}

func TestBuildControlPersistProxyRouteDialerCommandHTTP(t *testing.T) {
	dialer, connects, err := buildControlPersistProxyRouteDialer([]controlPersistProxyRoute{
		{Type: "command", Command: "ssh -W %h:%p bastion"},
		{Type: "http", Addr: "127.0.0.1", Port: "8080"},
	}, nil)
	if err != nil {
		t.Fatalf("buildControlPersistProxyRouteDialer() error = %v", err)
	}
	if len(connects) != 0 {
		t.Fatalf("len(connects) = %d, want 0", len(connects))
	}
	if _, ok := dialer.(proxy.ContextDialer); !ok {
		t.Fatal("dialer does not implement proxy.ContextDialer")
	}
}

func TestControlPersistAuthResolvedFromAuthMethodsBackwardCompatible(t *testing.T) {
	authMethod := CreateAuthMethodPassword("secret")
	resolved, err := (&ControlPersistAuth{AuthMethods: []ssh.AuthMethod{authMethod}}).resolved()
	if err != nil {
		t.Fatalf("resolved() error = %v", err)
	}
	if len(resolved) != 1 || resolved[0].Password != "secret" {
		t.Fatalf("unexpected resolved auth = %#v", resolved)
	}
}
