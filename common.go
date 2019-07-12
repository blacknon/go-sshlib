// Copyright (c) 2019 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"os/user"
	"path/filepath"
	"strings"
)

// getAbsPath return absolute path convert.
// Replace `~` with your home directory.
func getAbsPath(path string) string {
	// Replace home directory
	usr, _ := user.Current()
	path = strings.Replace(path, "~", usr.HomeDir, 1)

	path, _ = filepath.Abs(path)
	return path
}
