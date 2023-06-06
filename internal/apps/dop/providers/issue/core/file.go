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

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/conf"
	legacydao "github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/core/file/filetypes"
	labeldao "github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (i *IssueService) ExportExcelIssue(ctx context.Context, req *pb.ExportExcelIssueRequest) (*pb.ExportExcelIssueResponse, error) {
	switch req.OrderBy {
	case "":
	case "planStartedAt":
		req.OrderBy = "plan_started_at"
	case "planFinishedAt":
		req.OrderBy = "plan_finished_at"
	case "assignee":
		req.OrderBy = "assignee"
	case "updatedAt", "updated_at":
		req.OrderBy = "updated_at"
	default:
		return nil, apierrors.ErrExportExcelIssue.InvalidParameter("orderBy")
	}

	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrCreateIssuePropertyValue.NotLogin()
	}
	req.IdentityInfo = identityInfo
	if !apis.IsInternalClient(ctx) {
		req.External = true
	}

	req.Locale = apis.GetLang(ctx)
	if req.IsDownloadTemplate {
		// see `transport.WithEncoder` when `pb.RegisterIssueCoreServiceIm`
		return nil, nil
	}

	recordID, err := i.Export(req)
	if err != nil {
		return nil, apierrors.ErrExportExcelIssue.InternalError(err)
	}
	ok, _, err := i.testcase.GetFirstFileReady(apistructs.FileIssueActionTypeExport)
	if err != nil {
		return nil, apierrors.ErrExportExcelIssue.InternalError(err)
	}
	if ok {
		i.ExportChannel <- recordID
	}
	return &pb.ExportExcelIssueResponse{Data: recordID}, nil
}

func (i *IssueService) ImportExcelIssue(ctx context.Context, req *pb.ImportExcelIssueRequest) (*pb.ImportExcelIssueResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrCreateIssuePropertyValue.NotLogin()
	}
	req.IdentityInfo = identityInfo
	if req.FileID == "" {
		return nil, apierrors.ErrImportExcelIssue.InvalidParameter("apiFileUUID")
	}
	req.Locale = apis.GetLang(ctx)

	recordID, err := i.Import(req)
	if err != nil {
		return nil, apierrors.ErrImportExcelIssue.InternalError(err)
	}
	ok, _, err := i.testcase.GetFirstFileReady(apistructs.FileIssueActionTypeImport)
	if err != nil {
		return nil, err
	}
	if ok {
		i.ImportChannel <- recordID
	}
	return &pb.ImportExcelIssueResponse{Data: recordID}, nil
}

