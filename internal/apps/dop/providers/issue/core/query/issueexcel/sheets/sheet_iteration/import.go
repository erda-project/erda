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

package sheet_iteration

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/excel"
)

type Handler struct{ sheets.DefaultImporter }

func (h *Handler) SheetName() string { return vars.NameOfSheetIteration }

func (h *Handler) DecodeSheet(data *vars.DataForFulfill, s *excel.Sheet) error {
	if data.IsOldExcelFormat() {
		return nil
	}
	sheet := s.UnmergedSlice
	var iterations []*dao.Iteration
	for i, row := range sheet {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return fmt.Errorf("invalid iteration sheet, row: %d, len(row): %d", i, len(row))
		}
		var iteration dao.Iteration
		if err := json.Unmarshal([]byte(row[2]), &iteration); err != nil {
			return fmt.Errorf("failed to unmarshal iteration info, row: %d, err: %v", i, err)
		}
		iterations = append(iterations, &iteration)
	}
	data.ImportOnly.Sheets.Optional.IterationInfo = iterations

	return nil
}

func (h *Handler) Precheck(data *vars.DataForFulfill) {
	checkIssueSheetModelIterations(data)
}

func (h *Handler) BeforeCreateIssues(data *vars.DataForFulfill) error {
	// create iterations if not exists before issue create
	if err := createIterationsIfNotExistForImport(data, data.ImportOnly.Sheets.Optional.IterationInfo); err != nil {
		return fmt.Errorf("failed to create iterations, err: %v", err)
	}
	return nil
}

func checkIssueSheetModelIterations(data *vars.DataForFulfill) {
	for i := range data.ImportOnly.Sheets.Must.IssueInfo {
		model := &data.ImportOnly.Sheets.Must.IssueInfo[i]
		iterationName := model.Common.IterationName
		// check in current project
		if _, ok := data.IterationMapByName[iterationName]; ok {
			continue
		}
		// check in iteration sheet (waiting to create)
		for _, iteration := range data.ImportOnly.Sheets.Optional.IterationInfo {
			if iteration.Title == iterationName {
				continue
			}
		}
		// check as default iteration
		switch iterationName {
		case "待规划", "待办事项", "待处理":
			continue
		}
		// use default if not found
		defaultIterationName := data.IterationMapByID[-1].Title
		model.Common.IterationName = defaultIterationName
	}
}

// createIterationsIfNotExistForImport create iterations if not exist for import.
// check by name:
// - if not exist, create new iteration
// - if exist, do not update, take the existing one as the standard
func createIterationsIfNotExistForImport(data *vars.DataForFulfill, originalProjectIterations []*dao.Iteration) error {
	iterationsNeedCreate := make(map[string]*dao.Iteration) // only created sheet related iterations
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
	// create by order
	order := getOrderedIterationsTitles(iterationsNeedCreate)
	for _, title := range order {
		iteration := iterationsNeedCreate[title]
		if iteration.Title == "" {
			continue
		}
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

// 规则
// 1. 根据 iteration id 排序，id 小的在前，id 大的在后，0 在最后
// 2. 当 id = 0 时，按照 title 字典序排序，字典序小的在前，大的在后
func getOrderedIterationsTitles(m map[string]*dao.Iteration) []string {
	var (
		zeroIterations            []string
		greaterThanZeroIterations []string
	)
	for title := range m {
		if m[title].ID == 0 {
			zeroIterations = append(zeroIterations, title)
		} else {
			greaterThanZeroIterations = append(greaterThanZeroIterations, title)
		}
	}
	// 1. 根据 iteration id 排序，id 小的在前，id 大的在后，0 在最后
	sort.SliceStable(greaterThanZeroIterations, func(i, j int) bool {
		return m[greaterThanZeroIterations[i]].ID < m[greaterThanZeroIterations[j]].ID
	})
	// 2. 当 id = 0 时，按照 title 字典序排序，字典序小的在前，大的在后
	sort.SliceStable(zeroIterations, func(i, j int) bool {
		return zeroIterations[i] < zeroIterations[j]
	})
	return append(greaterThanZeroIterations, zeroIterations...)
}
