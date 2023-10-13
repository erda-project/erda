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

// DecodeToSheets decode Excel file to Sheets.
// So you can get sheet by sheetName.
func DecodeToSheets(r io.Reader) (*DecodedFile, error) {
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

	df := &DecodedFile{}

	sheets := Sheets{L: make([]*Sheet, 0, len(f.Sheets)), M: make(map[string]*Sheet, len(f.Sheets))}
	for _, sheet := range f.Sheets {
		// iterate sheet cells to set format
		sheet.ForEachRow(func(r *xlsx.Row) error {
			r.ForEachCell(func(c *xlsx.Cell) error {
				if c.IsTime() {
					c.SetFormat("yyyy-mm-dd hh:mm:ss") // the same as golang time format
				}
				return nil
			})
			return nil
		})

		s := NewSheet(sheet, df)

		sheets.L = append(sheets.L, s)
		sheets.M[sheet.Name] = s
	}

	// to slice unmerged after sheet iterated
	fileUnmergedSlice, err := f.ToSliceUnmerged()
	if err != nil {
		return nil, err
	}
	for i, sheet := range sheets.L {
		sheet.UnmergedSlice = fileUnmergedSlice[i]
	}

	df.File = &File{XlsxFile: f, UnmergedSlice: fileUnmergedSlice}
	df.Sheets = sheets

	return df, nil
}
