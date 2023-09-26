// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package excel

import (
	"fmt"
	"io"
	"os"

	"github.com/tealeg/xlsx/v3"
)

// Decode excel file to [][][]string
// return []sheet{[]row{[]cell}}
// cell 的值即使为空，也可通过下标访问，不会出现越界问题
func Decode(r io.Reader) ([][][]string, error) {
	tmpF, err := os.CreateTemp("", "excel-")
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

// DecodeToSheets decode excel file to map[sheetName]sheet.
// So you can get sheet by sheetName.
func DecodeToSheets(r io.Reader) (map[string]Sheet, error) {
	tmpF, err := os.CreateTemp("", "excel-")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(tmpF, r); err != nil {
		return nil, err
	}
	f, err := xlsx.OpenFile(tmpF.Name(), xlsx.ValueOnly(), xlsx.RowLimit(xlsx.NoRowLimit), xlsx.ColLimit(xlsx.NoColLimit))
	if err != nil {
		return nil, err
	}
	sheets := make(map[string]Sheet)
	for _, sheet := range f.Sheets {
		sheets[sheet.Name] = NewSheet(sheet)
	}
	return sheets, nil
}

func (sheet *Sheet) DecodeSheetToSlice() ([][]string, error) {
	output, err := sheet.File.ToSliceUnmerged()
	if err != nil {
		return nil, err
	}
	// get sheet order
	var found bool
	var sheetOrder int
	for i, s := range sheet.File.XlsxFile.Sheets {
		if s.Name == sheet.XlsxSheet.Name {
			sheetOrder = i
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("sheet not found, sheetName: %s", sheet.XlsxSheet.Name)
	}
	return output[sheetOrder], nil
}
