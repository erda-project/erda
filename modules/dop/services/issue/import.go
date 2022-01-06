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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issueproperty"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

const (
	fileUploadFrom     = "autotest-space"
	fileValidityPeriod = 7 * time.Hour * 24
	issueService       = "issue-service"
)

func (svc *Issue) Import(req apistructs.IssueImportExcelRequest, r *http.Request) (uint64, error) {
	f, fileHeader, err := r.FormFile("file")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	expiredAt := time.Now().Add(fileValidityPeriod)
	uploadReq := apistructs.FileUploadRequest{
		FileNameWithExt: fileHeader.Filename,
		FileReader:      f,
		From:            fileUploadFrom,
		IsPublic:        true,
		ExpiredAt:       &expiredAt,
	}
	file, err := svc.bdl.UploadFile(uploadReq)
	if err != nil {
		return 0, err
	}

	fileReq := apistructs.TestFileRecordRequest{
		FileName:     fileHeader.Filename,
		ProjectID:    req.ProjectID,
		Type:         apistructs.FileIssueActionTypeImport,
		ApiFileUUID:  file.UUID,
		State:        apistructs.FileRecordStatePending,
		IdentityInfo: req.IdentityInfo,
		Extra: apistructs.TestFileExtra{
			IssueFileExtraInfo: &apistructs.IssueFileExtraInfo{
				ImportRequest: &req,
			},
		},
	}
	id, err := svc.CreateFileRecord(fileReq)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (svc *Issue) ImportExcel(record *dao.TestFileRecord) {
	extra := record.Extra.IssueFileExtraInfo
	if extra == nil || extra.ImportRequest == nil {
		return
	}

	req := extra.ImportRequest
	id := record.ID
	if err := svc.updateIssueFileRecord(id, apistructs.FileRecordStateProcessing); err != nil {
		return
	}

	f, err := svc.bdl.DownloadDiceFile(record.ApiFileUUID)
	if err != nil {
		logrus.Errorf("%s failed to download excel file, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	defer f.Close()

	properties, err := svc.Ip.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: req.OrgID, PropertyIssueType: req.Type})
	if err != nil {
		logrus.Errorf("%s failed to get issue properties, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	memberQuery := apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(req.ProjectID),
		PageNo:    1,
		PageSize:  99999,
	}
	members, err := svc.bdl.ListMembers(memberQuery)
	if err != nil {
		logrus.Errorf("%s failed to get members, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}

	issues, instances, falseExcel, excelIndex, falseReason, allNumber, err := svc.decodeFromExcelFile(*req, f, properties)
	if err != nil {
		logrus.Errorf("%s failed to decode excel file, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	falseExcel, falseReason = svc.storeExcel2DB(*req, issues, instances, excelIndex, svc.Ip, falseExcel, falseReason, members)
	if len(falseExcel) <= 1 {
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateSuccess)
		return
	}
	ff, err := svc.bdl.DownloadDiceFile(record.ApiFileUUID)
	if err != nil {
		logrus.Errorf("%s failed to download excel file, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	defer ff.Close()
	res, err := svc.ExportFalseExcel(ff, falseExcel, falseReason, allNumber)
	if err != nil {
		logrus.Errorf("%s failed to export false excel, err: %v", issueService, err)
		svc.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	desc := fmt.Sprintf("事项总数: %d, 成功: %d, 失败: %d", res.SuccessNumber+res.FalseNumber, res.SuccessNumber, res.FalseNumber)
	svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, Description: desc, ApiFileUUID: res.UUID, State: apistructs.FileRecordStateFail})
}

func (svc *Issue) updateIssueFileRecord(id uint64, state apistructs.FileRecordState) error {
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: state}); err != nil {
		logrus.Errorf("%s failed to update file record, err: %v", issueService, err)
		return err
	}
	return nil
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
			if req.ManHour.EstimateTime > 0 {
				var oldManHour apistructs.IssueManHour
				json.Unmarshal([]byte(issue.ManHour), &oldManHour)
				oldManHour.EstimateTime = req.ManHour.EstimateTime
				if oldManHour.RemainingTime == 0 {
					oldManHour.RemainingTime = oldManHour.EstimateTime
				}
				newManHour, _ := json.Marshal(oldManHour)
				issue.ManHour = string(newManHour)
			}
			if err := svc.db.UpdateIssueType(&issue); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, fmt.Sprintf("failed to update issue: %s, err: %v", issue.Title, err))
				continue
			}
			relateds, err := svc.db.GetRelatedIssues(issue.ID, []string{apistructs.IssueRelationConnection})
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
			if req.ManHour.EstimateTime > 0 {
				newManHour, _ := json.Marshal(req.ManHour)
				create.ManHour = string(newManHour)
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
