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

// Package issue 封装 事件 相关操作
package issue

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/issuestream"
	"github.com/erda-project/erda/modules/dop/services/monitor"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

// Issue 事件操作封装
type Issue struct {
	db     *dao.DBClient
	bdl    *bundle.Bundle
	stream *issuestream.IssueStream
	uc     *ucauth.UCClient
}

// Option 定义 Issue 配置选项
type Option func(*Issue)

// New 新建 Issue 实例
func New(options ...Option) *Issue {
	itr := &Issue{}
	for _, op := range options {
		op(itr)
	}
	return itr
}

// WithDBClient 配置 Issue 数据库选项
func WithDBClient(db *dao.DBClient) Option {
	return func(issue *Issue) {
		issue.db = db
	}
}

// WithBundle 配置 bdl
func WithBundle(bdl *bundle.Bundle) Option {
	return func(issue *Issue) {
		issue.bdl = bdl
	}
}

// WithIssueStream 配置 事件流 选项
func WithIssueStream(stream *issuestream.IssueStream) Option {
	return func(issue *Issue) {
		issue.stream = stream
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(issue *Issue) {
		issue.uc = uc
	}
}

// Create 创建事件
func (svc *Issue) Create(req *apistructs.IssueCreateRequest) (*dao.Issue, error) {
	// 请求校验
	if req.ProjectID == 0 {
		return nil, apierrors.ErrCreateIssue.MissingParameter("projectID")
	}
	if req.Title == "" {
		return nil, apierrors.ErrCreateIssue.MissingParameter("title")
	}
	if req.Type == "" {
		return nil, apierrors.ErrCreateIssue.MissingParameter("type")
	}
	// 不归属任何迭代时，IterationID=-1
	if req.IterationID == 0 {
		return nil, apierrors.ErrCreateIssue.MissingParameter("iterationID")
	}
	// 工单允许处理人为空
	if req.Assignee == "" && req.Type != apistructs.IssueTypeTicket {
		return nil, apierrors.ErrCreateIssue.MissingParameter("assignee")
	}
	// 显式指定了创建人，则覆盖
	if req.Creator != "" {
		req.UserID = req.Creator
	}
	// 初始状态为排序级最高的状态
	initState, err := svc.db.GetIssuesStatesByProjectID(req.ProjectID, req.Type)
	if err != nil {
		return nil, err
	}
	if len(initState) == 0 {
		return nil, apierrors.ErrCreateIssue.InvalidParameter("缺少默认事件状态")
	}
	if req.IterationID == -1 {
		if req.Complexity == "" {
			req.Complexity = apistructs.IssueComplexityNormal
		}
		if req.Severity == "" {
			req.Severity = apistructs.IssueSeverityNormal
		}
		if req.Priority == "" {
			req.Priority = apistructs.IssuePriorityNormal
		}
	}
	// 创建 issue
	create := dao.Issue{
		PlanStartedAt:  req.PlanStartedAt,
		PlanFinishedAt: req.PlanFinishedAt,
		ProjectID:      req.ProjectID,
		IterationID:    req.IterationID,
		AppID:          req.AppID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		State:          int64(initState[0].ID),
		Priority:       req.Priority,
		Complexity:     req.Complexity,
		Severity:       req.Severity,
		Creator:        req.UserID,
		Assignee:       req.Assignee,
		Source:         req.Source,
		ManHour:        req.GetDBManHour(),
		External:       req.External,
		Stage:          req.GetStage(),
		Owner:          req.Owner,
	}
	if err := svc.db.CreateIssue(&create); err != nil {
		return nil, apierrors.ErrCreateIssue.InternalError(err)
	}

	// create subscribers
	issueID := int64(create.ID)
	req.Subscribers = append(req.Subscribers, create.Creator)
	req.Subscribers = strutil.DedupSlice(req.Subscribers)
	var subscriberModels []dao.IssueSubscriber
	for _, v := range req.Subscribers {
		subscriberModels = append(subscriberModels, dao.IssueSubscriber{IssueID: issueID, UserID: v})
	}
	if err := svc.db.BatchCreateIssueSubscribers(subscriberModels); err != nil {
		return nil, apierrors.ErrCreateIssue.InternalError(err)
	}

	// 生成活动记录
	users, err := svc.uc.FindUsers([]string{req.UserID})
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, errors.Errorf("not found user info")
	}
	streamReq := apistructs.IssueStreamCreateRequest{
		IssueID:      int64(create.ID),
		Operator:     req.UserID,
		StreamType:   apistructs.ISTCreate,
		StreamParams: apistructs.ISTParam{UserName: users[0].Nick},
	}
	// create stream and send issue create event
	if _, err := svc.stream.Create(&streamReq); err != nil {
		return nil, err
	}

	go monitor.MetricsIssueById(int(create.ID), svc.db, svc.uc, svc.bdl)

	return &create, nil
}

