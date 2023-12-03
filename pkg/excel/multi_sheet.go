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
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/tealeg/xlsx/v3"
	"golang.org/x/text/width"
)

type SheetHandler func(sheet *xlsx.Sheet) error

func NewSheetHandlerForDropList(startRow, startCol, endRow, endCol int, dropList []string) SheetHandler {
	return func(sheet *xlsx.Sheet) error {
		dv := xlsx.NewDataValidation(startRow, startCol, endRow, endCol, true)
		ensureDropList(&dropList, dataValidationFormulaStrLen)
		if err := dv.SetDropList(dropList); err != nil {
			return fmt.Errorf("failed to set drop list, err: %v", err)
		}
		dv.SetError(xlsx.StyleInformation, nil, nil)
		sheet.AddDataValidation(dv)
		return nil
	}
}

const (
	dataValidationFormulaStrLen = 257 // see: xlsx#dataValidationFormulaStrLen
)

func ensureDropList(dropList *[]string, maxLimit int) {
	formula := "\"" + strings.Join(*dropList, ",") + "\""
	if maxLimit >= utf8.RuneCountInString(formula) {
		return
	}
	endIndex := len(*dropList) - 1
	if endIndex <= 0 {
		*dropList = []string{}
		return
	}
	*dropList = (*dropList)[:endIndex]
	ensureDropList(dropList, maxLimit)
}

func NewSheetHandlerForTip(startRow, startCol, endRow, endCol int, title, msg string) SheetHandler {
	return func(sheet *xlsx.Sheet) error {
		dv := xlsx.NewDataValidation(startRow, startCol, endRow, endCol, true)
		dv.SetInput(&title, &msg)
		sheet.AddDataValidation(dv)
		return nil
	}
}

func NewSheetHandlerForAutoColWidth(totalColumNum int) SheetHandler {
	return func(sheet *xlsx.Sheet) error {
		// set column index
		for columnIndex := 0; columnIndex < totalColumNum; columnIndex++ {
			err := sheet.SetColAutoWidth(columnIndex+1, func(s string) float64 {
				// calculate proper Excel cell width by string length
				cellWidth := 0
				for _, r := range s {
					w := width.LookupRune(r)
					if w.Kind() == width.EastAsianWide || w.Kind() == width.EastAsianFullwidth {
						cellWidth += 2
					} else {
						cellWidth += 1
					}
				}
				atLeast := float64(9.6)
				max := float64(30)
				calculated := float64(cellWidth) * 2
				if calculated < atLeast {
					return atLeast
				}
				if calculated > max {
					return max
				}
				return calculated
			})
			if err != nil {
				return fmt.Errorf("failed to set column width, err: %v", err)
			}
		}
		return nil
	}
}

// AddSheetByCell add sheet by cell data. You can add multiple sheets by calling this method multiple times.
func AddSheetByCell(f *File, data [][]Cell, sheetName string, handlers ...SheetHandler) error {
	sheet, err := f.XlsxFile.AddSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
	}
	fulfillCellDataIntoSheet(sheet, data)
	// do handler after cell added
	if len(handlers) > 0 {
		for _, handler := range handlers {
			if err := handler(sheet); err != nil {
				return fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
			}
		}
	}
	return nil
}

func WriteFile(w io.Writer, f *File, filename string) error {
	// set headers to http ResponseWriter `w` before write into `w`.
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Add("Content-Disposition", "attachment;fileName="+url.QueryEscape(filename+".xlsx"))
		rw.Header().Add("Content-Type", "application/vnd.ms-excel")
	}

	if err := f.XlsxFile.Write(w); err != nil {
		return errors.Errorf("failed to write content, sheetName: %s, err: %v", filename, err)
	}

	return nil
}
