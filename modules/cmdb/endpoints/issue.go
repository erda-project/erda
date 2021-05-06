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

package endpoints

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateIssue 创建事件
func (e *Endpoints) CreateIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IssueCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssue.InvalidParameter(err).ToResp(), nil
	}
	//如果新建BUG 责任人默认为处理人
	if createReq.Type == apistructs.IssueTypeBug {
		createReq.Owner = createReq.Assignee
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssue.NotLogin().ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// issue 创建 校验用户在 当前 project 下是否拥有 CREATE ${ISSUE_TYPE} 权限
		if createReq.ProjectID == 0 {
			return apierrors.ErrCreateIssue.MissingParameter("projectID").ToResp(), nil
		}
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  createReq.ProjectID,
			Resource: createReq.Type.GetCorrespondingResource(),
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrCreateIssue.AccessDenied().ToResp(), nil
		}
		createReq.External = true
	}

	// 创建 issue
	issue, err := e.issue.Create(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  createReq.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	// 关联 测试计划用例
	if len(createReq.TestPlanCaseRelIDs) > 0 {
		// 批量查询测试计划用例
		testPlanCaseRels, err := e.bdl.ListTestPlanCaseRel(createReq.TestPlanCaseRelIDs)
		if err != nil {
			return nil, err
		}
		var issueCaseRels []dao.IssueTestCaseRelation
		for _, rel := range testPlanCaseRels {
			issueCaseRels = append(issueCaseRels, dao.IssueTestCaseRelation{
				IssueID:           uint64(issue.ID),
				TestPlanID:        rel.TestPlanID,
				TestPlanCaseRelID: rel.ID,
				TestCaseID:        rel.TestCaseID,
				CreatorID:         createReq.UserID,
			})
		}
		if err := e.db.BatchCreateIssueTestCaseRelations(issueCaseRels); err != nil {
			return nil, apierrors.ErrBatchCreateIssueTestCaseRel.InternalError(err)
		}
	}

	// 添加标签关联关系
	labels, err := e.label.ListByNames(issue.ProjectID, createReq.Labels)
	if err != nil {
		return apierrors.ErrCreateIssue.InternalError(err).ToResp(), nil
	}
	for _, v := range labels {
		lr := &dao.LabelRelation{
			LabelID: uint64(v.ID),
			RefType: apistructs.LabelTypeIssue,
			RefID:   uint64(issue.ID),
		}
		if err := e.label.CreateRelation(lr); err != nil {
			return apierrors.ErrCreateIssue.InternalError(err).ToResp(), nil
		}
	}

	// data 返回 ID
	return httpserver.OkResp(issue.ID)
}

// PagingIssues 分页查询事件
func (e *Endpoints) PagingIssues(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var pageReq apistructs.IssuePagingRequest
	if err := e.queryStringDecoder.Decode(&pageReq, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssues.InvalidParameter(err).ToResp(), nil
	}

	switch pageReq.OrderBy {
	case "":
	case "planStartedAt":
		pageReq.OrderBy = "plan_started_at"
	case "planFinishedAt":
		pageReq.OrderBy = "plan_finished_at"
	case "assignee":
		pageReq.OrderBy = "assignee"
	default:
		return apierrors.ErrPagingIssues.InvalidParameter("orderBy").ToResp(), nil
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingIssues.NotLogin().ToResp(), nil
	}
	pageReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// issue 分页查询 校验用户在 当前 project 下是否拥有 GET ${project} 权限
		if pageReq.ProjectID == 0 {
			return apierrors.ErrPagingIssues.MissingParameter("projectID").ToResp(), nil
		}
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  pageReq.ProjectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrPagingIssues.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrPagingIssues.AccessDenied().ToResp(), nil
		}
		// 外部创建的事件
		pageReq.External = true
	}

	// 分页查询
	issues, total, err := e.issue.Paging(pageReq)
	if err != nil {
		return apierrors.ErrPagingIssues.InternalError(err).ToResp(), nil
	}
	// userIDs
	userIDs := pageReq.GetUserIDs()
	for _, issue := range issues {
		userIDs = append(userIDs, issue.Creator, issue.Assignee, issue.Owner)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	// 返回
	return httpserver.OkResp(apistructs.IssuePagingResponseData{
		Total: total,
		List:  issues,
	}, userIDs)
}

