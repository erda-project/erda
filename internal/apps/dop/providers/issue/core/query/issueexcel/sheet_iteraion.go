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
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) genIterationSheet() (excel.Rows, error) {
	var lines excel.Rows
	// title: iteration id, iteration name, iteration info (JSON)
	title := excel.Row{
		excel.NewTitleCell("iteration id"),
		excel.NewTitleCell("iteration name"),
		excel.NewTitleCell("iteration detail (json)"),
	}
	lines = append(lines, title)
	// data
	for _, iteration := range data.IterationMapByID {
		iteration := iteration
		if iteration.ID <= 0 {
			continue
		}
		b, err := json.Marshal(iteration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal iteration info, iteration id: %d, err: %v", iteration.ID, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(strconv.FormatUint(iteration.ID, 10)),
			excel.NewCell(iteration.Title),
			excel.NewCell(string(b)),
		})
	}

	return lines, nil
}

func (data *DataForFulfill) decodeIterationSheet(excelSheets [][][]string) ([]*dao.Iteration, error) {
	if data.IsOldExcelFormat() {
		return nil, nil
	}
	sheet := excelSheets[indexOfSheetIteration]
	var iterations []*dao.Iteration
	for i, row := range sheet {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return nil, fmt.Errorf("invalid iteration sheet, row: %d, len(row): %d", i, len(row))
		}
		var iteration dao.Iteration
		if err := json.Unmarshal([]byte(row[2]), &iteration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal iteration info, row: %d, err: %v", i, err)
		}
		iterations = append(iterations, &iteration)
	}
	return iterations, nil
}

// createIterationsIfNotExistForImport create iterations if not exist for import.
// check by name:
// - if not exist, create new iteration
// - if exist, do not update, take the existing one as the standard
func (data *DataForFulfill) createIterationsIfNotExistForImport(originalProjectIterations []*dao.Iteration, issueSheetModels []IssueSheetModel) error {
	iterationsNeedCreate := make(map[string]*dao.Iteration) // only created sheet related iterations
	now := time.Now()
	currentDayBegin := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	currentDayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	for _, originalProjectIteration := range originalProjectIterations {
		originalProjectIteration := originalProjectIteration
		_, ok := data.IterationMapByName[originalProjectIteration.Title]
		if ok {
			continue
		}
		if _, ok := iterationsNeedCreate[originalProjectIteration.Title]; ok {
			continue
		}
		// create
		iterationsNeedCreate[originalProjectIteration.Title] = originalProjectIteration
	}
	for _, issueSheetModel := range issueSheetModels {
		_, ok := data.IterationMapByName[issueSheetModel.Common.IterationName]
		if ok {
			continue
		}
		switch issueSheetModel.Common.IterationName {
		case "待规划", "待办事项", "待处理":
			continue
		}
		newIteration := dao.Iteration{
			ProjectID:  data.ProjectID,
			Title:      issueSheetModel.Common.IterationName,
			State:      apistructs.IterationStateUnfiled,
			StartedAt:  &currentDayBegin,
			FinishedAt: &currentDayEnd,
		}
		if _, ok := iterationsNeedCreate[newIteration.Title]; ok {
			continue
		}
		iterationsNeedCreate[newIteration.Title] = &newIteration
	}
	for _, iteration := range iterationsNeedCreate {
		// create iteration if not exist
		iteration.ID = 0
		iteration.ProjectID = data.ProjectID
		if err := data.ImportOnly.DB.CreateIteration(iteration); err != nil {
			return fmt.Errorf("failed to create iteration, iteration: %+v, err: %v", iteration, err)
		}
		// add to iteration map
		data.IterationMapByID[int64(iteration.ID)] = iteration
		data.IterationMapByName[iteration.Title] = iteration
	}
	return nil
}
