package sshlib

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
)

type recordingChangeFS struct {
	chmodName  string
	chmodMode  os.FileMode
	chownName  string
	chownUID   int
	chownGID   int
	lchownName string
	lchownUID  int
	lchownGID  int
	chtimesName string
	atime      time.Time
	mtime      time.Time
	err        error
}

func (fs *recordingChangeFS) Chmod(name string, mode os.FileMode) error {
	fs.chmodName = name
	fs.chmodMode = mode
	return fs.err
}

func (fs *recordingChangeFS) Lchown(name string, uid, gid int) error {
	fs.lchownName = name
	fs.lchownUID = uid
	fs.lchownGID = gid
	return fs.err
}

func (fs *recordingChangeFS) Chown(name string, uid, gid int) error {
	fs.chownName = name
	fs.chownUID = uid
	fs.chownGID = gid
	return fs.err
}

func (fs *recordingChangeFS) Chtimes(name string, atime, mtime time.Time) error {
	fs.chtimesName = name
	fs.atime = atime
	fs.mtime = mtime
	return fs.err
}

func TestNewChangeSFTPFSImplementsBillyChange(t *testing.T) {
	fs := NewChangeSFTPFS(nil, "/srv/share")
	if _, ok := fs.(billy.Change); !ok {
		t.Fatal("NewChangeSFTPFS() should return a filesystem that implements billy.Change")
	}
}

func TestChangeChrootFSChangePath(t *testing.T) {
	fs := &changeChrootFS{root: "/srv/share"}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "dot maps to root", input: ".", want: "/srv/share"},
		{name: "slash maps to root", input: "/", want: "/srv/share"},
		{name: "relative path joins root", input: "nested/file.txt", want: "/srv/share/nested/file.txt"},
		{name: "leading slash is trimmed", input: "/nested/file.txt", want: "/srv/share/nested/file.txt"},
		{name: "cross boundary is rejected", input: "../outside", wantErr: billy.ErrCrossedBoundary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.changePath(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("changePath(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("changePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestChangeChrootFSForwardsChmod(t *testing.T) {
	change := &recordingChangeFS{}
	fs := &changeChrootFS{root: "/srv/share", change: change}

	if err := fs.Chmod("nested/file.txt", 0o640); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	if change.chmodName != "/srv/share/nested/file.txt" {
		t.Fatalf("Chmod() forwarded path = %q, want %q", change.chmodName, "/srv/share/nested/file.txt")
	}
	if change.chmodMode != 0o640 {
		t.Fatalf("Chmod() forwarded mode = %#o, want %#o", change.chmodMode, 0o640)
	}
}

func TestChangeChrootFSForwardsOwnershipAndTimes(t *testing.T) {
	change := &recordingChangeFS{}
	fs := &changeChrootFS{root: "/srv/share", change: change}
	atime := time.Unix(100, 0)
	mtime := time.Unix(200, 0)

	if err := fs.Chown("file.txt", 1000, 1001); err != nil {
		t.Fatalf("Chown() error = %v", err)
	}
	if err := fs.Lchown("link.txt", 2000, 2001); err != nil {
		t.Fatalf("Lchown() error = %v", err)
	}
	if err := fs.Chtimes("time.txt", atime, mtime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	if change.chownName != "/srv/share/file.txt" || change.chownUID != 1000 || change.chownGID != 1001 {
		t.Fatalf("Chown() forwarded unexpected values: path=%q uid=%d gid=%d", change.chownName, change.chownUID, change.chownGID)
	}
	if change.lchownName != "/srv/share/link.txt" || change.lchownUID != 2000 || change.lchownGID != 2001 {
		t.Fatalf("Lchown() forwarded unexpected values: path=%q uid=%d gid=%d", change.lchownName, change.lchownUID, change.lchownGID)
	}
	if change.chtimesName != "/srv/share/time.txt" || !change.atime.Equal(atime) || !change.mtime.Equal(mtime) {
		t.Fatalf("Chtimes() forwarded unexpected values: path=%q atime=%v mtime=%v", change.chtimesName, change.atime, change.mtime)
	}
}

func TestChangeChrootFSRejectsCrossBoundaryChanges(t *testing.T) {
	change := &recordingChangeFS{}
	fs := &changeChrootFS{root: "/srv/share", change: change}

	err := fs.Chmod("../outside", 0o644)
	if !errors.Is(err, billy.ErrCrossedBoundary) {
		t.Fatalf("Chmod() error = %v, want %v", err, billy.ErrCrossedBoundary)
	}
	if change.chmodName != "" {
		t.Fatalf("Chmod() should not forward cross-boundary path, got %q", change.chmodName)
	}
}
