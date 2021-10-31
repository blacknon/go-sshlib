module github.com/blacknon/go-sshlib

require (
	github.com/ScaleFT/sshkeys v0.0.0-20200327173127-6142f742bca5
	// github.com/ThalesIgnite/crypto11 v0.1.0
	github.com/ThalesIgnite/crypto11 v1.2.5
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/lunixbochs/vtclean v1.0.0
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/miekg/pkcs11 v1.0.3
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/net v0.0.0-20201026091529-146b70c837a4
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sys v0.0.0-20201026173827-119d4633e4d1 // indirect
)

replace github.com/miekg/pkcs11 => github.com/blacknon/pkcs11 v1.0.4-0.20201018135904-6038e308f617

go 1.15
