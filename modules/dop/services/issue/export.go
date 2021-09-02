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

package issue

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/tealeg/xlsx/v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type issueStage struct {
	Type  apistructs.IssueType
	Value string
}

func (svc *Issue) ExportExcel(issues []apistructs.Issue, properties []apistructs.IssuePropertyIndex, projectID uint64, isDownload bool, orgID int64) (io.Reader, string, error) {
	// list of  issue stage
	stages, err := svc.db.GetIssuesStageByOrgID(orgID)
	if err != nil {
		return nil, "", err
	}
	// get the stageMap
	stageMap := svc.getStageMap(stages)

	table, err := svc.convertIssueToExcelList(issues, properties, projectID, isDownload, stageMap)
	if err != nil {
		return nil, "", err
	}
	// replace userids by usernames
	userids := []string{}
	for _, t := range table[1:] {
		if t[4] != "" {
			userids = append(userids, t[4])
		}
		if t[5] != "" {
			userids = append(userids, t[5])
		}
		if t[6] != "" {
			userids = append(userids, t[6])
		}
	}
	userids = strutil.DedupSlice(userids, true)
	users, err := svc.uc.FindUsers(userids)
	if err != nil {
		return nil, "", err
	}
	usernames := map[string]string{}
	for _, u := range users {
		usernames[u.ID] = u.Nick
	}
	for i := 1; i < len(table); i++ {
		if table[i][4] != "" {
			if name, ok := usernames[table[i][4]]; ok {
				table[i][4] = name
			}
		}
		if table[i][5] != "" {
			if name, ok := usernames[table[i][5]]; ok {
				table[i][5] = name
			}
		}
		if table[i][6] != "" {
			if name, ok := usernames[table[i][6]]; ok {
				table[i][6] = name
			}
		}
	}
	tablename := "issuetable"
	if len(issues) > 0 {
		if issues[0].IterationID == -1 {
			tablename = "待办事项"
		} else {
			tablename = issues[0].Type.GetZhName()
		}
	}

	buf := bytes.NewBuffer([]byte{})
	if err := excel.ExportExcel(buf, table, tablename); err != nil {
		return nil, "", err
	}
	return buf, tablename, nil
}

// getStageMap return a map,the key is the struct of dice_issue_stage.Value and dice_issue_stage.IssueType,
// the value is dice_issue_stage.Name
func (svc *Issue) getStageMap(stages []dao.IssueStage) map[issueStage]string {
	stageMap := make(map[issueStage]string, len(stages))
	for _, v := range stages {
		if v.Value != "" && v.IssueType != "" {
			stage := issueStage{
				Type:  v.IssueType,
				Value: v.Value,
			}
			stageMap[stage] = v.Name
		}
	}
	return stageMap
}

func (svc *Issue) ExportFalseExcel(r io.Reader, falseIndex []int, falseReason []string, allNumber int) (*apistructs.IssueImportExcelResponse, error) {
	var res apistructs.IssueImportExcelResponse
	sheets, err := excel.Decode(r)
	if err != nil {
		return nil, err
	}
	var exportExcel [][]string
	indexMap := make(map[int]int)
	for i, v := range falseIndex {
		indexMap[v] = i + 1
	}
	rows := sheets[0]
	for i, row := range rows {
		if indexMap[i] > 0 {
			r := append(row, falseReason[indexMap[i]-1])
			exportExcel = append(exportExcel, r)
		}
	}
	tableName := "失败文件"
	uuid, err := svc.ExportExcel2(exportExcel, tableName)
	if err != nil {
		return nil, err
	}
	res.FalseNumber = len(falseIndex) - 1
	res.SuccessNumber = allNumber - res.FalseNumber
	res.UUID = uuid
	return &res, nil
}

func (svc *Issue) ExportExcel2(data [][]string, sheetName string) (string, error) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)
	if err != nil {
		return "", errors.Errorf("failed to add sheetName, sheetName: %s", sheetName)
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
	var buff bytes.Buffer
	if err := file.Write(&buff); err != nil {
		return "", errors.Errorf("failed to write content, sheetName: %s, err: %v", sheetName, err)
	}
	diceFile, err := svc.bdl.UploadFile(apistructs.FileUploadRequest{
		FileNameWithExt: sheetName + ".xlsx",
		ByteSize:        int64(buff.Len()),
		FileReader:      ioutil.NopCloser(&buff),
		From:            "issue",
		IsPublic:        true,
		Encrypt:         false,
		Creator:         "",
		ExpiredAt:       nil,
	})
	if err != nil {
		return "", err
	}
	return diceFile.UUID, nil
}
