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

package excel

import (
	"io"
	"io/ioutil"

	"github.com/tealeg/xlsx/v3"
)

// return []sheet{[]row{[]cell}}
// cell 的值即使为空，也可通过下标访问，不会出现越界问题
func Decode(r io.Reader) ([][][]string, error) {
	tmpF, err := ioutil.TempFile("", "excel-")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(tmpF, r); err != nil {
		return nil, err
	}
	// 不适用 xlsx.FileToSliceUnmerged，因为会有重复字段
	data, err := xlsx.FileToSlice(tmpF.Name())
	if err != nil {
		return nil, err
	}
	return data, nil
}
