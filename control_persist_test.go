package sshlib

import (
	"strings"
	"testing"
	"time"

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

func TestEncodeDecodeControlPersistPayloadRoundTrip(t *testing.T) {
	payload := controlPersistPayload{
		Host:                "example.com",
		Port:                "22",
		User:                "alice",
		ControlPath:         "/tmp/control.sock",
		ControlPersistNanos: int64(5 * time.Second),
		CheckKnownHosts:     true,
		OverwriteKnownHosts: true,
		KnownHostsFiles:     []string{"/tmp/known_hosts"},
		Auth: []controlPersistAuthMethodDefinition{
			{Type: "password", Password: "secret"},
		},
		ProxyRoute: []controlPersistProxyRoute{
			{Type: "http", Addr: "127.0.0.1", Port: "8080"},
		},
	}

	encoded, err := encodeControlPersistPayload(payload)
	if err != nil {
		t.Fatalf("encodeControlPersistPayload() error = %v", err)
	}

	decoded, err := decodeControlPersistPayload(encoded)
	if err != nil {
		t.Fatalf("decodeControlPersistPayload() error = %v", err)
	}

	if decoded.Host != payload.Host || decoded.Port != payload.Port || decoded.User != payload.User {
		t.Fatalf("decoded payload target = %#v, want %#v", decoded, payload)
	}
	if len(decoded.Auth) != 1 || decoded.Auth[0].Password != "secret" {
		t.Fatalf("decoded auth = %#v", decoded.Auth)
	}
	if len(decoded.ProxyRoute) != 1 || decoded.ProxyRoute[0].Addr != "127.0.0.1" {
		t.Fatalf("decoded proxy route = %#v", decoded.ProxyRoute)
	}
}

func TestDecodeControlPersistPayloadRejectsInvalidBase64(t *testing.T) {
	_, err := decodeControlPersistPayload("%%%")
	if err == nil {
		t.Fatal("decodeControlPersistPayload() error = nil, want non-nil")
	}
}

func TestSerializeControlPersistProxyRoutesRejectsSSHWithoutAuth(t *testing.T) {
	_, err := serializeControlPersistProxyRoutes([]ProxyRoute{
		{Type: "ssh", Addr: "bastion.example.com", User: "jump"},
	})
	if err == nil {
		t.Fatal("serializeControlPersistProxyRoutes() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "requires Auth") {
		t.Fatalf("serializeControlPersistProxyRoutes() error = %v, want requires Auth", err)
	}
}

func TestBuildControlPersistProxyRouteDialerRejectsUnsupportedRouteType(t *testing.T) {
	_, _, err := buildControlPersistProxyRouteDialer([]controlPersistProxyRoute{
		{Type: "gopher"},
	}, nil)
	if err == nil {
		t.Fatal("buildControlPersistProxyRouteDialer() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "unsupported control persist proxy route type") {
		t.Fatalf("buildControlPersistProxyRouteDialer() error = %v, want unsupported route type", err)
	}
}

func TestProxyRouteAuthMethodsNonSSHReturnsNil(t *testing.T) {
	authMethods, err := (ProxyRoute{Type: "http"}).authMethods(nil)
	if err != nil {
		t.Fatalf("authMethods() error = %v", err)
	}
	if authMethods != nil {
		t.Fatalf("authMethods() = %#v, want nil", authMethods)
	}
}

func TestProxyRouteAuthMethodsSSHRequiresAuth(t *testing.T) {
	_, err := (ProxyRoute{Type: "ssh"}).authMethods(nil)
	if err == nil {
		t.Fatal("authMethods() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "requires Auth") {
		t.Fatalf("authMethods() error = %v, want requires Auth", err)
	}
}