func (i *IssueService) Export(req *pb.ExportExcelIssueRequest) (uint64, error) {
	req.PageNo = 1
	req.PageSize = 99999
	fileReq := apistructs.TestFileRecordRequest{
		ProjectID: req.ProjectID,
		OrgID:     uint64(req.OrgID),
		Type:      apistructs.FileIssueActionTypeExport,
		State:     apistructs.FileRecordStatePending,
		IdentityInfo: apistructs.IdentityInfo{
			UserID:         req.IdentityInfo.UserID,
			InternalClient: req.IdentityInfo.InternalClient,
		},
		Extra: apistructs.TestFileExtra{
			IssueFileExtraInfo: &apistructs.IssueFileExtraInfo{
				ExportRequest: req,
			},
		},
	}
	id, err := i.testcase.CreateFileRecord(fileReq)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getIssuePagingRequest(req *pb.ExportExcelIssueRequest) pb.PagingIssueRequest {
	return pb.PagingIssueRequest{
		Title:                 req.Title,
		Type:                  req.Type,
		ProjectID:             req.ProjectID,
		IterationID:           req.IterationID,
		IterationIDs:          req.IterationIDs,
		AppID:                 req.AppID,
		RequirementID:         req.RequirementID,
		State:                 req.State,
		StateBelongs:          req.StateBelongs,
		Creator:               req.Creator,
		Assignee:              req.Assignee,
		Label:                 req.Label,
		StartCreatedAt:        req.StartCreatedAt,
		EndCreatedAt:          req.EndCreatedAt,
		StartFinishedAt:       req.StartFinishedAt,
		EndFinishedAt:         req.EndFinishedAt,
		IsEmptyPlanFinishedAt: req.IsEmptyPlanFinishedAt,
		StartClosedAt:         req.StartClosedAt,
		EndClosedAt:           req.EndClosedAt,
		Priority:              req.Priority,
		Complexity:            req.Complexity,
		Severity:              req.Severity,
		RelatedIssueId:        req.RelatedIssueId,
		Source:                req.Source,
		OrderBy:               req.OrderBy,
		TaskType:              req.TaskType,
		BugStage:              req.BugStage,
		Owner:                 req.Owner,
		WithProcessSummary:    req.WithProcessSummary,
		ExceptIDs:             req.ExceptIDs,
		Asc:                   req.Asc,
		IDs:                   req.IDs,
		IdentityInfo:          req.IdentityInfo,
		External:              req.External,
		CustomPanelID:         req.CustomPanelID,
		OnlyIdResult:          req.OnlyIdResult,
		NotIncluded:           req.NotIncluded,
		PageNo:                req.PageNo,
		PageSize:              req.PageSize,
		OrgID:                 req.OrgID,
		ProjectIDs:            req.ProjectIDs,
	}
}

func (i *IssueService) Import(req *pb.ImportExcelIssueRequest) (uint64, error) {
	fileReq := apistructs.TestFileRecordRequest{
		ProjectID:   req.ProjectID,
		Type:        apistructs.FileIssueActionTypeImport,
		ApiFileUUID: req.FileID,
		State:       apistructs.FileRecordStatePending,
		IdentityInfo: apistructs.IdentityInfo{
			UserID:         req.IdentityInfo.UserID,
			InternalClient: req.IdentityInfo.InternalClient,
		},
		Extra: apistructs.TestFileExtra{
			IssueFileExtraInfo: &apistructs.IssueFileExtraInfo{
				ImportRequest: req,
			},
		},
	}
	id, err := i.testcase.CreateFileRecord(fileReq)
	if err != nil {
		return 0, err
	}
	return id, nil
}

const issueService = "issue-service"

func (i *IssueService) updateIssueFileRecord(id uint64, state apistructs.FileRecordState, descOpt ...string) error {
	var desc string
	if len(descOpt) > 0 {
		desc = descOpt[0]
	}
	if err := i.testcase.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: state, Description: desc}); err != nil {
		logrus.Errorf("%s failed to update file record, err: %v", issueService, err)
		return err
	}
	return nil
}

