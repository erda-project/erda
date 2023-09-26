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
	"sync"

	"github.com/tealeg/xlsx/v3"
)

type Row []Cell
type Rows []Row
type Column []Cell
type Sheet struct {
	XlsxSheet *xlsx.Sheet
	File      File
}
type File struct {
	XlsxFile      *xlsx.File
	UnmergedSlice [][][]string

	decoded bool
	sync.Mutex
}

func NewFile() *File {
	return &File{XlsxFile: xlsx.NewFile()}
}

func NewSheet(sheet *xlsx.Sheet) Sheet {
	return Sheet{XlsxSheet: sheet, File: File{XlsxFile: sheet.File}}
}

func (f *File) ToSliceUnmerged() ([][][]string, error) {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if f.decoded {
		return f.UnmergedSlice, nil
	}
	result, err := f.XlsxFile.ToSliceUnmerged()
	if err != nil {
		return nil, err
	}
	f.decoded = true
	f.UnmergedSlice = result

	return result, nil
}
