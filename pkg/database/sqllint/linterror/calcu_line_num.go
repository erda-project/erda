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

package linterror

import (
	"bufio"
	"bytes"
)

// CalcLintLine calculates the lint error line number of the SQL script
func CalcLintLine(source, scope []byte, goal func(line []byte) bool) (line string, num int) {
	var firstLine []byte
	scanner := bufio.NewScanner(bytes.NewBuffer(scope))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimPrefix(line, []byte(" "))) > 0 {
			firstLine = line
			break
		}
	}

	idx := bytes.Index(source, scope)
	if idx < 0 {
		return "", -1
	}
	var (
		i       = 0
		lineNum = 0
		minNum  = 0
	)
	scanner = bufio.NewScanner(bytes.NewBuffer(source))
	for scanner.Scan() {
		line := scanner.Bytes()
		i += len(line) + 1
		lineNum++
		if i < idx {
			continue
		}
		if bytes.Equal(line, firstLine) {
			minNum = lineNum
		}
		if goal(line) && !bytes.HasPrefix(bytes.TrimSpace(line), []byte("--")) {
			return string(line), lineNum
		}
	}
	return string(firstLine), minNum
}
