module github.com/blacknon/go-sshlib

require (
	github.com/ScaleFT/sshkeys v0.0.0-20200327173127-6142f742bca5
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	// TODO: マージされたらベースのリポジトリに変更する
	// github.com/ThalesIgnite/crypto11 v1.2.6
	github.com/blacknon/crypto11 v1.2.7
	github.com/blacknon/go-x11auth v0.1.0
	github.com/lunixbochs/vtclean v1.0.0
	github.com/miekg/pkcs11 v1.1.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	golang.org/x/crypto v0.13.0
	golang.org/x/net v0.15.0
	golang.org/x/sys v0.19.0
	golang.org/x/term v0.19.0
)

require (
	github.com/go-git/go-billy v4.2.0+incompatible
	github.com/go-git/go-billy/v5 v5.5.0
	github.com/pkg/sftp v1.13.6
	github.com/willscott/go-nfs v0.0.2
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/dchest/bcrypt_pbkdf v0.0.0-20150205184540-83f37f9c154a // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rasky/go-xdr v0.0.0-20170124162913-1a41d1a06c93 // indirect
	github.com/thales-e-security/pool v0.0.2 // indirect
	github.com/willscott/go-nfs-client v0.0.0-20240104095149-b44639837b00 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
)

go 1.22.4

// NOTE: アップデートされたらベースのリポジトリに変更する(2024/08/09)
replace github.com/willscott/go-nfs v0.0.2 => ../go-nfs
