// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

// 【参考】
//   - https://github.com/tatsushid/minssh/commit/57eae8c5bcf5d94639891f3267f05251f05face4

package sshlib

import (
	"log"
	"os"

	windowsconsole "github.com/moby/term/windows"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/windows"
)

func getStdin() io.ReadCloser {
	h := uint32(windows.STD_INPUT_HANDLE)
	stdin := windowsconsole.NewAnsiReader(int(h))

	return stdin
}
