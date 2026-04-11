package sshlib

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestBillyPathFSGetAttrFromMemFS(t *testing.T) {
	fs := memfs.New()
	if err := fs.MkdirAll("dir", 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	f, err := fs.Create("dir/file.txt")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := f.Write([]byte("hello")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	_ = f.Close()

	pfs := newBillyPathFS(fs, "test")
	attr, status := pfs.(*billyPathFS).GetAttr("dir/file.txt", nil)
	if !status.Ok() {
		t.Fatalf("GetAttr() status = %v", status)
	}

	if attr.Size != 5 {
		t.Fatalf("unexpected size: got %d want %d", attr.Size, 5)
	}
	if attr.Mode&uint32(os.ModePerm) != 0666 {
		t.Fatalf("unexpected mode: got %#o want %#o", attr.Mode&uint32(os.ModePerm), 0666)
	}
}

func TestBillyFuseFileReadWriteFallback(t *testing.T) {
	file := &seekOnlyBillyFile{
		name: "test.txt",
		data: []byte("hello"),
	}

	fuseFile := newBillyFuseFile(file).(*billyFuseFile)

	buf := make([]byte, 5)
	result, status := fuseFile.Read(buf, 0)
	if !status.Ok() {
		t.Fatalf("Read() status = %v", status)
	}

	out, status := result.Bytes(buf)
	if !status.Ok() {
		t.Fatalf("ReadResult.Bytes() status = %v", status)
	}
	if string(out) != "hello" {
		t.Fatalf("unexpected read result: got %q want %q", string(out), "hello")
	}

	if written, status := fuseFile.Write([]byte(" world"), 5); !status.Ok() || written != 6 {
		t.Fatalf("Write() = (%d, %v), want (6, OK)", written, status)
	}

	if string(file.data) != "hello world" {
		t.Fatalf("unexpected file contents: got %q want %q", string(file.data), "hello world")
	}
}

type seekOnlyBillyFile struct {
	name string
	data []byte
	pos  int64
}

func (f *seekOnlyBillyFile) Name() string { return f.name }

func (f *seekOnlyBillyFile) Read(p []byte) (int, error) {
	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += int64(n)
	return n, nil
}

func (f *seekOnlyBillyFile) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if off+int64(n) >= int64(len(f.data)) {
		return n, io.EOF
	}
	return n, nil
}

func (f *seekOnlyBillyFile) Write(p []byte) (int, error) {
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

func (f *seekOnlyBillyFile) Seek(offset int64, whence int) (int64, error) {
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

func (f *seekOnlyBillyFile) Close() error { return nil }
func (f *seekOnlyBillyFile) Lock() error  { return nil }
func (f *seekOnlyBillyFile) Unlock() error {
	return nil
}

func (f *seekOnlyBillyFile) Truncate(size int64) error {
	switch {
	case size < int64(len(f.data)):
		f.data = f.data[:size]
	case size > int64(len(f.data)):
		next := make([]byte, size)
		copy(next, f.data)
		f.data = next
	}
	return nil
}

func (f *seekOnlyBillyFile) Stat() (os.FileInfo, error) {
	return seekOnlyFileInfo{name: f.name, size: int64(len(f.data)), mode: 0644, modTime: time.Unix(1710000000, 0)}, nil
}

type seekOnlyFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi seekOnlyFileInfo) Name() string       { return fi.name }
func (fi seekOnlyFileInfo) Size() int64        { return fi.size }
func (fi seekOnlyFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi seekOnlyFileInfo) ModTime() time.Time { return fi.modTime }
func (fi seekOnlyFileInfo) IsDir() bool        { return false }
func (fi seekOnlyFileInfo) Sys() interface{}   { return nil }

var _ fuse.ReadResult = fuse.ReadResultData(nil)
