package sshlib

import "golang.org/x/crypto/ssh"

func CreateSignerPassword(password string) (signers []ssh.Signer) {
	signers = append(ssh.Password(password))
	return
}

func CreateSignerPublicKey() {

}

func CreateSignerCertificate() {

}

func CreateSignerPKCS11(provider string, pin string) (signers []ssh.Signer, err error) {

}

func CreateSignerAgent() {

}
