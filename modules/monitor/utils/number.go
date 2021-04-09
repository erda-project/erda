// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"bytes"
	"strconv"
	"strings"
)

// ParseInt64 .
func ParseInt64(value string, def int64) int64 {
	if num, err := strconv.ParseInt(value, 10, 64); err == nil {
		return num
	}
	return def
}

// ParseInt64WithRadix .
func ParseInt64WithRadix(value string, def int64, radix int) int64 {
	if num, err := strconv.ParseInt(value, radix, 64); err == nil {
		return num
	}
	return def
}

// IsMobile .
func IsMobile(ua string) bool {
	ua = strings.ToLower(ua)
	return ua == "ios" || ua == "android"
}

// GetPath .
func GetPath(s string) string {
	return ReplaceNumber(s)
}

// ReplaceNumber .
func ReplaceNumber(path string) string {
	parts := strings.Split(path, "/")
	last := len(parts) - 1
	var buffer bytes.Buffer
	for i, item := range parts {
		if _, err := strconv.ParseInt(item, 10, 64); err == nil {
			buffer.WriteString("{number}")
		} else {
			buffer.WriteString(item)
		}
		if i != last {
			buffer.WriteString("/")
		}
	}
	return buffer.String()
}

// // ReadLine .
// func ReadLine(fileName string, handler func(string)) error {
// 	f, err := os.Open(fileName)
// 	if err != nil {
// 		return err
// 	}
// 	buf := bufio.NewReader(f)
// 	for {
// 		line, err := buf.ReadString('\n')
// 		if err != nil {
// 			if err == io.EOF {
// 				return nil
// 			}
// 			return err
// 		}
// 		line = strings.TrimSpace(line)
// 		handler(line)
// 	}
// }