// Paging 分页查询事件
func (svc *Issue) Paging(req apistructs.IssuePagingRequest) ([]apistructs.Issue, uint64, error) {
	// 请求校验
	if req.ProjectID == 0 {
		return nil, 0, apierrors.ErrPagingIssues.MissingParameter("projectID")
	}
	// 待办事项允许迭代id为-1即只能看未纳入迭代的事项，默认按照优先级排序
	if (req.IterationID == -1 || (len(req.IterationIDs) == 1 && req.IterationIDs[0] == -1)) && req.OrderBy == "" {
		// req.Type = apistructs.IssueTypeRequirement
		req.OrderBy = "FIELD(priority, 'LOW', 'NORMAL', 'HIGH', 'URGENT')"
	}
	if req.IterationID != 0 {
		req.IterationIDs = append(req.IterationIDs, req.IterationID)
	}

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	var (
		labelRelationIDs, issueRelationIDs []int64
		isLabel, isIssue                   bool
	)
	if len(req.Label) > 0 {
		isLabel = true
		// 获取标签关联关系
		lrs, err := svc.db.GetLabelRelationsByLabels(apistructs.LabelTypeIssue, req.Label)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
		for _, v := range lrs {
			labelRelationIDs = append(labelRelationIDs, int64(v.RefID))
		}
	}
	if len(req.RelatedIssueIDs) > 0 {
		isIssue = true
		// 获取事件关联关系
		irs, err := svc.db.GetIssueRelationsByIDs(req.RelatedIssueIDs)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
		for _, v := range irs {
			issueRelationIDs = append(issueRelationIDs, int64(v.RelatedIssue))
		}
	}
	if isLabel || isIssue {
		req.IDs = strutil.DedupInt64Slice(append(getRelatedIDs(labelRelationIDs, issueRelationIDs, isLabel, isIssue), req.IDs...))
	}

	// 该项目下全部的state信息，之后不再查询state  key: stateID value:state
	stateMap := make(map[int64]dao.IssueState)
	// 根据主状态过滤
	if len(req.StateBelongs) > 0 {
		err := svc.FilterByStateBelong(stateMap, &req)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
	}
	// 分页
	issueModels, total, err := svc.db.PagingIssues(req, isLabel || isIssue)
	if err != nil {
		return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
	}

	issues, err := svc.BatchConvert(issueModels, req.Type, req.IdentityInfo)
	if err != nil {
		return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
	}

	// issue 填充需求标题
	requirementIDs := make([]int64, 0, len(issues))
	for _, v := range issues {
		if v.RequirementID != nil && *v.RequirementID > 0 {
			requirementIDs = append(requirementIDs, *v.RequirementID)
		}
	}
	if len(requirementIDs) > 0 {
		requirements, err := svc.db.ListIssueByIDs(requirementIDs)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
		requirementTitleMap := make(map[int64]string, len(requirements))
		for _, r := range requirements {
			requirementTitleMap[int64(r.ID)] = r.Title
		}
		for i, v := range issues {
			if v.RequirementID != nil && *v.RequirementID > 0 {
				issues[i].RequirementTitle = requirementTitleMap[(*v.RequirementID)]
			}
		}
	}

	// 需求进度统计
	if req.WithProcessSummary {
		stateBelongMap := make(map[int64]apistructs.IssueStateBelong)
		stateBelong, err := svc.db.GetIssuesStatesByProjectID(req.ProjectID, "")
		if err != nil {
			return nil, 0, err
		}
		for _, v := range stateBelong {
			stateBelongMap[int64(v.ID)] = v.Belong
		}
		for _, t := range req.Type {
			if t == apistructs.IssueTypeRequirement || t == apistructs.IssueTypeEpic {
				requirementRelateIssueIDsMap := make(map[uint64][]uint64)
				issueIndex := make(map[uint64]int)
				// 获取所有的需求id
				var requirementIDs []uint64
				for i, v := range issues {
					if v.Type == apistructs.IssueTypeRequirement || t == apistructs.IssueTypeEpic {
						id := uint64(v.ID)
						issueIndex[id] = i
						requirementIDs = append(requirementIDs, id)
					}
				}

				// 获取需求id对应的关联事件ids
				relations, err := svc.db.GetIssueRelationsByIDs(requirementIDs)
				if err != nil {
					return nil, 0, err
				}
				for _, v := range relations {
					requirementRelateIssueIDsMap[v.IssueID] = append(requirementRelateIssueIDsMap[v.IssueID], v.RelatedIssue)
				}

				// 获取每个需求的IssueSummary
				for requirementID, relatedIDs := range requirementRelateIssueIDsMap {
					reqResult, err := svc.db.IssueStateCount2(relatedIDs)
					if err != nil {
						return nil, 0, err
					}

					var sum apistructs.IssueSummary
					for _, v := range reqResult {
						if stateBelongMap[v.State] == apistructs.IssueStateBelongDone || stateBelongMap[v.State] == apistructs.IssueStateBelongClosed {
							sum.DoneCount++
						} else {
							sum.ProcessingCount++
						}
					}

					issues[issueIndex[requirementID]].IssueSummary = &sum
				}
			}
		}
	}

	return issues, total, nil
}

// GetIssueNumByPros query by issue request and group by project id list, optimization for workbench issue num
func (svc *Issue) GetIssueNumByPros(projectIDS []uint64, req apistructs.IssuePagingRequest) ([]apistructs.IssueNum, error) {
	if len(projectIDS) <= 0 {
		return nil, apierrors.ErrPagingIssues.MissingParameter("projectIDS")
	}
	return svc.db.GetIssueNumByPros(projectIDS, req)
}

// get all undone issues by orgid
func (svc *Issue) PagingForWorkbench(req apistructs.IssuePagingRequest) ([]apistructs.Issue, uint64, error) {
	if (req.IterationID == -1 || (len(req.IterationIDs) == 1 && req.IterationIDs[0] == -1)) && req.OrderBy == "" {
		req.OrderBy = "FIELD(priority, 'LOW', 'NORMAL', 'HIGH', 'URGENT')"
	}
	if req.IterationID != 0 {
		req.IterationIDs = append(req.IterationIDs, req.IterationID)
	}

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	var (
		labelRelationIDs, issueRelationIDs []int64
		isLabel, isIssue                   bool
	)
	if len(req.Label) > 0 {
		isLabel = true
		lrs, err := svc.db.GetLabelRelationsByLabels(apistructs.LabelTypeIssue, req.Label)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
		for _, v := range lrs {
			labelRelationIDs = append(labelRelationIDs, int64(v.RefID))
		}
	}
	if len(req.RelatedIssueIDs) > 0 {
		isIssue = true
		irs, err := svc.db.GetIssueRelationsByIDs(req.RelatedIssueIDs)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
		for _, v := range irs {
			issueRelationIDs = append(issueRelationIDs, int64(v.RelatedIssue))
		}
	}
	if isLabel || isIssue {
		req.IDs = strutil.DedupInt64Slice(append(getRelatedIDs(labelRelationIDs, issueRelationIDs, isLabel, isIssue), req.IDs...))
	}

	// state  key: stateID value:state
	stateMap := make(map[int64]dao.IssueState)
	// filter by state
	if len(req.StateBelongs) > 0 {
		err := svc.FilterByStateBelong(stateMap, &req)
		if err != nil {
			return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
		}
	}
	// paging
	issueModels, total, err := svc.db.PagingIssues(req, isLabel || isIssue)
	if err != nil {
		return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
	}

	issues, err := svc.BatchConvert(issueModels, req.Type, req.IdentityInfo)
	if err != nil {
		return nil, 0, apierrors.ErrPagingIssues.InternalError(err)
	}

	return issues, total, nil
}