func (i *IssueService) createDataForFulfillCommon(locale string, orgID int64, projectID uint64, issueTypes []string) (*issueexcel.DataForFulfill, error) {
	// stage map
	stages, err := i.db.GetIssuesStageByOrgID(int64(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to get stages, err: %v", err)
	}
	stageMap := query.GetStageMap(stages)
	// iteration map
	iterationMapByID := make(map[int64]*dao.Iteration)
	iterationMapByName := make(map[string]*dao.Iteration)
	iterations, _, err := i.db.PagingIterations(apistructs.IterationPagingRequest{
		PageNo:    1,
		PageSize:  10000,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get iterations, err: %v", err)
	}
	for _, v := range iterations {
		v := v
		iterationMapByID[int64(v.ID)] = &v
		iterationMapByName[v.Title] = &v
	}
	backlogIteration := &dao.Iteration{Title: "待办事项"}
	iterationMapByID[-1] = backlogIteration
	iterationMapByName["待办事项"] = backlogIteration
	// state map
	stateMapByID := make(map[int64]string)
	stateMapByTypeAndName := make(map[string]map[string]int64) // outerkey: issueType, innerkey: stateName
	states, err := i.db.GetIssuesStatesByProjectID(projectID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get states, err: %v", err)
	}
	for _, v := range states {
		stateMapByID[int64(v.ID)] = v.Name
		if _, ok := stateMapByTypeAndName[v.IssueType]; !ok {
			stateMapByTypeAndName[v.IssueType] = make(map[string]int64)
		}
		stateMapByTypeAndName[v.IssueType][v.Name] = int64(v.ID)
	}
	// username map
	// get all org/project projectMember
	projectMemberQuery := apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(projectID),
		PageNo:    1,
		PageSize:  99999,
	}
	projectMember, err := i.bdl.ListMembers(projectMemberQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list projectMember, err: %v", err)
	}
	orgMemberQuery := apistructs.MemberListRequest{
		ScopeType:         apistructs.OrgScope,
		ScopeID:           orgID,
		PageNo:            1,
		PageSize:          99999,
		DesensitizeEmail:  true,
		DesensitizeMobile: true,
	}
	orgMember, err := i.bdl.ListMembers(orgMemberQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list orgMember, err: %v", err)
	}
	projectMemberMap := map[string]apistructs.Member{}
	for _, member := range projectMember {
		projectMemberMap[member.UserID] = member
	}
	orgMemberMap := map[string]apistructs.Member{}
	for _, member := range orgMember {
		orgMemberMap[member.UserID] = member
	}
	// label map
	labelMapByName := make(map[string]apistructs.ProjectLabel)
	resp, err := i.bdl.ListLabel(apistructs.ProjectLabelListRequest{
		ProjectID: projectID,
		Key:       "",
		Type:      apistructs.LabelTypeIssue,
		PageNo:    1,
		PageSize:  99999,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list labels, err: %v", err)
	}
	for _, v := range resp.List {
		labelMapByName[v.Name] = v
	}
	// custom fields
	properties, err := i.query.BatchGetProperties(int64(orgID), issueTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get properties, err: %v", err)
	}
	customFieldsMap := make(map[pb.PropertyIssueTypeEnum_PropertyIssueType][]*pb.IssuePropertyIndex)
	//customFieldsMapByName := make(map[string]*pb.IssuePropertyIndex)
	propertyEnumMap := make(map[query.PropertyEnumPair]string)
	for _, v := range properties {
		customFieldsMap[v.PropertyIssueType] = append(customFieldsMap[v.PropertyIssueType], v)
		//customFieldsMapByName[v.PropertyName] = v
		if common.IsOptions(v.PropertyType.String()) {
			for _, val := range v.EnumeratedValues {
				propertyEnumMap[query.PropertyEnumPair{PropertyID: v.PropertyID, ValueID: val.Id}] = val.Name
			}
		}
	}

	// result
	dataForFulfill := issueexcel.DataForFulfill{
		Locale:                i.bdl.GetLocale(locale),
		ProjectID:             projectID,
		OrgID:                 orgID,
		StageMap:              stageMap,
		IterationMapByID:      iterationMapByID,
		IterationMapByName:    iterationMapByName,
		StateMap:              stateMapByID,
		StateMapByTypeAndName: stateMapByTypeAndName,
		ProjectMemberMap:      projectMemberMap,
		OrgMemberMap:          orgMemberMap,
		LabelMapByName:        labelMapByName,
		CustomFieldMap:        customFieldsMap,
		//CustomFieldMapByName:  customFieldsMapByName,
		PropertyEnumMap: propertyEnumMap,
	}
	return &dataForFulfill, nil
}

func (i *IssueService) createDataForFulfillForImport(req *pb.ImportExcelIssueRequest) (*issueexcel.DataForFulfill, error) {
	data, err := i.createDataForFulfillCommon(req.Locale, req.OrgID, req.ProjectID, nil) // ignore issueTypes, use all types
	if err != nil {
		return nil, fmt.Errorf("failed to create data for fulfill common, err: %v", err)
	}
	// import only
	data.ImportOnly.DB = i.db
	data.ImportOnly.LabelDB = &labeldao.DBClient{DB: i.db.DB}
	data.ImportOnly.Bdl = i.bdl
	data.ImportOnly.Identity = i.identity
	data.ImportOnly.Property = i
	return data, nil
}

func (i *IssueService) createDataForFulfillForExport(req *pb.ExportExcelIssueRequest) (*issueexcel.DataForFulfill, error) {
	data, err := i.createDataForFulfillCommon(req.Locale, req.OrgID, req.ProjectID, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create data for fulfill common, err: %v", err)
	}
	// export only
	// issues
	issues, _, err := i.query.Paging(getIssuePagingRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to page issues, err: %v", err)
	}
	data.ExportOnly.Issues = issues
	data.ExportOnly.IsDownloadTemplate = req.IsDownloadTemplate
	data.ExportOnly.FileNameWithExt = "issue-export.xlsx"
	// propertyRelation map
	var issueIDs []int64
	for _, v := range issues {
		issueIDs = append(issueIDs, v.Id)
	}
	propertyRelationMap := make(map[int64][]dao.IssuePropertyRelation)
	propertyRelations, err := i.db.PagingPropertyRelationByIDs(issueIDs)
	if err != nil {
		return nil, err
	}
	for _, relation := range propertyRelations {
		propertyRelationMap[relation.IssueID] = append(propertyRelationMap[relation.IssueID], relation)
	}
	data.ExportOnly.IssuePropertyRelationMap = propertyRelationMap
	// inclusion map, connection map
	uint64IssueIDs := make([]uint64, 0)
	for _, v := range issueIDs {
		uint64IssueIDs = append(uint64IssueIDs, uint64(v))
	}
	relations, err := i.db.GetIssueRelationsByIDs(uint64IssueIDs, []string{apistructs.IssueRelationInclusion, apistructs.IssueRelationConnection})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue relations, err: %v", err)
	}
	inclusionMap := make(map[int64][]int64)
	connectionMap := make(map[int64][]int64)
	for _, relation := range relations {
		switch relation.Type {
		case apistructs.IssueRelationInclusion:
			inclusionMap[int64(relation.IssueID)] = append(inclusionMap[int64(relation.IssueID)], int64(relation.RelatedIssue))
		case apistructs.IssueRelationConnection:
			connectionMap[int64(relation.IssueID)] = append(connectionMap[int64(relation.IssueID)], int64(relation.RelatedIssue))
		}
	}
	data.ExportOnly.InclusionMap = inclusionMap
	data.ExportOnly.ConnectionMap = connectionMap
	return data, nil
}

func (i *IssueService) ExportExcelAsync(record *legacydao.TestFileRecord) {
	extra := record.Extra.IssueFileExtraInfo
	if extra == nil || extra.ExportRequest == nil {
		return
	}
	req := extra.ExportRequest
	id := record.ID
	if err := i.updateIssueFileRecord(id, apistructs.FileRecordStateProcessing); err != nil {
		return
	}

	// use new excel export
	var buffer bytes.Buffer
	dataForFulfill, err := i.createDataForFulfillForExport(req)
	if err != nil {
		logrus.Errorf("%s failed to create data for fulfill, err: %v", issueService, err)
		i.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	if err := issueexcel.ExportFile(&buffer, *dataForFulfill); err != nil {
		logrus.Errorf("%s failed to export excel, err: %v", issueService, err)
		i.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}

	expiredAt := time.Now().Add(time.Duration(conf.ExportIssueFileStoreDay()) * 24 * time.Hour)
	uploadReq := filetypes.FileUploadRequest{
		FileNameWithExt: dataForFulfill.ExportOnly.FileNameWithExt,
		FileReader:      io.NopCloser(&buffer),
		From:            issueService,
		IsPublic:        true,
		ExpiredAt:       &expiredAt,
	}
	fileUUID, err := i.bdl.UploadFile(uploadReq)
	if err != nil {
		logrus.Errorf("%s failed to upload file, err: %v", issueService, err)
		i.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	i.testcase.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateSuccess, ApiFileUUID: fileUUID.UUID})
}

func (i *IssueService) ImportExcel(record *legacydao.TestFileRecord) (err error) {
	defer func() {
		var desc string
		if r := recover(); r != nil {
			desc = fmt.Sprintf("%v", r)
			err = fmt.Errorf("%v", r)
			fmt.Println(string(debug.Stack()))
		}
		if err != nil {
			logrus.Errorf("%s failed to import excel, recordID: %d, err: %v", issueService, record.ID, err)
			i.updateIssueFileRecord(record.ID, apistructs.FileRecordStateFail, desc)
		}
	}()
	extra := record.Extra.IssueFileExtraInfo
	if extra == nil || extra.ImportRequest == nil {
		return
	}

	req := extra.ImportRequest
	id := record.ID
	if err = i.updateIssueFileRecord(id, apistructs.FileRecordStateProcessing); err != nil {
		return
	}

	f, err := i.bdl.DownloadDiceFile(record.ApiFileUUID)
	if err != nil {
		logrus.Errorf("%s failed to download excel file, err: %v", issueService, err)
		i.updateIssueFileRecord(id, apistructs.FileRecordStateFail)
		return
	}
	defer f.Close()

	data, err := i.createDataForFulfillForImport(req)
	if err != nil {
		panic(fmt.Errorf("failed to create data for fulfill, err: %v", err))
	}
	if err = issueexcel.ImportFile(f, *data); err != nil {
		panic(fmt.Errorf("failed to import excel, err: %v", err))
	}
	i.updateIssueFileRecord(id, apistructs.FileRecordStateSuccess)
	return
}

func getStageValue(issue pb.Issue, stages []dao.IssueStage) string {
	val := common.GetStage(&issue)
	for _, stage := range stages {
		if stage.Name == val && issue.Type.String() == stage.IssueType {
			return stage.Value
		}
	}
	return val
}

func (i *IssueService) storeExcel2DB(request *pb.ImportExcelIssueRequest, issues []pb.Issue, instances []*pb.CreateIssuePropertyInstanceRequest, excelIndex []int,
	falseIssue []int, falseReason []string, member []apistructs.Member) ([]int, []string) {
	memberMap := make(map[string]string)
	for _, m := range member {
		memberMap[m.Nick] = m.UserID
	}
	orgID, err := strconv.ParseInt(request.IdentityInfo.OrgID, 10, 64)
	if err != nil {
		falseIssue = append(falseIssue, excelIndex[0])
		falseReason = append(falseReason, "failed to parse orgID")
		return falseIssue, falseReason
	}
	stages, err := i.db.GetIssuesStageByOrgID(orgID)
	if err != nil {
		falseIssue = append(falseIssue, excelIndex[0])
		falseReason = append(falseReason, "get stages failed")
		return falseIssue, falseReason
	}
	for index, req := range issues {
		if req.Type.String() != request.Type {
			falseIssue = append(falseIssue, excelIndex[index])
			falseReason = append(falseReason, "创建任务失败, err:事件类型不符合")
			continue
		}
		if req.Id > 0 {
			issue, err := i.db.GetIssue(req.Id)
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
			issue.PlanStartedAt = common.ToIssueTime(req.PlanStartedAt)
			issue.PlanFinishedAt = common.ToIssueTime(req.PlanFinishedAt)
			issue.IterationID = req.IterationID
			issue.Type = req.Type.String()
			issue.Title = req.Title
			issue.Content = req.Content
			issue.State = req.State
			issue.Priority = req.Priority.String()
			issue.Complexity = req.Complexity.String()
			issue.Severity = req.Severity.String()
			issue.Creator = memberMap[req.Creator]
			issue.Assignee = memberMap[req.Assignee]
			issue.Source = req.Source
			issue.Stage = getStageValue(req, stages)
			issue.Owner = memberMap[req.Owner]
			if req.IssueManHour != nil && req.IssueManHour.EstimateTime > 0 {
				var oldManHour apistructs.IssueManHour
				json.Unmarshal([]byte(issue.ManHour), &oldManHour)
				oldManHour.EstimateTime = req.IssueManHour.EstimateTime
				if oldManHour.RemainingTime == 0 {
					oldManHour.RemainingTime = oldManHour.EstimateTime
				}
				newManHour, _ := json.Marshal(oldManHour)
				issue.ManHour = string(newManHour)
			}
			if err := i.db.UpdateIssueType(&issue); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, fmt.Sprintf("failed to update issue: %s, err: %v", issue.Title, err))
				continue
			}
			relateds, err := i.db.GetRelatedIssues(issue.ID, []string{apistructs.IssueRelationConnection})
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
					relatedIssue, err := i.db.GetIssue(int64(issueRelated))
					if err != nil {
						continue
					}
					if relatedIssue.ProjectID == request.ProjectID {
						_ = i.db.CreateIssueRelations(&dao.IssueRelation{
							IssueID:      issueRelated,
							RelatedIssue: issue.ID,
						})
					}
				}
			}
			// label relations
			labels, err := i.bdl.ListLabelByNameAndProjectID(req.ProjectID, req.Labels)
			if err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "failed to query labels, err: "+err.Error())
				continue
			}
			lrs, err := i.db.GetLabelRelationsByRef(string(apistructs.LabelTypeIssue), strconv.FormatUint(issue.ID, 10))
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
					_ = i.db.CreateLabelRelation(&dao.LabelRelation{
						LabelID:   uint64(label.ID),
						BaseModel: dbengine.BaseModel{},
						RefType:   apistructs.LabelTypeIssue,
						RefID:     strconv.FormatUint(issue.ID, 10),
					})
				}
			}
		} else {
			create := importIssueBuilder(req, request, memberMap, stages)
			if create.Type != request.Type {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "创建任务失败, err:事件类型不符合")
				continue
			}
			if err := i.db.CreateIssue(&create); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "创建任务失败, err:"+err.Error())
				continue
			}
			for _, issueRelated := range req.GetRelatedIssueIDs() {
				relatedIssue, err := i.db.GetIssue(int64(issueRelated))
				if err != nil {
					continue
				}
				if relatedIssue.ProjectID == request.ProjectID {
					_ = i.db.CreateIssueRelations(&dao.IssueRelation{
						IssueID:      issueRelated,
						RelatedIssue: create.ID,
					})
				}
			}
			// 添加标签关联关系
			labels, err := i.bdl.ListLabelByNameAndProjectID(req.ProjectID, req.Labels)
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
					RefID:     strconv.FormatUint(create.ID, 10),
				}
				if err := i.db.CreateLabelRelation(lr); err != nil {
					falseIssue = append(falseIssue, excelIndex[index])
					falseReason = append(falseReason, "任务已添加，标签添加失败, 自定义字段未添加, err:"+err.Error())
					continue
				}
			}
			// 添加自定义字段
			instances[index].IssueID = int64(create.ID)
			if err := i.query.CreatePropertyRelation(instances[index]); err != nil {
				falseIssue = append(falseIssue, excelIndex[index])
				falseReason = append(falseReason, "任务已添加，标签已添加，自定义字段添加失败, err:"+err.Error())
				continue
			}
		}
	}
	return falseIssue, falseReason
}

