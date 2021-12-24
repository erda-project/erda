// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tar

import (
	"archive/tar"
	"io"
	"io/fs"
)

// Tar is the handler to tape archive files.
// It is an io.Closer, please note to close it.
type Tar struct {
	w *tar.Writer
}

// New returns the *Tar
func New(buf io.ReadWriter) *Tar {
	w := tar.NewWriter(buf)
	return &Tar{w: w}
}

// Write writes the file into the tar package
func (t *Tar) Write(name string, mode fs.FileMode, data []byte) (int, error) {
	if err := t.w.WriteHeader(&tar.Header{
		Name: name,
		Size: int64(len(data)),
		Mode: int64(mode),
	}); err != nil {
		return 0, err
	}
	return t.w.Write(data)
}

// Close closes the handler
func (t *Tar) Close() error {
	return t.w.Close()
}