// ExportExcelIssue 导出事件到 excel
func (e *Endpoints) ExportExcelIssue(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (err error) {
	var pageReq apistructs.IssueExportExcelRequest
	if err := e.queryStringDecoder.Decode(&pageReq, r.URL.Query()); err != nil {
		return apierrors.ErrExportExcelIssue.InvalidParameter(err)
	}
	switch pageReq.OrderBy {
	case "":
	case "planStartedAt":
		pageReq.OrderBy = "plan_started_at"
	case "planFinishedAt":
		pageReq.OrderBy = "plan_finished_at"
	default:
		return apierrors.ErrExportExcelIssue.InvalidParameter("orderBy")
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrExportExcelIssue.NotLogin()
	}
	pageReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// issue 分页查询 校验用户在 当前 project 下是否拥有 GET ${project} 权限
		if pageReq.ProjectID == 0 {
			return apierrors.ErrPagingIssues.MissingParameter("projectID")
		}
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  pageReq.ProjectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrExportExcelIssue.InternalError(err)
		}
		if !access {
			return apierrors.ErrExportExcelIssue.AccessDenied()
		}
		// 外部创建的事件
		pageReq.External = true
	}
	pageReq.PageNo = 1
	pageReq.PageSize = 99999
	// 分页查询
	issues, _, err := e.issue.Paging(pageReq.IssuePagingRequest)
	if err != nil {
		return apierrors.ErrExportExcelIssue.InternalError(err)
	}
	pro, err := e.issueProperty.GetBatchProperties(pageReq.OrgID, pageReq.Type)
	if err != nil {
		return apierrors.ErrExportExcelIssue.InternalError(err)
	}
	reader, tablename, err := e.issue.ExportExcel(issues, pro, pageReq.ProjectID, pageReq.IsDownload)
	if err != nil {
		return apierrors.ErrExportExcelIssue.InternalError(err)
	}
	w.Header().Add("Content-Disposition", "attachment;fileName="+tablename+".xlsx")
	w.Header().Add("Content-Type", "application/vnd.ms-excel")

	if _, err := io.Copy(w, reader); err != nil {
		return apierrors.ErrExportExcelIssue.InternalError(err)
	}
	return nil
}

// ImportExcelIssue 从excel导入事项
func (e *Endpoints) ImportExcelIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrImportExcelIssue.NotLogin().ToResp(), nil
	}

	var req apistructs.IssueImportExcelRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrImportExcelIssue.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// TODO:鉴权
	if !identityInfo.IsInternalClient() {
		if req.ProjectID == 0 {
			return apierrors.ErrImportExcelIssue.MissingParameter("projectID").ToResp(), nil
		}
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectID,
			Resource: apistructs.IssueImportResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrImportExcelIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrImportExcelIssue.AccessDenied().ToResp(), nil
		}
		// 外部创建的事件
	}

	properties, err := e.issueProperty.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: req.OrgID, PropertyIssueType: req.Type})
	memberQuery := &apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(req.ProjectID),
		PageNo:    1,
		PageSize:  99999,
	}
	_, members, err := e.member.List(memberQuery)
	if err != nil {
		return apierrors.ErrImportExcelIssue.InternalError(err).ToResp(), nil
	}
	res, err := e.issue.ImportExcel(req, r, properties, members, e.label, e.issueProperty)
	if err != nil {
		return apierrors.ErrImportExcelIssue.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(res)
}

