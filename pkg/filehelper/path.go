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
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const apiFile string = "/api/files"

func Abs2Rel(path string) string {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "/") {
		path = fmt.Sprintf(".%s", path)
	}
	return filepath.Clean(path)
}

func FileUrlRetriever(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	if strings.HasPrefix(u.Path, apiFile) {
		return u.Path
	}
	return path
}

func FilterFilePath(content string) string {
	r := regexp.MustCompile(`\(([^)]+)\)`)
	for _, sub := range r.FindAllStringSubmatch(content, -1) {
		path := FileUrlRetriever(sub[1])
		if path != sub[1] {
			content = strings.Replace(content, sub[1], path, 1)
		}
	}
	return content
}
