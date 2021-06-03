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

package testapis

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed *.html
var staticFiles embed.FS

// FileSystemWithPrefix .
func FileSystemWithPrefix(prefix string) http.FileSystem {
	fs := http.FS(staticFiles)
	if len(prefix) <= 0 {
		return fs
	}
	return &prefixFS{
		fs:     fs,
		prefix: prefix,
	}
}

type prefixFS struct {
	fs     http.FileSystem
	prefix string
}

func (fs *prefixFS) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, fs.prefix) {
		name = name[len(fs.prefix):]
		if len(name) == 0 {
			name = "/"
		}
	}
	return fs.fs.Open(name)
}
