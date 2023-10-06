package sshlib

import (
	"log"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGetDisplay(t *testing.T) {

	for _, tc := range []struct {
		expect int
		input  string
	}{
		{0, ":0.0"},
		{123, ":123.0"},
		{123, ":123"},
		{0, "xxx"},
		{11, "localhost:11.0"},
		{123, "randomhost:123.0"},
	} {
		if act := getX11Display(tc.input); act != tc.expect {
			t.Errorf(`unexpected result for getX11Display("%s"), act="%v", exp="%v"`, tc.input, act, tc.expect)
		}
	}
}

func ExampleConnect_TCPForward() {
	// host
	host := "target.com"
	port := "22"
	user := "user"
	key := "~/.ssh/id_rsa"

	// port forwarding
	localAddr := "localhost:10022"
	remoteAddr := "localhost:22"

	// Create ssh.AuthMethod
	authMethod, _ := CreateAuthMethodPublicKey(key, "")

	// Create sshlib.Connect
	con := &Connect{}

	// PortForward
	con.TCPLocalForward(localAddr, remoteAddr)

	// Connect ssh server
	con.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
}

func ExampleConnect_X11Forward() {
	// Create session
	con := &Connect{}
	session, err := con.CreateSession()
	if err != nil {
		return
	}

	// X11 forwarding
	err = con.X11Forward(session)
	if err != nil {
		log.Fatal(err)
	}
}