func (svc *Issue) RequirementPool() ([]apistructs.Issue, error) {
	return nil, nil
}

// GetIssue 获取事件
func (svc *Issue) GetIssue(req apistructs.IssueGetRequest) (*apistructs.Issue, error) {
	// 请求校验
	if req.ID == 0 {
		return nil, apierrors.ErrGetIssue.MissingParameter("id")
	}
	// 查询事件
	model, err := svc.db.GetIssue(int64(req.ID))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssue.NotFound()
		}
		return nil, apierrors.ErrGetIssue.InternalError(err)
	}
	issue, err := svc.Convert(model, req.IdentityInfo)
	if err != nil {
		return nil, apierrors.ErrGetIssue.InternalError(err)
	}
	return issue, nil
}

// UpdateIssue 更新事件
func (svc *Issue) UpdateIssue(req apistructs.IssueUpdateRequest) error {
	// 请求校验
	if req.ID == 0 {
		return apierrors.ErrUpdateIssue.MissingParameter("id")
	}
	if req.IsEmpty() {
		return nil
	}
	// 获取事件更新前的版本
	issueModel, err := svc.db.GetIssue(int64(req.ID))
	if err != nil {
		return apierrors.ErrGetIssue.InternalError(err)
	}
	//如果是BUG从打开或者重新打开切换状态为已解决，修改责任人为当前用户
	if issueModel.Type == apistructs.IssueTypeBug {
		currentState, err := svc.db.GetIssueStateByID(issueModel.State)
		if err != nil {
			return apierrors.ErrGetIssue.InternalError(err)
		}
		if req.State != nil {
			newState, err := svc.db.GetIssueStateByID(*req.State)
			if err != nil {
				return apierrors.ErrGetIssue.InternalError(err)
			}
			if (currentState.Belong == apistructs.IssueStateBelongOpen || currentState.Belong == apistructs.IssueStateBelongReopen) && newState.Belong == apistructs.IssueStateBelongResloved {
				req.Owner = &req.IdentityInfo.UserID
			}
		}
	}
	canUpdateFields := issueModel.GetCanUpdateFields()
	// 请求传入的需要更新的字段
	changedFields := req.GetChangedFields(canUpdateFields["man_hour"].(string))
	// 检查修改的字段合法性
	if err := svc.checkChangeFields(changedFields); err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err)
	}

	// issueStreamFields 保存字段更新前后的值，用于生成活动记录
	issueStreamFields := make(map[string][]interface{})

	// 遍历请求更新的字段，删除无需更新的字段
	for field, v := range changedFields {
		// man_hour 是否更新已提前判断，跳过是否需要更新的判断
		if field == "man_hour" {
			goto STREAM
		}
		if field == "plan_started_at" {
			// 如已有开始时间，跳过
			if issueModel.PlanStartedAt != nil {
				delete(changedFields, field)
			}
			continue
		}
		if reflect.DeepEqual(v, canUpdateFields[field]) || v == canUpdateFields[field] {
			delete(changedFields, field)
			continue
		}
		if field == "state" {
			currentBelong, err := svc.db.GetIssueStateByID(issueModel.State)
			if err != nil {
				return apierrors.ErrGetIssueState.InternalError(err)
			}
			newBelong, err := svc.db.GetIssueStateByID(*req.State)
			if err != nil {
				return apierrors.ErrGetIssueState.InternalError(err)
			}
			if currentBelong.Belong != apistructs.IssueStateBelongDone && newBelong.Belong == apistructs.IssueStateBelongDone {
				changedFields["finish_time"] = time.Now()
			} else if currentBelong.Belong != apistructs.IssueStateBelongClosed && newBelong.Belong == apistructs.IssueStateBelongClosed {
				changedFields["finish_time"] = time.Now()
			}
			if currentBelong.Belong == apistructs.IssueStateBelongDone && newBelong.Belong != apistructs.IssueStateBelongDone {
				var nilTime *time.Time
				nilTime = nil
				changedFields["finish_time"] = nilTime
			} else if currentBelong.Belong == apistructs.IssueStateBelongClosed && newBelong.Belong != apistructs.IssueStateBelongClosed {
				var nilTime *time.Time
				nilTime = nil
				changedFields["finish_time"] = nilTime
			}
		}
	STREAM:
		issueStreamFields[field] = []interface{}{canUpdateFields[field], v}
	}

	// 校验实际需要更新的字段
	// 校验 state
	if err := svc.checkUpdateStatePermission(issueModel, changedFields, req.IdentityInfo); err != nil {
		return err
	}

	// 更新实际需要更新的字段
	if err := svc.db.UpdateIssue(req.ID, changedFields); err != nil {
		return apierrors.ErrUpdateIssue.InternalError(err)
	}

	// 需求迭代变更时，集联变更需求下的任务迭代
	// if issueModel.Type == apistructs.IssueTypeRequirement {
	// 	taskFields := map[string]interface{}{"iteration_id": req.IterationID}
	// 	if err := svc.db.UpdateIssues(req.ID, taskFields); err != nil {
	// 		return apierrors.ErrUpdateIssue.InternalError(err)
	// 	}
	// }

	// create stream and send issue update event
	if err := svc.CreateStream(req, issueStreamFields); err != nil {
		logrus.Errorf("create issue %d stream err: %v", req.ID, err)
	}

	go monitor.MetricsIssueById(int(req.ID), svc.db, svc.uc, svc.bdl)
	return nil
}

