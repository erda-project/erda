package storage

import (
	"io"
	"os"
)

type FS struct{}

func NewFS() *FS {
	return &FS{}
}

func (fs *FS) Type() Type {
	return TypeFileSystem
}

func (fs *FS) Read(path string) (io.Reader, error) {
	return os.Open(path)
}

func (fs *FS) Write(path string, r io.Reader) error {
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, r)
	return err
}

func (fs *FS) Delete(path string) error {
	return os.Remove(path)
}
