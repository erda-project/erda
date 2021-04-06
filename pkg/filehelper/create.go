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

package filehelper

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// CreateFile create file if not exist, or truncate file and then write content to file.
func CreateFile(absPath, content string, perm os.FileMode) error {
	return CreateFile2(absPath, bytes.NewBufferString(content), perm)
}

func CreateFile2(absPath string, r io.Reader, perm os.FileMode) error {
	f, err := CreateFile3(absPath, r, perm)
	if err != nil {
		return err
	}
	f.Close()
	return err
}

// CreateFile3 return the created file and error if have
func CreateFile3(absPath string, r io.Reader, perm os.FileMode) (*os.File, error) {
	if !filepath.IsAbs(absPath) {
		return nil, errors.Errorf("not an absolute path: %s", absPath)
	}
	err := os.MkdirAll(filepath.Dir(absPath), 0755)
	if err != nil {
		return nil, errors.Wrap(err, "make parent dir error")
	}
	f, err := os.OpenFile(filepath.Clean(absPath), os.O_CREATE|os.O_TRUNC|os.O_RDWR, perm)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return nil, errors.Wrap(err, "write content to file error")
	}
	return f, nil
}
