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

package api

import "fmt"

const pathStatic, pathField int8 = 0, 1

type pathSegment struct {
	typ  int8
	name string
}

func (s *pathSegment) String() string { return fmt.Sprint(*s) }

func buildPathToSegments(path string) (segs []*pathSegment) {
	chars := []rune(path)
	start, i, n := 0, 0, len(chars)
	for ; i < n; i++ {
		if chars[i] == '{' {
			if start < i {
				segs = append(segs, &pathSegment{
					typ:  pathStatic,
					name: string(chars[start:i]),
				})
			}
			i++
			for begin := i; i < n; i++ {
				switch chars[i] {
				case '}':
					segs = append(segs, &pathSegment{
						typ:  pathField,
						name: string(chars[begin:i]),
					})
					start = i + 1
					break
				case '=':
					segs = append(segs, &pathSegment{
						typ:  pathField,
						name: string(chars[begin:i]),
					})
					for ; i < n && chars[i] != '}'; i++ {
					}
					start = i + 1
					break
				}
			}
		}
	}
	if start < n {
		segs = append(segs, &pathSegment{
			typ:  pathStatic,
			name: string(chars[start:]),
		})
	}
	return
}
