package sshlib

import (
	"os/user"
	"path/filepath"
	"testing"

	"golang.org/x/net/proxy"
)

func TestGetAbsPathExpandsHomePrefix(t *testing.T) {
	usr, err := user.Current()
	if err != nil || usr == nil || usr.HomeDir == "" {
		t.Skip("current user home directory is unavailable")
	}

	got := getAbsPath("~/sshlib-test")
	want := filepath.Join(usr.HomeDir, "sshlib-test")
	if got != want {
		t.Fatalf("getAbsPath(%q) = %q, want %q", "~/sshlib-test", got, want)
	}
}

func TestGetAbsPathPreservesTildeInsideWindowsShortNameStylePath(t *testing.T) {
	input := filepath.Join("tmp", "RUNNERC~1", "sshlib")
	got := getAbsPath(input)
	want, err := filepath.Abs(input)
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	if got != want {
		t.Fatalf("getAbsPath(%q) = %q, want %q", input, got, want)
	}
}

func TestCreateHttpProxyDialerIncludesAddressPortAndAuth(t *testing.T) {
	p := &Proxy{
		Type:      "http",
		Addr:      "proxy.example.com",
		Port:      "8080",
		User:      "alice",
		Password:  "secret",
		Forwarder: &stubContextDialer{},
	}

	dialer, err := p.CreateHttpProxyDialer()
	if err != nil {
		t.Fatalf("CreateHttpProxyDialer() error = %v", err)
	}

	httpDialer, ok := dialer.(*httpProxy)
	if !ok {
		t.Fatalf("CreateHttpProxyDialer() type = %T, want *httpProxy", dialer)
	}
	if httpDialer.host != "proxy.example.com:8080" {
		t.Fatalf("CreateHttpProxyDialer() host = %q, want %q", httpDialer.host, "proxy.example.com:8080")
	}
	if !httpDialer.haveAuth || httpDialer.username != "alice" || httpDialer.password != "secret" {
		t.Fatalf("CreateHttpProxyDialer() auth = have:%t user:%q pass:%q", httpDialer.haveAuth, httpDialer.username, httpDialer.password)
	}
	if httpDialer.forward != p.Forwarder {
		t.Fatal("CreateHttpProxyDialer() did not preserve custom forwarder")
	}
}

func TestCreateHttpProxyDialerDefaultsToDirectForwarder(t *testing.T) {
	p := &Proxy{
		Type: "http",
		Addr: "proxy.example.com",
	}

	dialer, err := p.CreateHttpProxyDialer()
	if err != nil {
		t.Fatalf("CreateHttpProxyDialer() error = %v", err)
	}

	httpDialer, ok := dialer.(*httpProxy)
	if !ok {
		t.Fatalf("CreateHttpProxyDialer() type = %T, want *httpProxy", dialer)
	}
	if httpDialer.forward != proxy.Direct {
		t.Fatal("CreateHttpProxyDialer() should default to proxy.Direct forwarder")
	}
}

func TestProxyRoutePortOrDefault(t *testing.T) {
	if got := (ProxyRoute{Type: "ssh"}).portOrDefault(); got != "22" {
		t.Fatalf("portOrDefault() = %q, want %q", got, "22")
	}
	if got := (ProxyRoute{Type: "http", Port: "8080"}).portOrDefault(); got != "8080" {
		t.Fatalf("portOrDefault() = %q, want %q", got, "8080")
	}
}
