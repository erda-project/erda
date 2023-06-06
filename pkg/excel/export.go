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
	"time"

	"github.com/sirupsen/logrus"
)

// ExportExcel 导出 excel
// 参数w: 返回http.ResponseWriter
// 参数sheetName: 生成表单的名字
// data数据内容为: data[row][col]，由 title+content组成
// 例子: 第一行: ["name", "age", "city"]
//
//	第二行: ["excel", "15", "hangzhou"]
//
// 注：每一行和每一列需要完全对应，根据 row 和 col 导出 xlsx 格式的 excel
func ExportExcel(w io.Writer, data [][]string, sheetName string) error {
	// convert data to cells
	cells := convertStringDataToCellData(data)
	// invoke ExportExcelByCell
	return ExportExcelByCell(w, cells, sheetName)
}

// ExportExcelByCell 支持 cell 粒度配置，可以实现单元格合并，样式调整等
func ExportExcelByCell(w io.Writer, data [][]Cell, sheetName string) error {
	begin := time.Now()
	defer func() {
		end := time.Now()
		logrus.Debugf("export excel cost: %fs", end.Sub(begin).Seconds())
	}()

	file := NewXLSXFile()
	if err := AddSheetByCell(file, data, sheetName); err != nil {
		return fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
	}
	return WriteFile(w, file, sheetName)
}