// UpdateIssueType 转换issue类型
func (svc *Issue) UpdateIssueType(req *apistructs.IssueTypeUpdateRequest) (int64, error) {
	issueModel, err := svc.db.GetIssue(req.ID)
	if err != nil {
		return 0, err
	}
	states, err := svc.db.GetIssuesStatesByProjectID(uint64(req.ProjectID), req.Type)
	if err != nil {
		return 0, err
	}
	if len(states) == 0 {
		return 0, apierrors.ErrUpdateIssue.InvalidParameter("该类型缺少默认事件状态")
	}
	issueModel.Type = req.Type
	issueModel.State = int64(states[0].ID)
	if issueModel.Type == apistructs.IssueTypeRequirement {
		issueModel.Stage = ""
		issueModel.Owner = ""
	} else if issueModel.Type == apistructs.IssueTypeBug {
		issueModel.Stage = "codeDevelopment"
		issueModel.Owner = issueModel.Assignee
	} else if issueModel.Type == apistructs.IssueTypeTask {
		issueModel.Stage = "dev"
		issueModel.Owner = issueModel.Assignee
	}
	err = svc.db.UpdateIssueType(&issueModel)
	if err != nil {
		return 0, err
	}
	return int64(issueModel.ID), nil
}

// BatchUpdateIssue 批量更新 issue
func (svc *Issue) BatchUpdateIssue(req *apistructs.IssueBatchUpdateRequest) error {
	if err := req.CheckValid(); err != nil {
		return apierrors.ErrBatchUpdateIssue.InvalidParameter(err)
	}
	// 获取待批量更新记录
	issues, err := svc.db.GetBatchUpdateIssues(req)
	if err != nil {
		return apierrors.ErrBatchUpdateIssue.InternalError(err)
	}

	//如果更新的是状态，单独鉴权
	if req.State != 0 {
		changedFields := map[string]interface{}{"state": req.State}
		for _, v := range issues {
			if err := svc.checkUpdateStatePermission(v, changedFields, req.IdentityInfo); err != nil {
				return apierrors.ErrBatchUpdateIssue.InternalError(errors.Errorf(
					"update to %v failed: no permission or some issue's state couldn't update to %v directly",
					req.State, req.State))
			}
		}
	}

	// 批量更新
	if err := svc.db.BatchUpdateIssues(req); err != nil {
		return apierrors.ErrBatchUpdateIssue.InternalError(err)
	}

	if req.NewIterationID != 0 {
		for _, v := range issues {
			// 需求迭代变更时，集联变更需求下的任务迭代
			if v.Type == apistructs.IssueTypeRequirement {
				taskFields := map[string]interface{}{"iteration_id": req.NewIterationID}
				if err := svc.db.UpdateIssues(uint64(v.ID), taskFields); err != nil {
					logrus.Errorf("[alert] failed to update task iteration, reqID: %+v, err: %v", v.ID, err)
					continue
				}
			}
		}
	}

	if req.State != 0 {
		svc.batchCreateStateChaningStream(req, issues)
	} else if req.Assignee != "" {
		if err := svc.batchCreateAssignChaningStream(req, issues); err != nil {
			return apierrors.ErrBatchUpdateIssue.InternalError(err)
		}
	}

	go func(issues []dao.Issue) {
		for _, v := range issues {
			go monitor.MetricsIssueById(int(v.ID), svc.db, svc.uc, svc.bdl)
		}
	}(issues)

	return nil
}

// Delete .
func (svc *Issue) Delete(issueID uint64, identityInfo apistructs.IdentityInfo) error {
	// 事件详情
	issueModel, err := svc.db.GetIssue(int64(issueID))
	if err != nil {
		return apierrors.ErrDeleteIssue.InternalError(err)
	}

	// Authorize, only project manager & creator can delete
	if !identityInfo.IsInternalClient() {
		if identityInfo.UserID != issueModel.Creator && identityInfo.UserID != issueModel.Assignee {
			access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
				UserID:   identityInfo.UserID,
				Scope:    apistructs.ProjectScope,
				ScopeID:  issueModel.ProjectID,
				Resource: issueModel.Type.GetCorrespondingResource(),
				Action:   apistructs.DeleteAction,
			})
			if err != nil {
				return apierrors.ErrDeleteIssue.InternalError(err)
			}
			if !access.Access {
				return apierrors.ErrDeleteIssue.AccessDenied()
			}

		}
	}

	// 删除需求时，判断需求下是否有任务存在
	// if issueModel.Type == apistructs.IssueTypeRequirement {
	// 	requirementID := int64(issueID)
	// 	listReq := apistructs.IssuePagingRequest{
	// 		IssueListRequest: apistructs.IssueListRequest{
	// 			RequirementID: &requirementID,
	// 		},
	// 		PageNo:   1,
	// 		PageSize: 20,
	// 	}
	// 	_, total, err := svc.db.PagingIssues(listReq)
	// 	if err != nil {
	// 		return apierrors.ErrDeleteIssue.InvalidState("task is not empty")
	// 	}
	// 	if total > 0 {
	// 		return errors.Errorf("task is not empty")
	// 	}
	// }

	// 删除史诗前判断是否关联了事件
	if issueModel.Type == apistructs.IssueTypeEpic {
		relatingIssueIDs, err := svc.db.GetRelatingIssues(uint64(issueModel.ID))
		if err != nil {
			return err
		}
		if len(relatingIssueIDs) > 0 {
			return apierrors.ErrDeleteIssue.InvalidState("史诗下关联了事件,不可删除")
		}
	}

	// 删除issueIssueRelation级连
	if err := svc.db.CleanIssueRelation(issueID); err != nil {
		return apierrors.ErrDeleteIssue.InternalError(err)
	}
	// 删除自定义字段
	if err := svc.db.DeletePropertyRelationByIssueID(int64(issueID)); err != nil {
		return apierrors.ErrDeleteIssue.InternalError(err)
	}
	// 删除测试计划用例关联
	if issueModel.Type == apistructs.IssueTypeBug {
		if err := svc.db.DeleteIssueTestCaseRelationsByIssueIDs([]uint64{issueID}); err != nil {
			return apierrors.ErrDeleteIssue.InternalError(err)
		}
	}

	err = svc.db.DeleteIssue(issueID)
	if err == nil {
		go monitor.MetricsIssueById(int(issueID), svc.db, svc.uc, svc.bdl)
	}

	return err
}

