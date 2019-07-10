package sshlib

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// TODO(blacknon):
//     Confで情報を渡していたが、ライブラリ化にあたってもっと汎用的な方法に切り替えたい。
//     Signerの生成を外出しして、Proxyの作成については認証情報のMapを渡せばいいのか？？
//     特にProxyについてはちゃんと考える必要あり？？

// Connect structure to store contents about ssh connection.
type Connect struct {
	Client *ssh.Client

	ForwardLocal  string
	ForwardRemote string

	ForwardX11 bool
}

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
