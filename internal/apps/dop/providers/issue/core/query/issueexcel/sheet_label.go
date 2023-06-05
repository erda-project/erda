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

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) genLabelSheet() (excel.Rows, error) {
	var lines excel.Rows
	// title: label id, label name, label detail (JSON)
	title := excel.Row{
		excel.NewTitleCell("label id"),
		excel.NewTitleCell("label name"),
		excel.NewTitleCell("label detail (json)"),
	}
	lines = append(lines, title)
	// data
	// collect labels from issues
	labelMap := make(map[int64]*pb.ProjectLabel)
	for _, issue := range data.ExportOnly.Issues {
		for _, label := range issue.LabelDetails {
			if _, ok := labelMap[label.Id]; !ok {
				labelMap[label.Id] = label
			}
		}
	}
	for _, label := range labelMap {
		labelInfo, err := json.Marshal(label)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label info, label id: %d, err: %v", label.Id, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(strconv.FormatInt(label.Id, 10)),
			excel.NewCell(label.Name),
			excel.NewCell(string(labelInfo)),
		})
	}

	return lines, nil
}

func (data DataForFulfill) decodeLabelSheet(excelSheets [][][]string) ([]*pb.ProjectLabel, error) {
	if data.IsOldExcelFormat() {
		return nil, nil
	}
	sheet := excelSheets[indexOfSheetLabel]
	// check title
	if len(sheet) < 1 {
		return nil, fmt.Errorf("label sheet is empty")
	}
	var labels []*pb.ProjectLabel
	for _, row := range sheet[1:] {
		var label pb.ProjectLabel
		if err := json.Unmarshal([]byte(row[2]), &label); err != nil {
			return nil, fmt.Errorf("failed to unmarshal label info, label id: %s, err: %v", row[0], err)
		}
		labels = append(labels, &label)
	}
	return labels, nil
}

func (data DataForFulfill) mergeLabelsForCreate(labelsFromLabelSheet []*pb.ProjectLabel, issueModels []IssueSheetModel) []*pb.ProjectLabel {
	labelsFromLabelSheetMap := make(map[string]*pb.ProjectLabel, len(labelsFromLabelSheet))
	for _, label := range labelsFromLabelSheet {
		labelsFromLabelSheetMap[label.Name] = label
	}

	for _, model := range issueModels {
		for _, labelName := range model.Common.Labels {
			if _, ok := labelsFromLabelSheetMap[labelName]; ok {
				continue
			}
			if _, ok := data.LabelMapByName[labelName]; ok {
				continue
			}
			labelsFromLabelSheet = append(labelsFromLabelSheet, &pb.ProjectLabel{
				Name:      labelName,
				Type:      pb.ProjectLabelTypeEnum_issue,
				ProjectID: data.ProjectID,
				Creator:   "system",
			})
		}
	}
	return labelsFromLabelSheet
}

func (data *DataForFulfill) createLabelIfNotExistsForImport(labels []*pb.ProjectLabel) error {
	// create label if not exists
	for _, label := range labels {
		_, ok := data.LabelMapByName[label.Name]
		if ok {
			continue
		}
		// create label
		newLabel := dao.Label{
			Name:      label.Name,
			Type:      apistructs.ProjectLabelType(label.Type.String()),
			Color:     label.Color,
			ProjectID: data.ProjectID,
			Creator:   label.Creator,
		}
		if err := data.ImportOnly.LabelDB.CreateLabel(&newLabel); err != nil {
			return fmt.Errorf("failed to create label, label name: %s, err: %v", label.Name, err)
		}
		// set to label map
		data.LabelMapByName[label.Name] = apistructs.ProjectLabel{
			ID:        newLabel.ID,
			Name:      newLabel.Name,
			Type:      newLabel.Type,
			Color:     newLabel.Color,
			ProjectID: newLabel.ProjectID,
			Creator:   newLabel.Creator,
			CreatedAt: newLabel.CreatedAt,
			UpdatedAt: newLabel.UpdatedAt,
		}
	}
	return nil
}

func (data DataForFulfill) createIssueLabelRelations(issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*IssueSheetModel) error {
	var relations []issuedao.LabelRelation
	for _, issue := range issues {
		model, ok := issueModelMapByIssueID[issue.ID]
		if !ok {
			return fmt.Errorf("issue model not found, issue id: %d", issue.ID)
		}
		for _, labelName := range model.Common.Labels {
			label, ok := data.LabelMapByName[labelName]
			if !ok {
				return fmt.Errorf("label not found, label name: %s", labelName)
			}
			relation := issuedao.LabelRelation{
				LabelID: uint64(label.ID),
				RefType: apistructs.LabelTypeIssue,
				RefID:   strconv.FormatUint(issue.ID, 10),
			}
			relations = append(relations, relation)
		}
	}
	if err := data.ImportOnly.DB.BatchCreateLabelRelations(relations); err != nil {
		return fmt.Errorf("failed to batch create label relations, err: %v", err)
	}
	return nil
}