// batchCreateStateChaningStream 批量创建状态转换活动记录
func (svc *Issue) batchCreateStateChaningStream(req *apistructs.IssueBatchUpdateRequest, issues []dao.Issue) {
	// 批量生成状态转换活动记录
	for _, v := range issues {
		streamReq := apistructs.IssueStreamCreateRequest{
			IssueID:      int64(v.ID),
			Operator:     req.UserID,
			StreamType:   apistructs.ISTTransferState,
			StreamParams: apistructs.ISTParam{CurrentState: fmt.Sprintf("%d", v.State), NewState: fmt.Sprintf("%d", v.State)},
		}
		// create stream
		if _, err := svc.stream.Create(&streamReq); err != nil {
			logrus.Errorf("[alert] failed to create issueStream when update issue, req: %+v, err: %v", streamReq, err)
			continue
		}
	}
}

// batchCreateAssignChaningStream 批量生成处理人变更活动记录
func (svc *Issue) batchCreateAssignChaningStream(req *apistructs.IssueBatchUpdateRequest, issues []dao.Issue) error {
	// 批量生成处理人变更活动记录
	userIds := make([]string, 0, len(issues))
	for _, v := range issues {
		userIds = append(userIds, v.Assignee)
	}
	userIds = append(userIds, req.Assignee)

	users, err := svc.uc.FindUsers(userIds)
	if err != nil {
		return err
	}
	userInfo := make(map[string]string, len(users))
	for _, v := range users {
		userInfo[v.ID] = v.Nick
	}

	for _, v := range issues {
		streamReq := apistructs.IssueStreamCreateRequest{
			IssueID:      int64(v.ID),
			Operator:     req.UserID,
			StreamType:   apistructs.ISTChangeAssignee,
			StreamParams: apistructs.ISTParam{CurrentAssignee: userInfo[v.Assignee], NewAssignee: userInfo[req.Assignee]},
		}
		// create stream
		if _, err := svc.stream.Create(&streamReq); err != nil {
			logrus.Errorf("[alert] failed to create issueStream when update issue, req: %+v, err: %v", streamReq, err)
			continue
		}
	}

	return nil
}