// UpdateIssue 更新事件
func (e *Endpoints) UpdateIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssueUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}
	updateReq.ID = id
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssue.NotLogin().ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo
	// 事件详情
	issueModel, err := e.DBClient().GetIssue(int64(id))
	if err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
		if identityInfo.UserID != issueModel.Creator && identityInfo.UserID != issueModel.Assignee {
			access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
				UserID:   identityInfo.UserID,
				Scope:    apistructs.ProjectScope,
				ScopeID:  issueModel.ProjectID,
				Resource: issueModel.Type.GetCorrespondingResource(),
				Action:   apistructs.UpdateAction,
			})
			if err != nil {
				return apierrors.ErrUpdateIssue.InternalError(err).ToResp(), nil
			}
			if !access {
				return apierrors.ErrUpdateIssue.AccessDenied().ToResp(), nil
			}
		}
	}
	// 更新
	if err := e.issue.UpdateIssue(updateReq); err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  issueModel.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	// 更新 关联测试计划用例
	if len(updateReq.TestPlanCaseRelIDs) > 0 && !updateReq.RemoveTestPlanCaseRelIDs {
		// 批量查询测试计划用例
		testPlanCaseRels, err := e.bdl.ListTestPlanCaseRel(updateReq.TestPlanCaseRelIDs)
		if err != nil {
			return nil, err
		}
		// 批量删除原有关联
		if err := e.db.DeleteIssueTestCaseRelationsByIssueID(updateReq.ID); err != nil {
			return nil, apierrors.ErrDeleteIssueTestCaseRel.InternalError(err)
		}
		// 批量创建关联
		var issueCaseRels []dao.IssueTestCaseRelation
		for _, rel := range testPlanCaseRels {
			issueCaseRels = append(issueCaseRels, dao.IssueTestCaseRelation{
				IssueID:           updateReq.ID,
				TestPlanID:        rel.TestPlanID,
				TestPlanCaseRelID: rel.ID,
				TestCaseID:        rel.TestCaseID,
				CreatorID:         updateReq.UserID,
			})
		}
		if err := e.db.BatchCreateIssueTestCaseRelations(issueCaseRels); err != nil {
			return nil, apierrors.ErrBatchCreateIssueTestCaseRel.InternalError(err)
		}
	}
	// 批量删除原有关联
	if updateReq.RemoveTestPlanCaseRelIDs {
		if err := e.db.DeleteIssueTestCaseRelationsByIssueID(updateReq.ID); err != nil {
			return nil, apierrors.ErrDeleteIssueTestCaseRel.InternalError(err)
		}
	}

	// 标签是否修改
	issue, err := e.issue.GetIssue(apistructs.IssueGetRequest{ID: updateReq.ID})
	if err != nil {
		return nil, apierrors.ErrGetLabels.InternalError(err)
	}
	currentLabelMap := make(map[string]bool)
	newLabelMap := make(map[string]bool)
	for _, v := range issue.Labels {
		currentLabelMap[v] = true
	}
	for _, v := range updateReq.Labels {
		newLabelMap[v] = true
	}
	if reflect.DeepEqual(currentLabelMap, newLabelMap) == false {
		// 删除该事件已有标签关联关系
		if err := e.label.DeleteRelations(apistructs.LabelTypeIssue, updateReq.ID); err != nil {
			return apierrors.ErrUpdateIssue.InternalError(err).ToResp(), nil
		}
		labels, err := e.label.ListByNames(issueModel.ProjectID, updateReq.Labels)
		if err != nil {
			return apierrors.ErrUpdateIssue.InternalError(err).ToResp(), nil
		}
		// 重新添加标签关联关系
		for _, v := range labels {
			lr := &dao.LabelRelation{
				LabelID: uint64(v.ID),
				RefType: apistructs.LabelTypeIssue,
				RefID:   updateReq.ID,
			}
			if err := e.label.CreateRelation(lr); err != nil {
				return apierrors.ErrUpdateIssue.InternalError(err).ToResp(), nil
			}
		}

		// 生成活动记录
		// issueStreamFields 保存字段更新前后的值，用于生成活动记录
		issueStreamFields := make(map[string][]interface{})
		issueStreamFields["label"] = []interface{}{"1", "2"}
		_ = e.issue.CreateStream(updateReq, issueStreamFields)

	}

	return httpserver.OkResp(issueModel.ID)
}

// UpdateIssueType 转换事件类型
func (e *Endpoints) UpdateIssueType(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssueTypeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrPagingIssueStream.InvalidParameter(err).ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	updateReq.IdentityInfo = identityInfo
	//鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(updateReq.ProjectID),
			Resource: apistructs.IssueTypeResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrBatchUpdateIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrBatchUpdateIssue.AccessDenied().ToResp(), nil
		}
	}
	id, err := e.issue.UpdateIssueType(&updateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// 删除原有类型配置的自定义字段
	if err := e.issueProperty.DeletePropertyRelation(updateReq.ID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(id)
}

// DeleteIssue
func (e *Endpoints) DeleteIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteIssue.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteIssue.NotLogin().ToResp(), nil
	}
	issue, err := e.issue.GetIssue(apistructs.IssueGetRequest{ID: id, IdentityInfo: identityInfo})
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// delete issue
	if err := e.issue.Delete(id, identityInfo); err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  issue.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp(issue)
}

