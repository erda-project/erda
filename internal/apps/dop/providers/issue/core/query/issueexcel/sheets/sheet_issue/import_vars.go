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

package sheet_issue

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/pkg/excel"
)

type IssueSheetColumnUUID string
type IssueSheetColumnUUIDParts []string

const issueSheetColumnUUIDSplitter = "---"
const uuidPartsMustLength = 3

func NewIssueSheetColumnUUID(parts ...string) IssueSheetColumnUUID {
	return IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}
func (uuid *IssueSheetColumnUUID) Decode() IssueSheetColumnUUIDParts {
	if *uuid == "" {
		return nil
	}
	return strings.Split(string(*uuid), issueSheetColumnUUIDSplitter)
}
func (uuid *IssueSheetColumnUUID) AutoComplete() {
	parts := uuid.Decode()
	if len(parts) > uuidPartsMustLength {
		panic("issue sheet column uuid have max 3 parts")
	}
	if len(parts) == 0 {
		panic("issue sheet column uuid must have at least 1 part")
	}
	// auto fulfill with last elem if not enough to uuidPartsMustLength
	for len(parts) < uuidPartsMustLength {
		parts = append(parts, parts[len(parts)-1])
	}
	// set back to uuid
	*uuid = IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}
func (uuid *IssueSheetColumnUUID) String() string {
	uuid.AutoComplete()
	return string(*uuid)
}
func (uuid *IssueSheetColumnUUID) AddPart(part string) {
	parts := uuid.Decode()
	parts = append(parts, part)
	*uuid = IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}

type IssueSheetModelCellInfoByColumns struct {
	M            IssueSheetModelCellMapByColumns
	OrderedUUIDs []IssueSheetColumnUUID
}

type IssueSheetModelCellMapByColumns map[IssueSheetColumnUUID]excel.Column

// Add
// isDemoModel just used to set all custom fields' uuids in order, can't be added to truly data M
func (info *IssueSheetModelCellInfoByColumns) Add(uuid IssueSheetColumnUUID, cellValue string) {
	uuid.AutoComplete()
	// ordered uuids only add once
	if _, ok := info.M[uuid]; !ok {
		info.OrderedUUIDs = append(info.OrderedUUIDs, uuid)
	}
	info.M[uuid] = append(info.M[uuid], excel.Cell{Value: cellValue})
}

func (info *IssueSheetModelCellInfoByColumns) ConvertToExcelSheet() (excel.Rows, error) {
	// create [][]excel.Cell
	var dataRowLength int
	// get data row length
	for _, columnCells := range info.M {
		dataRowLength = len(columnCells)
		break
	}
	rows := make(excel.Rows, uuidPartsMustLength+dataRowLength)
	for i := range rows {
		rows[i] = make(excel.Row, len(info.M))
	}
	// set by (x,y)
	columnIndex := 0
	for _, uuid := range info.OrderedUUIDs {
		column, ok := info.M[uuid]
		if !ok {
			panic(fmt.Sprintf("uuid: %s not found in info.M", uuid))
		}
		// set column title cells
		parts := uuid.Decode()
		for i, uuidPart := range parts {
			rows[i][columnIndex] = excel.Cell{Value: uuidPart, Style: &excel.CellStyle{IsTitle: true}}
		}
		// auto merge title cells with same value
		autoMergeTitleCellsWithSameValue(rows[:uuidPartsMustLength])
		// set column data cells
		for i, cell := range column {
			rows[uuidPartsMustLength+i][columnIndex] = cell
		}
		columnIndex++
	}
	return rows, nil
}

// excel sheet 聚合时，主动对 title cell 进行探测
// 水平探测:
// - 从右向左探测，如果左侧的 cell 和当前 cell 一致，则左边 cell 的 HMergeNum+1
// - 依次向左
//
// 纵向探测:
// - 从下往上探测，如果上面的 cell 和当前 cell 一致，则上面的 cell 的 VMergeNum+1
func autoMergeTitleCellsWithSameValue(rows excel.Rows) {
	// horizontal merge
	for rowIndex := range rows {
		for columnIndex := len(rows[rowIndex]) - 1; columnIndex >= 0; columnIndex-- {
			if columnIndex == 0 {
				continue
			}
			if rows[rowIndex][columnIndex].Value == rows[rowIndex][columnIndex-1].Value {
				cell := rows[rowIndex][columnIndex]
				cell.HorizontalMergeNum++
				rows[rowIndex][columnIndex-1] = cell
			}
		}
	}
	// vertical merge
	for columnIndex := range rows[0] {
		for rowIndex := len(rows) - 1; rowIndex >= 0; rowIndex-- {
			if rowIndex == 0 {
				continue
			}
			if rows[rowIndex][columnIndex].Value == rows[rowIndex-1][columnIndex].Value {
				cell := rows[rowIndex][columnIndex]
				cell.VerticalMergeNum++
				rows[rowIndex-1][columnIndex] = cell
			}
		}
	}
}