func importIssueBuilder(issue pb.Issue, request *pb.ImportExcelIssueRequest, memberMap map[string]string, stages []dao.IssueStage) dao.Issue {
	create := dao.Issue{
		PlanStartedAt:  common.ToIssueTime(issue.PlanStartedAt),
		PlanFinishedAt: common.ToIssueTime(issue.PlanFinishedAt),
		ProjectID:      uint64(request.ProjectID),
		IterationID:    issue.IterationID,
		AppID:          &issue.AppID,
		Type:           issue.Type.String(),
		Title:          issue.Title,
		Content:        issue.Content,
		State:          issue.State,
		Priority:       issue.Priority.String(),
		Complexity:     issue.Complexity.String(),
		Severity:       pb.IssueSeverityEnum_NORMAL.String(),
		Creator:        request.IdentityInfo.UserID,
		Assignee:       memberMap[issue.Assignee],
		Source:         issue.Source,
		External:       true,
		Stage:          getStageValue(issue, stages),
		Owner:          memberMap[issue.Owner],
	}
	if issue.IssueManHour != nil && issue.IssueManHour.EstimateTime > 0 {
		newManHour, _ := json.Marshal(issue.IssueManHour)
		create.ManHour = string(newManHour)
	}
	return create
}

func (i *IssueService) ExportFailedExcel(r io.Reader, falseIndex []int, falseReason []string, allNumber int) (*apistructs.IssueImportExcelResponse, error) {
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
	uuid, err := i.ExportExcel2(exportExcel, tableName)
	if err != nil {
		return nil, err
	}
	res.FalseNumber = len(falseIndex) - 1
	res.SuccessNumber = allNumber - res.FalseNumber
	res.UUID = uuid
	return &res, nil
}

