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
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/gogap/errors"
	"github.com/tealeg/xlsx/v3"
)

// ExportExcel 导出 excel
// 参数w: 返回http.ResponseWriter
// 参数sheetName: 生成表单的名字
// data数据内容为: data[row][col]，由 title+content组成
// 例子: 第一行: ["name", "age", "city"]
//      第二行: ["excel", "15", "hangzhou"]
// 注：每一行和每一列需要完全对应，根据 row 和 col 导出 xlsx 格式的 excel
func ExportExcel(w io.Writer, data [][]string, sheetName string) error {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)
	if err != nil {
		return errors.Errorf("failed to add sheetName, sheetName: %s", sheetName)
	}

	for row := 0; row < len(data); row++ {
		if len(data[row]) == 0 {
			continue
		}
		rowContent := sheet.AddRow()
		rowContent.SetHeightCM(1)
		for col := 0; col < len(data[row]); col++ {
			cell := rowContent.AddCell()
			cell.Value = data[row][col]
		}
	}

	return write(w, file, sheetName)
}

func write(w io.Writer, file *xlsx.File, sheetName string) error {
	// set headers to http ResponseWriter `w` before write into `w`.
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Add("Content-Disposition", "attachment;fileName="+url.QueryEscape(sheetName+".xlsx"))
		rw.Header().Add("Content-Type", "application/vnd.ms-excel")
	}

	var buff bytes.Buffer
	if err := file.Write(&buff); err != nil {
		return errors.Errorf("failed to write content, sheetName: %s, err: %v", sheetName, err)
	}

	if _, err := io.Copy(w, &buff); err != nil {
		return errors.Errorf("failed to copy excel content, err: %v", err)
	}

	return nil
}
