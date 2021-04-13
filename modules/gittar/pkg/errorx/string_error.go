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

package errorx

import "fmt"

// StringError 简单的字符串型错误
type StringError string

// New 创建一个具有文本信息的错误
func New(text string) error {
	return StringError(text)
}

// Errorf 创建一个具有文本信息的错误
func Errorf(format string, args ...interface{}) error {
	return StringError(fmt.Sprintf(format, args...))
}

func (err StringError) Error() string {
	return string(err)
}