// CreateStream 创建事件流
func (svc *Issue) CreateStream(updateReq apistructs.IssueUpdateRequest, streamFields map[string][]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[alert] failed to create issueStream when update issue: %v err: %v", updateReq.ID, fmt.Sprintf("%+v", r))
		}
	}()

	for field, v := range streamFields {
		streamReq := apistructs.IssueStreamCreateRequest{
			IssueID:  int64(updateReq.ID),
			Operator: updateReq.UserID,
		}
		switch field {
		case "title":
			streamReq.StreamType = apistructs.ISTChangeTitle
			streamReq.StreamParams = apistructs.ISTParam{CurrentTitle: v[0].(string), NewTitle: v[1].(string)}
		case "state":
			CurrentState, err := svc.db.GetIssueStateByID(v[0].(int64))
			if err != nil {
				return err
			}
			NewState, err := svc.db.GetIssueStateByID(v[1].(int64))
			if err != nil {
				return err
			}

			streamReq.StreamType = apistructs.ISTTransferState
			streamReq.StreamParams = apistructs.ISTParam{CurrentState: CurrentState.Name, NewState: NewState.Name}
		case "plan_started_at":
			streamReq.StreamType = apistructs.ISTChangePlanStartedAt
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentPlanStartedAt: formatTime(v[0]),
				NewPlanStartedAt:     formatTime(v[1])}
		case "plan_finished_at":
			streamReq.StreamType = apistructs.ISTChangePlanFinishedAt
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentPlanFinishedAt: formatTime(v[0]),
				NewPlanFinishedAt:     formatTime(v[1])}
		case "owner":
			userIds := make([]string, 0, len(v))
			for _, uid := range v {
				userIds = append(userIds, uid.(string))
			}
			users, err := svc.uc.FindUsers(userIds)
			if err != nil {
				return err
			}
			if len(users) != 2 {
				return errors.Errorf("failed to fetch userInfo when create issue stream, %v", v)
			}
			streamReq.StreamType = apistructs.ISTChangeOwner
			streamReq.StreamParams = apistructs.ISTParam{CurrentOwner: users[0].Nick, NewOwner: users[1].Nick}
		case "stage":
			issue, err := svc.db.GetIssue(int64(updateReq.ID))
			if err != nil {
				return err
			}
			if issue.Type == apistructs.IssueTypeTask {
				streamReq.StreamType = apistructs.ISTChangeTaskType
			} else if issue.Type == apistructs.IssueTypeBug {
				streamReq.StreamType = apistructs.ISTChangeBugStage
			}
			project, err := svc.bdl.GetProject(issue.ProjectID)
			if err != nil {
				return err
			}
			stage, err := svc.db.GetIssuesStage(int64(project.OrgID), issue.Type)
			for _, s := range stage {
				if v[0].(string) == s.Value {
					v[0] = s.Name
				}
				if v[1].(string) == s.Value {
					v[1] = s.Name
				}
			}
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentStage: v[0].(string),
				NewStage:     v[1].(string)}
		case "priority":
			streamReq.StreamType = apistructs.ISTChangePriority
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentPriority: v[0].(apistructs.IssuePriority).GetZhName(),
				NewPriority:     v[1].(apistructs.IssuePriority).GetZhName()}
		case "complexity":
			streamReq.StreamType = apistructs.ISTChangeComplexity
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentComplexity: v[0].(apistructs.IssueComplexity).GetZhName(),
				NewComplexity:     v[1].(apistructs.IssueComplexity).GetZhName()}
		case "severity":
			streamReq.StreamType = apistructs.ISTChangeSeverity
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentSeverity: v[0].(apistructs.IssueSeverity).GetZhName(),
				NewSeverity:     v[1].(apistructs.IssueSeverity).GetZhName()}
		case "content":
			// 不显示修改详情
			streamReq.StreamType = apistructs.ISTChangeContent
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentContent: v[0].(string),
				NewContent:     v[1].(string)}
		case "label":
			// 不显示修改详情
			streamReq.StreamType = apistructs.ISTChangeLabel
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentLabel: "1",
				NewLabel:     "2"}
		case "assignee":
			userIds := make([]string, 0, len(v))
			for _, uid := range v {
				userIds = append(userIds, uid.(string))
			}
			users, err := svc.uc.FindUsers(userIds)
			if err != nil {
				return err
			}
			if len(users) != 2 {
				return errors.Errorf("failed to fetch userInfo when create issue stream, %v", v)
			}
			streamReq.StreamType = apistructs.ISTChangeAssignee
			streamReq.StreamParams = apistructs.ISTParam{CurrentAssignee: users[0].Nick, NewAssignee: users[1].Nick}
		case "iteration_id":
			// 迭代
			currentIteration, err := svc.db.GetIteration(uint64(v[0].(int64)))
			if err != nil {
				return err
			}
			newIteration, err := svc.db.GetIteration(uint64(v[1].(int64)))
			if err != nil {
				return err
			}
			streamReq.StreamType = apistructs.ISTChangeIteration
			streamReq.StreamParams = apistructs.ISTParam{CurrentIteration: currentIteration.Title, NewIteration: newIteration.Title}
		case "man_hour":
			// 工时
			var currentManHour, newManHour apistructs.IssueManHour
			json.Unmarshal([]byte(v[0].(string)), &currentManHour)
			json.Unmarshal([]byte(v[1].(string)), &newManHour)
			streamReq.StreamType = apistructs.ISTChangeManHour
			streamReq.StreamParams = apistructs.ISTParam{
				CurrentEstimateTime:  currentManHour.GetFormartTime("EstimateTime"),
				CurrentElapsedTime:   currentManHour.GetFormartTime("ElapsedTime"),
				CurrentRemainingTime: currentManHour.GetFormartTime("RemainingTime"),
				CurrentStartTime:     currentManHour.StartTime,
				CurrentWorkContent:   currentManHour.WorkContent,
				NewEstimateTime:      newManHour.GetFormartTime("EstimateTime"),
				NewElapsedTime:       newManHour.GetFormartTime("ElapsedTime"),
				NewRemainingTime:     newManHour.GetFormartTime("RemainingTime"),
				NewStartTime:         newManHour.StartTime,
				NewWorkContent:       newManHour.WorkContent,
			}
		default:
			continue
		}

		// create stream and send issue event
		if _, err := svc.stream.Create(&streamReq); err != nil {
			logrus.Errorf("[alert] failed to create issueStream when update issue, req: %+v, err: %v", streamReq, err)
		}
	}

	return nil
}

// checkUpdateStatePermission 判断当前状态是否能更新至新状态，如果能，再判断是否有权限更新至新状态
func (svc *Issue) checkUpdateStatePermission(model dao.Issue, changedFields map[string]interface{}, identityInfo apistructs.IdentityInfo) error {
	// 校验新状态是否合法
	newStateInterface, ok := changedFields["state"]
	if !ok {
		return nil
	}
	newState, ok := newStateInterface.(int64)
	if !ok {
		return apierrors.ErrUpdateIssueState.InvalidParameter(fmt.Sprintf("%v", changedFields["state_id"]))
	}
	//if !svc.ConvertWithoutButton(model).ValidState(string(newState)) {
	//	return apierrors.ErrUpdateIssueState.InvalidParameter(newState)
	//}

	// 权限校验只校验目标状态
	permCheckItem := makeButtonCheckPermItem(model, newState)
	permCheckItems := map[string]bool{permCheckItem: false}

	_, err := svc.generateButton(model, identityInfo, permCheckItems, nil, nil, nil)
	if err != nil {
		return err
	}

	if !permCheckItems[permCheckItem] {
		return apierrors.ErrUpdateIssueState.InvalidParameter(newState)
	}

	return nil
}

// GetIssuesByIssueIDs 通过issueIDs获取事件列表
func (svc *Issue) GetIssuesByIssueIDs(issueIDs []uint64, identityInfo apistructs.IdentityInfo) ([]apistructs.Issue, error) {
	issueModels, err := svc.db.GetIssueByIssueIDs(issueIDs)
	if err != nil {
		return nil, err
	}

	issues, err := svc.BatchConvert(issueModels, apistructs.IssueTypes, identityInfo)
	if err != nil {
		return nil, apierrors.ErrPagingIssues.InternalError(err)
	}

	return issues, nil
}

// formatTime format time to string.
// If input is nil, return "";
// return string with custom format or default format("2006-01-02")
func formatTime(input interface{}, format ...string) string {
	if reflect.ValueOf(input).IsNil() {
		return ""
	}
	if len(format) > 0 {
		return input.(*time.Time).Format(format[0])
	}
	return input.(*time.Time).Format("2006-01-02")
}

