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
	"fmt"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issueproperty"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func (svc *Issue) ImportExcel(req apistructs.IssueImportExcelRequest, r *http.Request, properties []apistructs.IssuePropertyIndex, ip *issueproperty.IssueProperty, member []apistructs.Member) (*apistructs.IssueImportExcelResponse, error) {
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
	falseExcel, falseReason = svc.storeExcel2DB(req, issues, instances, excelIndex, ip, falseExcel, falseReason, member)
	return svc.ExportFalseExcel(ff, falseExcel, falseReason, allNumber)
}

func (svc *Issue) storeExcel2DB(request apistructs.IssueImportExcelRequest, issues []apistructs.Issue, instances []apistructs.IssuePropertyRelationCreateRequest, excelIndex []int,
	ip *issueproperty.IssueProperty, falseIssue []int, falseReason []string, member []apistructs.Member) ([]int, []string) {
	memberMap := make(map[string]string)
	for _, m := range member {
		memberMap[m.Nick] = m.UserID
	}
	for index, req := range issues {
		if string(req.Type) != string(request.Type) {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "创建任务失败, err:事件类型不符合")
			continue
		}
		if req.ID > 0 {
			issue, err := svc.db.GetIssue(req.ID)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, fmt.Sprintf("failed to get issue: %s, err: %v", req.Title, err))
				continue
			}
			if issue.ProjectID != request.ProjectID {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, fmt.Sprintf("issue : %s not belong to project: %d", req.Title, request.ProjectID))
				continue
			}
			issue.PlanStartedAt = req.PlanStartedAt
			issue.PlanFinishedAt = req.PlanFinishedAt
			issue.IterationID = req.IterationID
			issue.Type = req.Type
			issue.Title = req.Title
			issue.Content = req.Content
			issue.State = req.State
			issue.Priority = req.Priority
			issue.Complexity = req.Complexity
			issue.Severity = req.Severity
			issue.Creator = memberMap[req.Creator]
			issue.Assignee = memberMap[req.Assignee]
			issue.Source = req.Source
			issue.Stage = req.GetStage()
			issue.Owner = memberMap[req.Owner]
			if err := svc.db.UpdateIssueType(&issue); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, fmt.Sprintf("failed to update issue: %s, err: %v", issue.Title, err))
				continue
			}
			relateds, err := svc.db.GetRelatedIssues(issue.ID)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "failed to get related issues, er: "+err.Error())
				continue
			}

			relatedMap := map[uint64]bool{}
			for _, re := range relateds {
				relatedMap[re] = true
			}

			for _, issueRelated := range req.GetRelatedIssueIDs() {
				if !relatedMap[issueRelated] && issueRelated != issue.ID {
					// check related issue
					relatedIssue, err := svc.db.GetIssue(int64(issueRelated))
					if err != nil {
						continue
					}
					if relatedIssue.ProjectID == request.ProjectID {
						_ = svc.db.CreateIssueRelations(&dao.IssueRelation{
							IssueID:      issueRelated,
							RelatedIssue: issue.ID,
						})
					}
				}
			}
			// label relations
			labels, err := svc.bdl.ListLabelByNameAndProjectID(req.ProjectID, req.Labels)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "failed to query labels, err: "+err.Error())
				continue
			}
			lrs, err := svc.db.GetLabelRelationsByRef(apistructs.LabelTypeIssue, issue.ID)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "failed to query label relations, err: "+err.Error())
				continue
			}
			labelsMap := map[uint64]bool{}
			for _, lr := range lrs {
				labelsMap[lr.LabelID] = true
			}
			for _, label := range labels {
				if !labelsMap[uint64(label.ID)] {
					_ = svc.db.CreateLabelRelation(&dao.LabelRelation{
						LabelID:   uint64(label.ID),
						BaseModel: dbengine.BaseModel{},
						RefType:   apistructs.LabelTypeIssue,
						RefID:     issue.ID,
					})
				}
			}
		} else {
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
			for _, issueRelated := range req.GetRelatedIssueIDs() {
				relatedIssue, err := svc.db.GetIssue(int64(issueRelated))
				if err != nil {
					continue
				}
				if relatedIssue.ProjectID == request.ProjectID {
					_ = svc.db.CreateIssueRelations(&dao.IssueRelation{
						IssueID:      issueRelated,
						RelatedIssue: create.ID,
					})
				}
			}
			// 添加标签关联关系
			labels, err := svc.bdl.ListLabelByNameAndProjectID(req.ProjectID, req.Labels)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "任务已添加，标签添加失败,自定义字段未添加,err:"+err.Error())
				continue
			}
			for _, v := range labels {
				lr := &dao.LabelRelation{
					BaseModel: dbengine.BaseModel{},
					LabelID:   uint64(v.ID),
					RefType:   apistructs.LabelTypeIssue,
					RefID:     create.ID,
				}
				if err := svc.db.CreateLabelRelation(lr); err != nil {
					falseIssue = append(falseIssue, excelIndex[index])
					falseReason = append(falseReason, "任务已添加，标签添加失败, 自定义字段未添加, err:"+err.Error())
					continue
				}
			}
			// 添加自定义字段
			instances[index].IssueID = int64(create.ID)
			if err := ip.CreatePropertyRelation(&instances[index]); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "任务已添加，标签已添加，自定义字段添加失败, err:"+err.Error())
				continue
			}
		}
	}
	return falseIssue, falseReason
}
