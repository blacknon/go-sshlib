package sshlib

import (
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
)

func TestDefaultSMBShareName(t *testing.T) {
	if got := defaultSMBShareName(""); got != "share" {
		t.Fatalf("defaultSMBShareName(\"\") = %q, want %q", got, "share")
	}
	if got := defaultSMBShareName("docs"); got != "docs" {
		t.Fatalf("defaultSMBShareName(\"docs\") = %q, want %q", got, "docs")
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