// checkChangeFields 更新事件时检查需要更新的字段
func (svc *Issue) checkChangeFields(fields map[string]interface{}) error {
	// 检查迭代是否存在
	if _, ok := fields["iteration_id"]; !ok {
		return nil
	}

	iterationID := fields["iteration_id"].(int64)
	if iterationID == -1 {
		return nil
	}

	iteration, err := svc.db.GetIteration(uint64(iterationID))
	if err != nil || iteration == nil {
		return errors.New("the iteration does not exit")
	}

	return nil
}

// getRelatedIDs 当同时通过lable和issue关联关系过滤时，需要取交集
func getRelatedIDs(lableRelationIDs []int64, issueRelationIDs []int64, isLabel, isIssue bool) []int64 {
	// 取交集
	if isLabel && isIssue {
		return strutil.IntersectionInt64Slice(lableRelationIDs, issueRelationIDs)
	}

	if isLabel {
		return lableRelationIDs
	}

	if isIssue {
		return issueRelationIDs
	}

	return nil
}

// GetIssueManHourSum 事件流任务总和查询
func (svc *Issue) GetIssueManHourSum(req apistructs.IssuesStageRequest) (*apistructs.IssueManHourSumResponse, error) {
	// 请求校验
	if req.RangeID < 1 && (req.StatisticRange != "iteration" || req.RangeID != -1) {
		return nil, apierrors.ErrGetIssueManHourSum.MissingParameter("rangeId")
	}
	// 查询事件
	res, err := svc.db.GetIssueManHourSum(req)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssueManHourSum.NotFound()
		}
		return nil, apierrors.ErrGetIssueManHourSum.InternalError(err)
	}
	return &res, nil
}

// GetIssueBugPercentage 缺陷率
func (svc *Issue) GetIssueBugPercentage(req apistructs.IssuesStageRequest) (*apistructs.IssueBugPercentageResponse, error) {
	// 请求校验
	if req.RangeID < 1 && (req.StatisticRange != "iteration" || req.RangeID != -1) {
		return nil, apierrors.ErrGetIssueBugPercentage.MissingParameter("rangeId")
	}
	// 查询事件
	res, total, err := svc.db.GetIssueBugByRange(req)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssueBugPercentage.NotFound()
		}
		return nil, apierrors.ErrGetIssueBugPercentage.InternalError(err)
	}
	bugs := map[string]float32{}
	for _, each := range res {
		bugs[each.Stage]++
	}
	ans := apistructs.IssueBugPercentageResponse{}
	for k, v := range bugs {
		if k == "" {
			continue
		}
		ans.BugPercentage = append(ans.BugPercentage, apistructs.Percentage{
			Name:  k,
			Value: v / total,
		})
	}
	return &ans, nil
}

// GetIssueBugStatusPercentage 缺陷状态
func (svc *Issue) GetIssueBugStatusPercentage(req apistructs.IssuesStageRequest) ([]apistructs.IssueBugStatusPercentageResponse, error) {
	// 请求校验
	if req.RangeID < 1 && (req.StatisticRange != "iteration" || req.RangeID != -1) {
		return nil, apierrors.ErrGetIssueBugStatusPercentage.MissingParameter("rangeId")
	}
	// 查询事件
	res, _, err := svc.db.GetIssueBugByRange(req)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssueBugStatusPercentage.NotFound()
		}
		return nil, apierrors.ErrGetIssueBugStatusPercentage.InternalError(err)
	}
	bugs := map[string]float32{}
	var stateID []int64
	for _, each := range res {
		stateID = append(stateID, each.State)
		bugs[each.Stage]++
	}
	// state ID 对应 state name
	stateName := map[int64]string{}
	issueState, err := svc.db.GetIssueStateByIDs(stateID)
	if err != nil {
		return nil, apierrors.ErrGetIssueBugStatusPercentage.InternalError(err)
	}
	for _, state := range issueState {
		stateName[int64(state.ID)] = state.Name
	}

	var ans []apistructs.IssueBugStatusPercentageResponse
	for k, v := range bugs {
		if k == "" {
			continue
		}
		state := map[int64]float32{}
		for _, each := range res {
			if each.Stage != k {
				continue
			}
			state[each.State]++
		}
		status := apistructs.IssueBugStatusPercentage{}
		for key, value := range state {
			status.Status = append(status.Status, apistructs.Percentage{
				Name:  stateName[key],
				Value: value / v,
			})
		}
		ans = append(ans, apistructs.IssueBugStatusPercentageResponse{
			StageName: k,
			Status:    status,
		})
	}
	return ans, nil
}

// GetIssueBugSeverityPercentage 缺陷等级
func (svc *Issue) GetIssueBugSeverityPercentage(req apistructs.IssuesStageRequest) ([]apistructs.IssueBugSeverityPercentageResponse, error) {
	// 请求校验
	if req.RangeID < 1 && (req.StatisticRange != "iteration" || req.RangeID != -1) {
		return nil, apierrors.ErrGetIssueBugSeverityPercentage.MissingParameter("rangeId")
	}
	// 查询事件
	res, _, err := svc.db.GetIssueBugByRange(req)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssueBugSeverityPercentage.NotFound()
		}
		return nil, apierrors.ErrGetIssueBugSeverityPercentage.InternalError(err)
	}
	bugs := map[string]float32{}
	var stateID []int64
	for _, each := range res {
		stateID = append(stateID, each.State)
		bugs[each.Stage]++
	}

	var ans []apistructs.IssueBugSeverityPercentageResponse
	for k, v := range bugs {
		if k == "" {
			continue
		}
		severityNum := map[apistructs.IssueSeverity]float32{}
		for _, each := range res {
			if each.Stage != k {
				continue
			}
			severityNum[each.Severity]++
		}
		severity := apistructs.IssueBugSeverityPercentage{}
		for key, value := range severityNum {
			severity.Severity = append(severity.Severity, apistructs.Percentage{
				Name:  key.GetZhName(),
				Value: value / v,
			})
		}
		ans = append(ans, apistructs.IssueBugSeverityPercentageResponse{
			StageName: k,
			Severity:  severity,
		})
	}
	return ans, nil
}

