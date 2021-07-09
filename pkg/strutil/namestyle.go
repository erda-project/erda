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

package strutil

import (
	"bytes"
)

// SnakeToUpCamel make a snake style name to up-camel style
func SnakeToUpCamel(name string) string {
	var (
		buf = bytes.NewBuffer(nil)
		big = true
	)
	for _, s := range name {
		if s == '_' {
			big = true
			continue
		}
		if big && s >= 'a' && s <= 'z' {
			buf.WriteRune(s - 32)
		} else {
			buf.WriteRune(s)
		}
		big = false
	}
	return buf.String()
}
