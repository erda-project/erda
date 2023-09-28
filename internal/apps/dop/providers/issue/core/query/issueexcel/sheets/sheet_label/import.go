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

package sheet_label

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type Handler struct{}

func (h *Handler) SheetName() string { return vars.NameOfSheetLabel }

func (h *Handler) ImportSheet(data *vars.DataForFulfill, df excel.DecodedFile) error {
	if data.IsOldExcelFormat() {
		return nil
	}
	s, ok := df.Sheets.M[h.SheetName()]
	if !ok {
		return nil
	}
	sheet := s.UnmergedSlice
	// check title
	if len(sheet) < 1 {
		return fmt.Errorf("label sheet is empty")
	}
	var labels []*pb.ProjectLabel
	for _, row := range sheet[1:] {
		var label pb.ProjectLabel
		if err := json.Unmarshal([]byte(row[2]), &label); err != nil {
			return fmt.Errorf("failed to unmarshal label info, label id: %s, err: %v", row[0], err)
		}
		labels = append(labels, &label)
	}
	data.ImportOnly.Sheets.Optional.LabelInfo = labels

	// merge labels for create
	mergedLabels := mergeLabelsForCreate(data, data.ImportOnly.Sheets.Optional.LabelInfo)
	if err := createLabelIfNotExistsForImport(data, mergedLabels); err != nil {
		return fmt.Errorf("failed to create label, err: %v", err)
	}
	return nil
}

func mergeLabelsForCreate(data *vars.DataForFulfill, labelsFromLabelSheet []*pb.ProjectLabel) []*pb.ProjectLabel {
	labelsFromLabelSheetMap := make(map[string]*pb.ProjectLabel, len(labelsFromLabelSheet))
	for _, label := range labelsFromLabelSheet {
		labelsFromLabelSheetMap[label.Name] = label
	}

	for _, model := range data.ImportOnly.Sheets.Must.IssueInfo {
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

func createLabelIfNotExistsForImport(data *vars.DataForFulfill, labels []*pb.ProjectLabel) error {
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
			Color:     strutil.FirstNoneEmpty(label.Color, randomColor()),
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

const (
	ColorRed         = "red"
	ColorBlue        = "blue"
	ColorOrange      = "orange"
	ColorGreen       = "green"
	ColorGray        = "gray"
	ColorYellow      = "yellow"
	ColorPurple      = "purple"
	ColorWaterBlue   = "water-blue"
	ColorMagenta     = "magenta"
	ColorCyan        = "cyan"
	ColorYellowGreen = "yellow-green"
)

func randomColor() string {
	colors := []string{
		ColorRed,
		ColorBlue,
		ColorOrange,
		ColorGreen,
		ColorGray,
		ColorYellow,
		ColorPurple,
		ColorWaterBlue,
		ColorMagenta,
		ColorCyan,
		ColorYellowGreen,
	}
	return colors[rand.Intn(len(colors))]
}

func CreateIssueLabelRelations(data *vars.DataForFulfill, issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*vars.IssueSheetModel) error {
	// batch delete label relations
	var issueIDs []uint64
	for _, issue := range issues {
		issueIDs = append(issueIDs, issue.ID)
	}
	if err := data.ImportOnly.DB.BatchDeleteIssueLabelRelations(issueIDs); err != nil {
		return fmt.Errorf("failed to batch delete issue label relations before create label relations, err: %v", err)
	}

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
