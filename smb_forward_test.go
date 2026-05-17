package sshlib

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
)

func TestDefaultSMBShareName(t *testing.T) {
	if got := defaultSMBShareName(""); got != "share" {
		t.Fatalf("defaultSMBShareName(\"\") = %q, want %q", got, "share")
	}
	if got := defaultSMBShareName("   "); got != "share" {
		t.Fatalf("defaultSMBShareName(\"   \") = %q, want %q", got, "share")
	}
	if got := defaultSMBShareName("docs"); got != "docs" {
		t.Fatalf("defaultSMBShareName(\"docs\") = %q, want %q", got, "docs")
	}
}

func TestAbsCleanPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		cwd  string
		want string
	}{
		{name: "empty path uses cwd", path: "", cwd: "/work", want: "/work"},
		{name: "relative path joins cwd", path: "dir/file.txt", cwd: "/work", want: "/work/dir/file.txt"},
		{name: "parent segments are cleaned", path: "../other/file.txt", cwd: "/work/base", want: "/work/other/file.txt"},
		{name: "absolute path is preserved", path: "/var/data", cwd: "/work", want: "/var/data"},
		{name: "windows separators are normalized", path: `dir\sub\file.txt`, cwd: "/work", want: "/work/dir/sub/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := absCleanPath(tt.path, tt.cwd); got != tt.want {
				t.Fatalf("absCleanPath(%q, %q) = %q, want %q", tt.path, tt.cwd, got, tt.want)
			}
		})
	}
}

func TestBillyPathAndPathHelpers(t *testing.T) {
	if got := billyPath("/"); got != "." {
		t.Fatalf("billyPath(\"/\") = %q, want %q", got, ".")
	}
	if got := billyPath("/dir/file.txt"); got != "dir/file.txt" {
		t.Fatalf("billyPath(\"/dir/file.txt\") = %q, want %q", got, "dir/file.txt")
	}
	if got := pathClean("/dir/./nested/../file.txt"); got != "/dir/file.txt" {
		t.Fatalf("pathClean() = %q, want %q", got, "/dir/file.txt")
	}
	if got := pathJoin("/dir", "nested", "..", "file.txt"); got != "/dir/file.txt" {
		t.Fatalf("pathJoin() = %q, want %q", got, "/dir/file.txt")
	}
}

func TestGetRemoteAbsPath(t *testing.T) {
	tests := []struct {
		name string
		wd   string
		path string
		want string
	}{
		{name: "tilde expands to home", wd: "/home/test", path: "~/repo", want: "/home/test/repo"},
		{name: "relative path joins home", wd: "/home/test", path: "repo", want: "/home/test/repo"},
		{name: "absolute path stays absolute", wd: "/home/test", path: "/srv/share", want: "/srv/share"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRemoteAbsPath(tt.wd, tt.path); got != tt.want {
				t.Fatalf("getRemoteAbsPath(%q, %q) = %q, want %q", tt.wd, tt.path, got, tt.want)
			}
		})
	}
}

func TestAbsBillyFSReadDirAndReadFile(t *testing.T) {
	bfs := memfs.New()
	if err := bfs.MkdirAll("dir", 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	f, err := bfs.Create("dir/file.txt")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := f.Write([]byte("hello")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	_ = f.Close()

	afs := newAbsBillyFS(bfs)

	data, err := afs.ReadFile("/dir/file.txt")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected data: got %q want %q", string(data), "hello")
	}

	entries, err := afs.ReadDir("/dir")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "file.txt" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestAbsBillyFileWriteAtFallback(t *testing.T) {
	file := &seekWriteBillyFile{
		name: "test.txt",
		data: []byte("hello"),
	}

	af := newAbsBillyFile(file)
	n, err := af.WriteAt([]byte(" world"), 5)
	if err != nil {
		t.Fatalf("WriteAt() error = %v", err)
	}
	if n != 6 {
		t.Fatalf("WriteAt() wrote %d, want %d", n, 6)
	}
	if string(file.data) != "hello world" {
		t.Fatalf("unexpected data: got %q want %q", string(file.data), "hello world")
	}
}

func TestAbsBillyFSChdirAndGetwd(t *testing.T) {
	bfs := memfs.New()
	if err := bfs.MkdirAll("dir/sub", 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	afs := newAbsBillyFS(bfs).(*absBillyFS)
	if err := afs.Chdir("dir/sub"); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	wd, err := afs.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if wd != "/dir/sub" {
		t.Fatalf("Getwd() = %q, want %q", wd, "/dir/sub")
	}
	if got := afs.clean("../file.txt"); got != filepath.ToSlash("dir/file.txt") {
		t.Fatalf("clean() = %q, want %q", got, "dir/file.txt")
	}
}

type seekWriteBillyFile struct {
	name string
	data []byte
	pos  int64
}

func (f *seekWriteBillyFile) Name() string { return f.name }

func (f *seekWriteBillyFile) Read(p []byte) (int, error) {
	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += int64(n)
	return n, nil
}

func (f *seekWriteBillyFile) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if off+int64(n) >= int64(len(f.data)) {
		return n, io.EOF
	}
	return n, nil
}

func (f *seekWriteBillyFile) Write(p []byte) (int, error) {
	end := f.pos + int64(len(p))
	if end > int64(len(f.data)) {
		next := make([]byte, end)
		copy(next, f.data)
		f.data = next
	}
	copy(f.data[f.pos:end], p)
	f.pos = end
	return len(p), nil
}

func (f *seekWriteBillyFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos += offset
	case io.SeekEnd:
		f.pos = int64(len(f.data)) + offset
	default:
		return 0, os.ErrInvalid
	}
	return f.pos, nil
}

func (f *seekWriteBillyFile) Close() error                       { return nil }
func (f *seekWriteBillyFile) Lock() error                        { return nil }
func (f *seekWriteBillyFile) Unlock() error                      { return nil }
func (f *seekWriteBillyFile) Truncate(size int64) error          { return nil }
func (f *seekWriteBillyFile) Sync() error                        { return nil }
func (f *seekWriteBillyFile) Readdir(int) ([]os.FileInfo, error) { return nil, fs.ErrInvalid }

func (f *seekWriteBillyFile) Stat() (os.FileInfo, error) {
	return smbTestFileInfo{name: f.name, size: int64(len(f.data)), mode: 0644, modTime: time.Unix(1710000000, 0)}, nil
}

type smbTestFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi smbTestFileInfo) Name() string       { return fi.name }
func (fi smbTestFileInfo) Size() int64        { return fi.size }
func (fi smbTestFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi smbTestFileInfo) ModTime() time.Time { return fi.modTime }
func (fi smbTestFileInfo) IsDir() bool        { return false }
func (fi smbTestFileInfo) Sys() interface{}   { return nil }
