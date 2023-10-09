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
	"github.com/tealeg/xlsx/v3"
)

type Row []Cell
type Rows []Row
type Column []Cell

type File struct {
	XlsxFile *xlsx.File

	UnmergedSlice [][][]string
}

type DecodedFile struct {
	File   *File
	Sheets Sheets
}

type Sheets struct {
	L []*Sheet
	M map[string]*Sheet
}

type Sheet struct {
	XlsxSheet *xlsx.Sheet
	RefFile   *File

	UnmergedSlice [][]string
}

func NewFile() *File {
	return &File{XlsxFile: xlsx.NewFile()}
}

func NewSheet(sheet *xlsx.Sheet) *Sheet {
	return &Sheet{XlsxSheet: sheet, RefFile: &File{XlsxFile: sheet.File}}
}

func EmptyDecodedFile() DecodedFile { return DecodedFile{} }
func EmptySheets() Sheets           { return Sheets{L: make([]*Sheet, 0), M: make(map[string]*Sheet)} }
