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

package issueexcel

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/dop/conf"
	"github.com/erda-project/erda/pkg/excel"
)

type DataForFulfillImportOnlyBaseInfo struct {
	OriginalErdaPlatform  string // get from dop conf.DiceClusterName()
	OriginalErdaProjectID uint64
	AllProjectIssues      bool
}

func (data DataForFulfill) genBaseInfoSheet() (excel.Rows, error) {
	// only one row, k=meta, v=JSON(dataForFulfillImportOnlyBaseInfo)
	meta := DataForFulfillImportOnlyBaseInfo{
		OriginalErdaPlatform:  conf.DiceClusterName(),
		OriginalErdaProjectID: data.ProjectID,
		AllProjectIssues:      data.ExportOnly.AllProjectIssues,
	}
	b, err := json.Marshal(&meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta, err: %v", err)
	}
	var row excel.Row
	row = append(row, excel.NewCell("meta"))
	row = append(row, excel.NewCell(string(b)))
	return excel.Rows{row}, nil
}

func (data *DataForFulfill) decodeBaseInfoSheet(df excel.DecodedFile) (*DataForFulfillImportOnlyBaseInfo, error) {
	if data.IsOldExcelFormat() {
		return nil, nil
	}
	s, ok := df.Sheets.M[nameOfSheetBaseInfo]
	if !ok {
		return nil, nil
	}
	sheet := s.UnmergedSlice
	if len(sheet) != 1 {
		return nil, fmt.Errorf("invalid base info sheet, rows: %d", len(sheet))
	}
	if len(sheet[0]) != 2 {
		return nil, fmt.Errorf("invalid base info sheet, cols: %d", len(sheet[0]))
	}
	if sheet[0][0] != "meta" {
		return nil, fmt.Errorf("invalid base info sheet, first col: %s", sheet[0][0])
	}
	var meta DataForFulfillImportOnlyBaseInfo
	if err := json.Unmarshal([]byte(sheet[0][1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta, err: %v", err)
	}
	// set into data
	data.ImportOnly.BaseInfo = &meta
	return &meta, nil
}
