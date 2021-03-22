package storage

import "io"

type Storager interface {
	Type() Type
	Read(path string) (io.Reader, error)
	Write(path string, r io.Reader) error
	Delete(path string) error
}

type Type string

var (
	TypeFileSystem Type = "fs"
	TypeOSS        Type = "oss"
)