func (i *IssueService) ExportExcel2(data [][]string, sheetName string) (string, error) {
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
	diceFile, err := i.bdl.UploadFile(filetypes.FileUploadRequest{
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

func (i *IssueService) decodeFromExcelFile(req *pb.ImportExcelIssueRequest, r io.Reader, properties []*pb.IssuePropertyIndex) ([]pb.Issue,
	[]*pb.CreateIssuePropertyInstanceRequest, []int, []int, []string, int, error) {
	var (
		falseExcel, excelIndex []int
		falseReason            []string
		allIssue               []pb.Issue
		allInstance            []*pb.CreateIssuePropertyInstanceRequest
	)
	sheets, err := excel.Decode(r)
	// filter empty row
	sheetLst := make([][][]string, 0)
	for _, rows := range sheets {
		rowLst := make([][]string, 0)
		for _, row := range rows {
			if strings.Join(row, "") == "" {
				continue
			}
			rowLst = append(rowLst, row)
		}
		sheetLst = append(sheetLst, rowLst)
	}
	sheets = sheetLst
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to decode excel, err: %v", err)
	}
	if len(sheets) == 0 {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("not found sheet")
	}
	rows := sheets[0]
	// 校验：至少有1行 title
	if len(rows) < 1 {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("invalid title format")
	}
	falseExcel = append(falseExcel, 0)
	falseReason = append(falseReason, "错误原因")
	// 获取状态
	states, err := i.db.GetIssuesStatesByProjectID(req.ProjectID, req.Type)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to get state, err: %v", err)
	}
	stateMap := make(map[string]int64) // key: state  value: id
	for _, s := range states {
		stateMap[s.Name] = int64(s.ID)
	}
	// 获取迭代信息
	iterations, err := i.db.FindIterations(req.ProjectID)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	iterationMap := make(map[string]int64) // key: iterationName value: iterationID
	for _, it := range iterations {
		iterationMap[it.Title] = int64(it.ID)
	}
	iterationMap["待办事项"] = -1
	// 获取自定义字段
	type propertyValue struct {
		PropertyID int64
		Value      string
	}
	propertyNameMap := make(map[string]*pb.IssuePropertyIndex) // key: propertyName value: property
	propertyMap := make(map[propertyValue]int64)               // key: propertyID+value  value: valueID
	for _, pro := range properties {
		propertyNameMap[pro.PropertyName] = pro
		if common.IsOptions(pro.PropertyType.String()) == true {
			for _, val := range pro.EnumeratedValues {
				propertyMap[propertyValue{pro.PropertyID, val.Name}] = val.Id
			}
		}
	}
	// 第一行是列名,之后每行都是一个事件
	for i, row := range rows[1:] {
		issue := pb.Issue{
			Title:     row[1],
			Content:   row[2],
			ProjectID: req.ProjectID,
		}
		if row[0] != "" {
			issueID, err := strconv.Atoi(row[0])
			if err != nil {
				falseExcel = append(falseExcel, i+1)
				falseReason = append(falseReason, fmt.Sprintf("failed to convert id: %s, err: %v", row[0], err))
				continue
			}
			issue.Id = int64(issueID)
		}
		if stateMap[row[3]] != 0 {
			issue.State = stateMap[row[3]]
		} else {
			falseExcel = append(falseExcel, i+1)
			falseReason = append(falseReason, "无法找到该状态")
			continue
		}
		issue.Creator = row[4]
		issue.Assignee = row[5]
		issue.Owner = row[6]
		issue.TaskType = row[7]
		issue.BugStage = row[7]
		issue.Priority = GetProperty(row[8])
		if val, ok := iterationMap[row[9]]; !ok {
			falseExcel = append(falseExcel, i+1)
			falseReason = append(falseReason, "无法找到该迭代")
			continue
		} else {
			issue.IterationID = val
		}
		issue.Complexity = GetComplexity(row[10])
		issue.Severity = GetSeverity(row[11])
		issue.Labels = strutil.Split(row[12], ",", true)
		issue.Type = GetType(row[13])
		if row[14] != "" {
			finishedTime, err := time.ParseInLocation("2006-01-02 15:04:05", row[14], time.Local)
			if err != nil {
				falseExcel = append(falseExcel, i+1)
				falseReason = append(falseReason, "无法解析任务截止时间, 正确格式: 2006-01-02 15:04:05")
				continue
			}
			issue.PlanFinishedAt = timestamppb.New(finishedTime)
		}

		// firstLine[15]is created time, jump over
		// row[16] RelatedIssueIDs
		if err := SetRelatedIssueIDs(&issue, row[16]); err != nil {
			falseExcel = append(falseExcel, i+1)
			falseReason = append(falseReason, fmt.Sprintf("failed to convert related issue ids: %s, err: %v", row[17], err))
			continue
		}
		// row[17] EstimateTime
		if len(row) >= 18 && row[17] != "" {
			manHour, err := issueexcel.NewManhour(row[17])
			if err != nil {
				falseExcel = append(falseExcel, i+1)
				falseReason = append(falseReason, fmt.Sprintf("failed to convert estimate time: %s, err: %v", row[17], err))
				continue
			}
			issue.IssueManHour = &manHour
		}

		// row[18] finish time, jump over

		// row[19] plan start time
		if len(row) >= 20 && row[19] != "" {
			planStartAt, err := time.ParseInLocation("2006-01-02 15:04:05", row[19], time.Local)
			if err != nil {
				falseExcel = append(falseExcel, i+1)
				falseReason = append(falseReason, fmt.Sprintf("failed to convert plan start time: %s, err: %v", row[19], err))
				continue
			}
			issue.PlanStartedAt = timestamppb.New(planStartAt)
		}

		// row[20] reopen count, jump over

		// 获取自定义字段
		relation := &pb.CreateIssuePropertyInstanceRequest{
			OrgID:     req.OrgID,
			ProjectID: int64(req.ProjectID),
		}
		if len(row) >= 22 {
			for indexx, line := range row[20:] {
				index := indexx + 20
				// 获取字段名对应的字段
				propertyIndex, ok := propertyNameMap[rows[0][index]]
				if !ok {
					falseExcel = append(falseExcel, i+1)
					falseReason = append(falseReason, fmt.Sprintf("custom property %s is not defined in org", row[index]))
					continue
				}
				instance := &pb.IssuePropertyInstance{
					PropertyID:        propertyIndex.PropertyID,
					ScopeID:           propertyIndex.ScopeID,
					ScopeType:         propertyIndex.ScopeType,
					OrgID:             propertyIndex.OrgID,
					PropertyName:      propertyIndex.PropertyName,
					DisplayName:       propertyIndex.DisplayName,
					PropertyType:      propertyIndex.PropertyType,
					Required:          propertyIndex.Required,
					PropertyIssueType: propertyIndex.PropertyIssueType,
					Relation:          propertyIndex.Relation,
					Index:             propertyIndex.Index,
					EnumeratedValues:  propertyIndex.EnumeratedValues,
					Values:            propertyIndex.Values,
					RelatedIssue:      propertyIndex.RelatedIssue,
				}
				if !common.IsOptions(instance.PropertyType.String()) {
					instance.ArbitraryValue = structpb.NewStringValue(line)
				} else {
					values := strutil.Split(line, ",", true)
					for _, val := range values {
						instance.Values = append(instance.Values, propertyMap[propertyValue{instance.PropertyID, val}])
					}
				}
				relation.Property = append(relation.Property, instance)
			}
		}
		allIssue = append(allIssue, issue)
		allInstance = append(allInstance, relation)
		excelIndex = append(excelIndex, i+1)
	}

	return allIssue, allInstance, falseExcel, excelIndex, falseReason, len(rows) - 1, nil
}

// SetRelatedIssueIDs set RelatedIssueIDs from excel
func SetRelatedIssueIDs(s *pb.Issue, ids string) error {
	if ids == "" {
		return nil
	}
	idStrs := strings.Split(ids, ",")
	dp := map[uint64]bool{}
	relatedIssueIDs := make([]uint64, 0)
	for _, id := range idStrs {
		issueID, err := strconv.Atoi(id)
		if err != nil {
			return err
		}
		if dp[uint64(issueID)] {
			continue
		}
		dp[uint64(issueID)] = true
		relatedIssueIDs = append(relatedIssueIDs, uint64(issueID))
	}
	s.RelatedIssueIDs = relatedIssueIDs
	return nil
}

func GetProperty(zh string) pb.IssuePriorityEnum_Priority {
	switch zh {
	case "紧急":
		return pb.IssuePriorityEnum_URGENT
	case "高":
		return pb.IssuePriorityEnum_HIGH
	case "中":
		return pb.IssuePriorityEnum_NORMAL
	case "低":
		return pb.IssuePriorityEnum_LOW
	default:
		return pb.IssuePriorityEnum_NORMAL
	}
}

func GetComplexity(zh string) pb.IssueComplexityEnum_Complextity {
	switch zh {
	case "复杂":
		return pb.IssueComplexityEnum_HARD
	case "中":
		return pb.IssueComplexityEnum_NORMAL
	case "容易":
		return pb.IssueComplexityEnum_EASY
	default:
		return pb.IssueComplexityEnum_NORMAL
	}
}

func GetSeverity(zh string) pb.IssueSeverityEnum_Severity {
	switch zh {
	case "致命":
		return pb.IssueSeverityEnum_FATAL
	case "严重":
		return pb.IssueSeverityEnum_SERIOUS
	case "一般":
		return pb.IssueSeverityEnum_NORMAL
	case "轻微":
		return pb.IssueSeverityEnum_SLIGHT
	case "建议":
		return pb.IssueSeverityEnum_SUGGEST
	default:
		return pb.IssueSeverityEnum_NORMAL
	}
}

func GetType(s string) pb.IssueTypeEnum_Type {
	switch s {
	case "需求":
		return pb.IssueTypeEnum_REQUIREMENT
	case "任务":
		return pb.IssueTypeEnum_TASK
	case "缺陷":
		return pb.IssueTypeEnum_BUG
	case "工单":
		return pb.IssueTypeEnum_TICKET
	case "史诗":
		return pb.IssueTypeEnum_EPIC
	default:
		return pb.IssueTypeEnum_REQUIREMENT
	}
}
