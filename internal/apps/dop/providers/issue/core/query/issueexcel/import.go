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
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/excel"
)

func ImportFile(r io.Reader, data DataForFulfill) error {
	// decode to sheets
	df, err := excel.DecodeToSheets(r)
	if err != nil {
		return fmt.Errorf("failed to decode excel, err: %v", err)
	}
	// compatible
	data.JudgeIfIsOldExcelFormat(df)

	// base info sheet first
	if err := data.decodeBaseInfoSheet(df); err != nil {
		return fmt.Errorf("failed to decode base info sheet, err: %v", err)
	}
	// issue sheet
	if err := data.DecodeIssueSheet(df); err != nil {
		return fmt.Errorf("failed to decode issue sheet, err: %v", err)
	}
	// user sheet
	if err := data.decodeUserSheet(df); err != nil {
		return fmt.Errorf("failed to decode user sheet, err: %v", err)
	}
	if err := data.mapMemberForImport(data.ImportOnly.Sheets.Optional.UserInfo); err != nil {
		return fmt.Errorf("failed to map member, err: %v", err)
	}
	// label sheet
	if err := data.decodeLabelSheet(df); err != nil {
		return fmt.Errorf("failed to decode label sheet, err: %v", err)
	}
	mergedLabels := data.mergeLabelsForCreate(data.ImportOnly.Sheets.Optional.LabelInfo)
	if err := data.createLabelIfNotExistsForImport(mergedLabels); err != nil {
		return fmt.Errorf("failed to create label, err: %v", err)
	}
	// custom field sheet
	if err := data.decodeCustomFieldSheet(df); err != nil {
		return fmt.Errorf("failed to decode custom field sheet, err: %v", err)
	}
	if err := data.createCustomFieldIfNotExistsForImport(data.ImportOnly.Sheets.Optional.CustomFieldInfo); err != nil {
		return fmt.Errorf("failed to create custom field, err: %v", err)
	}
	// iteration sheet
	if err := data.decodeIterationSheet(df); err != nil {
		return fmt.Errorf("failed to decode iteration sheet, err: %v", err)
	}
	// create iterations if not exists before issue create
	if err := data.createIterationsIfNotExistForImport(data.ImportOnly.Sheets.Optional.IterationInfo); err != nil {
		return fmt.Errorf("failed to create iterations, err: %v", err)
	}
	// state sheet
	if err := data.decodeStateSheet(df); err != nil {
		return fmt.Errorf("failed to decode state sheet, err: %v", err)
	}
	if err := data.syncState(data.ImportOnly.Sheets.Optional.StateInfo); err != nil {
		return fmt.Errorf("failed to sync custom issue state, err: %v", err)
	}

	// 先创建或更新所有 issues，再创建或更新所有关联关系

	// 创建或更新 issues
	// 更新 model 里的相关关联 ID 字段，比如 L1 转换为具体的 ID
	issues, issueModelMapByIssueID, err := data.createOrUpdateIssues(data.ImportOnly.Sheets.Must.IssueInfo)
	if err != nil {
		return fmt.Errorf("failed to create or update issues, err: %v", err)
	}

	// 先将数据进行合并，以 label 为例:
	// - 收集 issue 里的 label
	// - 与 label sheet 里的 label 进行合并
	// - 创建或更新 label
	// - 创建或更新关联 issue 与 label 的关联关系

	// create label relation
	if err := data.createIssueLabelRelations(issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue label relations, err: %v", err)
	}
	// create custom field relation
	if err := data.createIssueCustomFieldRelation(issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue custom field relations, err: %v", err)
	}
	// create issue relation
	if err := data.createIssueRelations(issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue relations, err: %v", err)
	}
	return nil
}

func changePointerTimeToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func mustGetJsonManHour(estimateTime string) string {
	manHour, err := NewManhour(estimateTime)
	if err != nil {
		panic(fmt.Errorf("failed to get man hour from estimate time, err: %v", err))
	}
	b, _ := json.Marshal(&manHour)
	return string(b)
}

func getIssueStage(model IssueSheetModel) string {
	if model.Common.IssueType == pb.IssueTypeEnum_TASK {
		return model.TaskOnly.TaskType
	}
	if model.Common.IssueType == pb.IssueTypeEnum_BUG {
		return model.BugOnly.Source
	}
	return ""
}

var estimateRegexp, _ = regexp.Compile(`(\d+)([wdhm]?)`)

func NewManhour(manhour string) (pb.IssueManHour, error) {
	if manhour == "" {
		return pb.IssueManHour{}, nil
	}
	if !estimateRegexp.MatchString(manhour) {
		return pb.IssueManHour{}, fmt.Errorf("invalid estimate time: %s", manhour)
	}
	matches := estimateRegexp.FindAllStringSubmatch(manhour, -1)
	var totalMinutes int64
	for _, match := range matches {
		timeVal, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return pb.IssueManHour{}, fmt.Errorf("invalid man hour: %s, err: %v", manhour, err)
		}
		timeType := match[2]
		switch timeType {
		case "m":
			totalMinutes += int64(timeVal)
		case "h":
			totalMinutes += int64(timeVal) * 60
		case "d":
			totalMinutes += int64(timeVal) * 60 * 8
		case "w":
			totalMinutes += int64(timeVal) * 60 * 8 * 5
		default:
			return pb.IssueManHour{}, fmt.Errorf("invalid man hour: %s", manhour)
		}
	}
	return pb.IssueManHour{EstimateTime: totalMinutes, RemainingTime: totalMinutes}, nil
}
