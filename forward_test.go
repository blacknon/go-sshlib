package sshlib_test

import (
	"log"

	"github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

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
	authMethod := sshlib.CreateAuthMethodPublicKey(key, "")

	// Create sshlib.Connect
	con := &sshlib.Connect{}

	// PortForward
	con.TCPForward(localAddr, remoteAddr)

	// Connect ssh server
	con.CreateClient(host, user, port, []ssh.AuthMethod{authMethod})
}

func ExampleConnect_X11Forward() {
	// Create session
	session, err := c.CreateSession()
	if err != nil {
		return
	}

	// X11 forwarding
	err = c.X11Forward(session)
	if err != nil {
		log.Fatal(err)
	}
}
