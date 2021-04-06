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
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/issueproperty"
	"github.com/erda-project/erda/modules/cmdb/services/label"
)

func (svc *Issue) ImportExcel(req apistructs.IssueImportExcelRequest, r *http.Request, properties []apistructs.IssuePropertyIndex,
	member []model.Member, l *label.Label, ip *issueproperty.IssueProperty) (*apistructs.IssueImportExcelResponse, error) {
	// 获取测试用例数据
	f, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	issues, instances, falseExcel, excelIndex, falseReason, allNumber, err := svc.decodeFromExcelFile(req, f, properties)
	ff, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer ff.Close()
	svc.storeExcel2DB(req, issues, instances, excelIndex, l, ip, falseExcel, falseReason, member)
	return svc.ExportFalseExcel(ff, falseExcel, falseReason, allNumber)
}

func (svc *Issue) storeExcel2DB(request apistructs.IssueImportExcelRequest, issues []apistructs.Issue, instances []apistructs.IssuePropertyRelationCreateRequest, excelIndex []int,
	l *label.Label, ip *issueproperty.IssueProperty, falseIssue []int, falseReason []string, member []model.Member) int {
	memberMap := make(map[string]string)
	for _, m := range member {
		memberMap[m.Nick] = m.UserID
	}
	for index, req := range issues {
		// 创建 issue
		create := dao.Issue{
			PlanStartedAt:  req.PlanStartedAt,
			PlanFinishedAt: req.PlanFinishedAt,
			ProjectID:      uint64(request.ProjectID),
			IterationID:    req.IterationID,
			AppID:          req.AppID,
			Type:           req.Type,
			Title:          req.Title,
			Content:        req.Content,
			State:          req.State,
			Priority:       req.Priority,
			Complexity:     req.Complexity,
			Severity:       apistructs.IssueSeverityNormal,
			Creator:        memberMap[req.Creator],
			Assignee:       memberMap[req.Assignee],
			Source:         req.Source,
			External:       true,
			Stage:          req.GetStage(),
			Owner:          memberMap[req.Owner],
			//ManHour:      req.GetDBManHour(),
		}
		if string(create.Type) != string(request.Type) {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "创建任务失败, err:事件类型不符合")
			continue
		}
		if err := svc.db.CreateIssue(&create); err != nil {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "创建任务失败, err:"+err.Error())
			continue
		}
		// 添加标签关联关系
		labels, err := l.ListByNames(req.ProjectID, req.Labels)
		if err != nil {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "任务已添加，标签添加失败,自定义字段未添加,err:"+err.Error())
			continue
		}
		for _, v := range labels {
			lr := &dao.LabelRelation{
				BaseModel: dao.BaseModel{},
				LabelID:   uint64(v.ID),
				RefType:   apistructs.LabelTypeIssue,
				RefID:     uint64(create.ID),
			}
			if err := l.CreateRelation(lr); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "任务已添加，标签添加失败, 自定义字段未添加, err:"+err.Error())
				continue
			}
		}
		// 添加自定义字段
		instances[index].IssueID = create.ID
		if err := ip.CreatePropertyRelation(&instances[index]); err != nil {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "任务已添加，标签已添加，自定义字段添加失败, err:"+err.Error())
			continue
		}
	}
	return len(falseIssue) - 1
}