// FilterByStateBelong 根据主状态过滤
func (svc *Issue) FilterByStateBelong(stateMap map[int64]dao.IssueState, req *apistructs.IssuePagingRequest) error {
	var states []int64
	belongMap := make(map[apistructs.IssueStateBelong]bool)
	for _, v := range req.StateBelongs {
		belongMap[v] = true
	}
	// 获取主状态下的子状态
	projectStates, err := svc.db.GetIssuesStatesByProjectID(req.ProjectID, "")

	if err != nil {
		return err
	}
	for _, s := range projectStates {
		stateMap[int64(s.ID)] = s
		if belongMap[s.Belong] {
			states = append(states, int64(s.ID))
		}
	}

	if len(req.State) == 0 {
		req.State = states
	} else {
		var newState []int64
		for _, v := range states {
			for _, vv := range req.State {
				if v == vv {
					newState = append(newState, vv)
					break
				}
			}
		}
		req.State = newState
	}
	return nil
}

// FilterByStateBelong
func (svc *Issue) FilterByStateBelongForPros(stateMap map[int64]dao.IssueState, projectIDList []uint64, req *apistructs.IssuePagingRequest) error {
	var states []int64
	belongMap := make(map[apistructs.IssueStateBelong]bool)
	for _, v := range req.StateBelongs {
		belongMap[v] = true
	}
	projectStates, err := svc.db.GetIssuesStatesByProjectIDList(projectIDList)

	if err != nil {
		return err
	}
	for _, s := range projectStates {
		stateMap[int64(s.ID)] = s
		if belongMap[s.Belong] {
			states = append(states, int64(s.ID))
		}
	}

	if len(req.State) == 0 {
		req.State = states
	} else {
		var newState []int64
		for _, v := range states {
			for _, vv := range req.State {
				if v == vv {
					newState = append(newState, vv)
					break
				}
			}
		}
		req.State = newState
	}
	return nil
}

// StateCheckPermission 事件状态Button鉴权
func (svc *Issue) StateCheckPermission(req *apistructs.PermissionCheckRequest, st int64, ed int64) (bool, error) {
	logrus.Debugf("invoke permission, time: %s, req: %+v", time.Now().Format(time.RFC3339), req)
	// 是否是内部服务账号
	resp, err := svc.bdl.StateCheckPermission(req)
	if err != nil {
		return false, err
	}
	if resp.Access {
		return true, nil
	}
	for _, role := range resp.Roles {
		rp, err := svc.db.GetIssueStatePermission(role, st, ed)
		if err != nil {
			return false, err
		}
		if rp != nil {
			return true, nil
		}
	}
	return false, nil
}

// Subscribe subscribe issue
func (svc *Issue) Subscribe(id int64, identityInfo apistructs.IdentityInfo) error {
	issue, err := svc.db.GetIssue(id)
	if err != nil {
		return err
	}

	if !identityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issue.ProjectID,
			Resource: issue.Type.GetCorrespondingResource(),
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrSubscribeIssue.InternalError(err)
		}
		if !access.Access {
			return apierrors.ErrSubscribeIssue.AccessDenied()
		}
	}

	is, err := svc.db.GetIssueSubscriber(id, identityInfo.UserID)
	if err != nil {
		return err
	}
	if is != nil {
		return errors.New("already subscribed")
	}

	// 创建 issue
	create := dao.IssueSubscriber{
		IssueID: int64(issue.ID),
		UserID:  identityInfo.UserID,
	}

	return svc.db.CreateIssueSubscriber(&create)
}

// Unsubscribe unsubscribe issue
func (svc *Issue) Unsubscribe(id int64, identityInfo apistructs.IdentityInfo) error {
	issue, err := svc.db.GetIssue(id)
	if err != nil {
		return err
	}

	if !identityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issue.ProjectID,
			Resource: issue.Type.GetCorrespondingResource(),
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrSubscribeIssue.InternalError(err)
		}
		if !access.Access {
			return apierrors.ErrSubscribeIssue.AccessDenied()
		}
	}

	return svc.db.DeleteIssueSubscriber(id, identityInfo.UserID)
}

// BatchUpdateIssuesSubscriber batch update issue subscriber
func (svc *Issue) BatchUpdateIssuesSubscriber(req apistructs.IssueSubscriberBatchUpdateRequest) error {
	issue, err := svc.db.GetIssue(req.IssueID)
	if err != nil {
		return err
	}

	identityInfo := req.IdentityInfo
	if !identityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issue.ProjectID,
			Resource: issue.Type.GetCorrespondingResource(),
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrSubscribeIssue.InternalError(err)
		}
		if !access.Access {
			return apierrors.ErrSubscribeIssue.AccessDenied()
		}
	}

	oldSubscribers, err := svc.db.GetIssueSubscribersByIssueID(req.IssueID)
	if err != nil {
		return err
	}

	subscriberMap := make(map[string]struct{}, 0)
	req.Subscribers = strutil.DedupSlice(req.Subscribers)
	for _, v := range req.Subscribers {
		subscriberMap[v] = struct{}{}
	}

	var needDeletedSubscribers []string
	for _, v := range oldSubscribers {
		_, exist := subscriberMap[v.UserID]
		if exist {
			delete(subscriberMap, v.UserID)
		} else {
			needDeletedSubscribers = append(needDeletedSubscribers, v.UserID)
		}
	}

	if len(needDeletedSubscribers) != 0 {
		if err := svc.db.BatchDeleteIssueSubscribers(req.IssueID, needDeletedSubscribers); err != nil {
			return err
		}
	}

	var subscribers []dao.IssueSubscriber
	issueID := int64(req.IssueID)
	for k := range subscriberMap {
		subscribers = append(subscribers, dao.IssueSubscriber{IssueID: issueID, UserID: k})
	}
	if len(subscribers) != 0 {
		if err := svc.db.BatchCreateIssueSubscribers(subscribers); err != nil {
			return err
		}
	}

	return nil
}
