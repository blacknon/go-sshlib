module github.com/abakum/go-sshlib

// replace github.com/abakum/term => ../term

require (
	github.com/Microsoft/go-winio v0.5.2
	github.com/ScaleFT/sshkeys v0.0.0-20200327173127-6142f742bca5
	// TODO: マージされたらベースのリポジトリに変更する
	github.com/ThalesIgnite/crypto11 v1.2.5
	github.com/abakum/go-ansiterm v0.0.0-20240209124652-4fc46d492442
	github.com/abakum/pageant v0.0.0-20231124135236-c9f79a77a513
	github.com/abakum/term v0.0.0-20240212164236-135562d7e4cf
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/miekg/pkcs11 v1.1.1
	golang.org/x/crypto v0.19.0
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.17.0
	golang.org/x/term v0.17.0
)

require (
	github.com/dchest/bcrypt_pbkdf v0.0.0-20150205184540-83f37f9c154a // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/thales-e-security/pool v0.0.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.21.4

replace github.com/ThalesIgnite/crypto11 v1.2.5 => github.com/blacknon/crypto11 v1.2.6
