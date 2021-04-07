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

package desensitize

import (
	"strings"
	"unicode/utf8"
)

func Mobile(str string) string {
	return desensitize(str, 3, 4)
}

func Email(str string) string {
	parts := strings.SplitN(str, "@", 2)
	if len(parts) == 2 {
		return desensitize(parts[0], 3, 1) + "@" + parts[1]
	}
	return desensitize(str, 3, 1)
}

func Name(str string) string {
	return desensitize(str, 1, 1)
}

func desensitize(str string, before, end int) string {
	l := utf8.RuneCountInString(str)
	r := []rune(str)
	switch l {
	case 0:
		return ""
	case 1:
		return "*"
	case 2:
		return string(r[0:1]) + "*"
	}
	k := l / 3
	// make sure before+end <= l-k (at least k '*')
	if before+k+1 > l {
		before = l - k - 1
	}
	if before+end+k > l {
		end = l - before - k
	}
	var sb strings.Builder
	for i := 0; i < l; i++ {
		if i < before || i >= l-end {
			sb.WriteRune(r[i])
		} else {
			// replace to '*'
			sb.WriteString("*")
		}
	}
	return sb.String()
}
