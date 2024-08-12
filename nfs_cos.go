// Copyright (c) 2024 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package sshlib

import (
	"os"
	"time"

	"github.com/go-git/go-billy/v5"
)

// NewChangeOSFS wraps billy osfs to add the change interface
func NewChangeOSFS(fs billy.Filesystem) billy.Filesystem {
	return COS{fs}
}

// COS or OSFS + Change wraps a billy.FS to not fail the `Change` interface.
type COS struct {
	billy.Filesystem
}

// Chmod changes mode
func (fs COS) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(fs.Join(fs.Root(), name), mode)
}

// Lchown changes ownership
func (fs COS) Lchown(name string, uid, gid int) error {
	return os.Lchown(fs.Join(fs.Root(), name), uid, gid)
}

// Chown changes ownership
func (fs COS) Chown(name string, uid, gid int) error {
	return os.Chown(fs.Join(fs.Root(), name), uid, gid)
}

// Chtimes changes access time
func (fs COS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(fs.Join(fs.Root(), name), atime, mtime)
}
