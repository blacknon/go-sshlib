package sshlib

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// TODO(blacknon):
//     Confで情報を渡していたが、ライブラリ化にあたってもっと汎用的な方法に切り替えたい。
//     Signerの生成を外出しして、別の関数側で生成した認証情報Signerを渡せばいいか？
//     Proxyについてはどうやるかイメージができてない…ちゃんと考える必要あり？？

// Connect structure to store contents about ssh connection.
type Connect struct {
	Client *ssh.Client

	IsTty bool

	ForwardLocal  string
	ForwardRemote string

	ForwardX11 bool

	signers []ssh.Signer
}

func (c *Connect) CreateClient() {}

func (c *Connect) CreateSession() {}

// SendKeepAlive send packet to session.
func (c *Connect) SendKeepAlive(session *ssh.Session) {
	for {
		_, _ = session.SendRequest("keepalive@lssh.com", true, nil)
		time.Sleep(15 * time.Second)
	}
}

// CheckClientAlive check alive ssh.Client.
func (c *Connect) CheckClientAlive() error {
	_, _, err := c.Client.SendRequest("keepalive@lssh.com", true, nil)
	if err == nil || err.Error() == "request failed" {
		return nil
	}
	return err
}