// BatchUpdateIssue 批量更新事件(仅支持需求、缺陷更新处理人、状态、迭代)
func (e *Endpoints) BatchUpdateIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssueBatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrBatchUpdateIssue.InvalidParameter(err).ToResp(), nil
	}

	if updateReq.ProjectID == 0 {
		return apierrors.ErrBatchUpdateIssue.MissingParameter("projectID").ToResp(), nil
	}

	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchUpdateIssue.NotLogin().ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo

	if updateReq.CurrentIterationID != 0 {
		updateReq.CurrentIterationIDs = append(updateReq.CurrentIterationIDs, updateReq.CurrentIterationID)
	}
	updateReq.CurrentIterationIDs = strutil.DedupInt64Slice(updateReq.CurrentIterationIDs)
	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  updateReq.ProjectID,
			Resource: updateReq.Type.GetCorrespondingResource(),
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrBatchUpdateIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrBatchUpdateIssue.AccessDenied().ToResp(), nil
		}
	}
	// 更新
	if err := e.issue.BatchUpdateIssue(&updateReq); err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  updateReq.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp("update succ")
}

// GetIssue 事件详情
func (e *Endpoints) GetIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetIssue.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssue.NotLogin().ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssue(apistructs.IssueGetRequest{ID: id, IdentityInfo: identityInfo})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issue.ProjectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrGetIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrGetIssue.AccessDenied().ToResp(), nil
		}
	}
	// userIDs
	userIDs := strutil.DedupSlice([]string{issue.Creator, issue.Assignee, issue.Owner}, true)

	return httpserver.OkResp(issue, userIDs)
}

// PagingIssueStreams 事件流分页查询
func (e *Endpoints) PagingIssueStreams(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrPagingIssueStream.InvalidParameter(err).ToResp(), nil
	}
	var pagingReq apistructs.IssueStreamPagingRequest
	if err := e.queryStringDecoder.Decode(&pagingReq, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssueStream.InvalidParameter(err).ToResp(), nil
	}
	pagingReq.IssueID = id
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingIssueStream.NotLogin().ToResp(), nil
	}
	// issue 详情
	issueModel, err := e.DBClient().GetIssue(int64(pagingReq.IssueID))
	if err != nil {
		return apierrors.ErrPagingIssueStream.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issueModel.ProjectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrPagingIssueStream.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrPagingIssueStream.AccessDenied().ToResp(), nil
		}
	}
	// 查询事件流列表
	streamRespData, err := e.issueStream.Paging(&pagingReq)
	if err != nil {
		return apierrors.ErrPagingIssueStream.InternalError(err).ToResp(), nil
	}
	// userIDs
	var userIDs []string
	for _, stream := range streamRespData.List {
		userIDs = append(userIDs, stream.Operator)
	}
	userIDs = strutil.DedupSlice(userIDs, true)

	return httpserver.OkResp(streamRespData, userIDs)
}

// GetIssueManHourSum 事件流任务总和查询
func (e *Endpoints) GetIssueManHourSum(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuesStageRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssues.InvalidParameter(err).ToResp(), nil
	}
	if req.StatisticRange != "project" && req.StatisticRange != "iteration" {
		return apierrors.ErrGetIssueManHourSum.InvalidParameter("statisticRange").ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssueManHourSum(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issue)
}

// GetIssueBugPercentage 缺陷率查询
func (e *Endpoints) GetIssueBugPercentage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuesStageRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueBugPercentage.InvalidParameter(err).ToResp(), nil
	}
	if req.StatisticRange != "project" && req.StatisticRange != "iteration" {
		return apierrors.ErrGetIssueBugPercentage.InvalidParameter("statisticRange").ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssueBugPercentage(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issue)
}

// GetIssueBugPercentage 缺陷状态发布查询
func (e *Endpoints) GetIssueBugStatusPercentage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuesStageRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueBugStatusPercentage.InvalidParameter(err).ToResp(), nil
	}
	if req.StatisticRange != "project" && req.StatisticRange != "iteration" {
		return apierrors.ErrGetIssueBugStatusPercentage.InvalidParameter("statisticRange").ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssueBugStatusPercentage(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issue)
}

// GetIssueBugPercentage 缺陷状态发布查询
func (e *Endpoints) GetIssueBugSeverityPercentage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuesStageRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueBugSeverityPercentage.InvalidParameter(err).ToResp(), nil
	}
	if req.StatisticRange != "project" && req.StatisticRange != "iteration" {
		return apierrors.ErrGetIssueBugSeverityPercentage.InvalidParameter("statisticRange").ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssueBugSeverityPercentage(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issue)
}
