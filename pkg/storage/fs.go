// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
