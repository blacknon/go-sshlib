// Copyright (c) 2021 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.
//go:build windows
// +build windows

package sshlib

import (
	"io"

	termw "github.com/abakum/term/windows"
	"golang.org/x/sys/windows"
)

func GetStdin() io.ReadCloser {
	h := uint32(windows.STD_INPUT_HANDLE)
	stdin := termw.NewAnsiReader(int(h))

	return stdin
}
