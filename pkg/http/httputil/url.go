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

package httputil

import (
	"net/url"
	"strings"
)

// JoinPath .
func JoinPath(appendRoot bool, segments ...string) string {
	sb := &strings.Builder{}
	if appendRoot {
		sb.WriteString("/")
	}
	last := len(segments) - 1
	for i, s := range segments {
		sb.WriteString(url.PathEscape(s))
		if i < last {
			sb.WriteRune('/')
		}
	}
	return sb.String()
}

// JoinPathR .
func JoinPathR(segments ...string) string {
	return JoinPath(true)
}
