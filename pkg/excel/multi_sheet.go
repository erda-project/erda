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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tealeg/xlsx/v3"
)

type XlsxFile struct {
	*xlsx.File
}

func NewXLSXFile() *XlsxFile {
	return &XlsxFile{File: xlsx.NewFile()}
}

// AddSheetByCell add sheet by cell data. You can add multiple sheets by calling this method multiple times.
func AddSheetByCell(f *XlsxFile, data [][]Cell, sheetName string) error {
	sheet, err := f.File.AddSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
	}
	fulfillCellDataIntoSheet(sheet, data)
	return nil
}

func WriteFile(w io.Writer, f *XlsxFile, filename string) error {
	// set headers to http ResponseWriter `w` before write into `w`.
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Add("Content-Disposition", "attachment;fileName="+url.QueryEscape(filename+".xlsx"))
		rw.Header().Add("Content-Type", "application/vnd.ms-excel")
	}

	var buff bytes.Buffer
	if err := f.File.Write(&buff); err != nil {
		return errors.Errorf("failed to write content, sheetName: %s, err: %v", filename, err)
	}

	if _, err := io.Copy(w, &buff); err != nil {
		return errors.Errorf("failed to copy excel content, err: %v", err)
	}

	return nil
}
