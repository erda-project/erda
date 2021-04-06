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

package issue

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tealeg/xlsx/v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (svc *Issue) ExportExcel(issues []apistructs.Issue, properties []apistructs.IssuePropertyIndex, projectID uint64, isDownload bool) (io.Reader, string, error) {
	table, err := svc.convertIssueToExcelList(issues, properties, projectID, isDownload)
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
		usernames[strconv.FormatUint(u.ID, 10)] = u.Nick
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
		tablename = issues[0].Type.GetZhName()
	}

	if issues[0].IterationID == -1 {
		tablename = "待办事项"
	}

	buf := bytes.NewBuffer([]byte{})
	if err := excel.ExportExcel(buf, table, tablename); err != nil {
		return nil, "", err
	}
	return buf, tablename, nil
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
	diceFile, err := svc.fileSvc.UploadFile(apistructs.FileUploadRequest{
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
