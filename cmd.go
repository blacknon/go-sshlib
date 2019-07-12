// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import "io"

// Cmd connect and run command over ssh.
//
func (c Connect) Cmd(command string, input chan io.Writer, output chan []byte) {}
