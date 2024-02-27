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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateIteration 创建迭代
func (e *Endpoints) CreateIteration(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IterationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIteration.InvalidParameter(err).ToResp(), nil
	}
	if err := createReq.Check(); err != nil {
		return apierrors.ErrCreateIteration.InvalidParameter(err).ToResp(), nil
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIteration.NotLogin().ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		if createReq.ProjectID == 0 {
			return apierrors.ErrCreateIteration.InvalidParameter("missing projectID").ToResp(), nil
		}
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  createReq.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateIteration.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateIteration.AccessDenied().ToResp(), nil
		}
	}

	// 重名检测
	iteration, err := e.iteration.GetByTitle(createReq.ProjectID, createReq.Title)
	if err != nil {
		return apierrors.ErrCreateIteration.InternalError(err).ToResp(), nil
	}
	if iteration != nil {
		return apierrors.ErrCreateIteration.AlreadyExists().ToResp(), nil
	}

	// 创建 iteration
	id, err := e.iteration.Create(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	labels, err := e.bdl.ListLabelByNameAndProjectID(createReq.ProjectID, createReq.Labels)
	if err != nil {
		return nil, apierrors.ErrCreateIteration.InternalError(err)
	}
	for _, v := range labels {
		lr := &dao.LabelRelation{
			LabelID: uint64(v.ID),
			RefType: apistructs.LabelTypeIteration,
			RefID:   strconv.FormatUint(id.ID, 10),
		}
		if err := e.db.CreateLabelRelation(lr); err != nil {
			return nil, apierrors.ErrCreateIssue.InternalError(err)
		}
	}

	// 更新项目活跃时间
	if err := e.bdl.UpdateProjectActiveTime(apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  createReq.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp(id)
}

// UpdateIteration 更新 iteration
func (e *Endpoints) UpdateIteration(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIteration.NotLogin().ToResp(), nil
	}

	iterationID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateIteration.InvalidParameter("id").ToResp(), nil
	}
	iteration, err := e.iteration.Get(iterationID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 解析 request
	var updateReq apistructs.IterationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrUpdateIteration.InvalidParameter(err).ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateIteration.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateIteration.AccessDenied().ToResp(), nil
		}
	}
	// 更新 iteration
	if err := e.iteration.Update(iterationID, updateReq); err != nil {
		return errorresp.ErrResp(err)
	}

	// update labels
	currentLabelMap := make(map[string]bool)
	newLabelMap := make(map[string]bool)
	currentLabels, _ := e.getIterationLabelDetails(iterationID)
	for _, v := range currentLabels {
		currentLabelMap[v] = true
	}
	for _, v := range updateReq.Labels {
		newLabelMap[v] = true
	}
	if len(updateReq.Labels) > 0 && !reflect.DeepEqual(currentLabelMap, newLabelMap) {
		if err = e.db.DeleteLabelRelations(apistructs.LabelTypeIteration, strconv.FormatUint(iterationID, 10), nil); err != nil {
			return apierrors.ErrUpdateIteration.InternalError(err).ToResp(), nil
		}
		labels, err := e.bdl.ListLabelByNameAndProjectID(iteration.ProjectID, updateReq.Labels)
		if err != nil {
			return apierrors.ErrUpdateIteration.InternalError(err).ToResp(), nil
		}
		labelRelations := make([]dao.LabelRelation, 0, len(labels))
		for _, v := range labels {
			labelRelations = append(labelRelations, dao.LabelRelation{
				LabelID: uint64(v.ID),
				RefType: apistructs.LabelTypeIteration,
				RefID:   strconv.FormatUint(iterationID, 10),
			})
		}
		if err := e.db.BatchCreateLabelRelations(labelRelations); err != nil {
			return apierrors.ErrUpdateIteration.InternalError(err).ToResp(), nil
		}
	}

	// 更新项目活跃时间
	if err := e.bdl.UpdateProjectActiveTime(apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  iteration.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp(iterationID)
}

// DeleteIteration 删除 iteration
func (e *Endpoints) DeleteIteration(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteIteration.NotLogin().ToResp(), nil
	}

	iterationID, err := strconv.ParseUint(vars["id"], 10, 64)
	onlyIteration, err := strconv.ParseBool(r.URL.Query().Get("onlyIteration"))
	if err != nil {
		return apierrors.ErrDeleteIteration.InvalidParameter("id").ToResp(), nil
	}
	iteration, err := e.iteration.Get(iterationID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteIteration.AccessDenied().ToResp(), nil
		}
	}

	// 删除 iteration
	if err := e.iteration.Delete(iterationID); err != nil {
		return errorresp.ErrResp(err)
	}

	itemIDs, err := e.issueService.DBClient().GetIssuesIDByIterationID(iterationID)
	if err != nil {
		return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
	}

	if len(itemIDs) > 0 {
		ctx = apis.WithUserIDContext(ctx, identityInfo.UserID)
		if onlyIteration {
			err = e.issueService.BatchUpdateIssueIterationIDByIterationID(ctx, iterationID, -1)
			if err != nil {
				return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
			}
		} else {
			err = e.issueService.BatchDeleteIssueByIterationID(ctx, iterationID)
			if err != nil {
				return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
			}
		}
	}
	// delete relation labels
	if err = e.db.DeleteLabelRelations(apistructs.LabelTypeIteration, strconv.FormatUint(iterationID, 10), nil); err != nil {
		return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
	}

	// 更新项目活跃时间
	if err := e.bdl.UpdateProjectActiveTime(apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  iteration.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp(iteration)
}

// GetIteration 获取迭代详情
func (e *Endpoints) GetIteration(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIteration.NotLogin().ToResp(), nil
	}

	// unassigned iterationID is virtual data
	if vars["id"] == strconv.Itoa(apistructs.UnassignedIterationID) {
		return httpserver.OkResp(nil)
	}

	iterationID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrGetIteration.InvalidParameter("id").ToResp(), nil
	}
	iteration, err := e.iteration.Get(iterationID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrGetIteration.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrGetIteration.AccessDenied().ToResp(), nil
		}
	}
	issueSummary, err := e.iteration.GetIterationSummary(iteration.ProjectID, iteration.ID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	convertIteration := iteration.Convert(e.getIterationLabelDetails(iterationID))
	convertIteration.IssueSummary = *issueSummary

	userIDs := []string{iteration.Creator}
	return httpserver.OkResp(convertIteration, userIDs)
}

// PagingIterations 分页查询迭代
func (e *Endpoints) PagingIterations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var pageReq apistructs.IterationPagingRequest
	if err := e.queryStringDecoder.Decode(&pageReq, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIterations.InvalidParameter(err).ToResp(), nil
	}

	orgIDStr := r.Header.Get("Org-ID")
	if orgIDStr == "" {
		return nil, errors.Errorf("invalid request, orgId is empty")
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return nil, errors.Errorf("invalid request, orgId is invalid")
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingIterations.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		if pageReq.ProjectID == 0 {
			return apierrors.ErrPagingIterations.InvalidParameter("missing projectID").ToResp(), nil
		}
		per := &apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  pageReq.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.ListAction,
		}
		access, err := e.bdl.CheckPermission(per)
		if err != nil {
			return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			// 再次校验企业的权限
			per.Scope = apistructs.OrgScope
			per.ScopeID = orgID
			access, err := e.bdl.CheckPermission(per)
			if err != nil {
				return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
			}
			if !access.Access {
				return apierrors.ErrPagingIterations.AccessDenied().ToResp(), nil
			}
		}
	}
	if len(pageReq.LabelIDs) > 0 {
		lrs, err := e.db.GetLabelRelationsByLabels(apistructs.LabelTypeIteration, pageReq.LabelIDs)
		if err != nil {
			return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
		}
		if pageReq.IDs == nil {
			pageReq.IDs = make([]uint64, 0)
		}
		for _, v := range lrs {
			id, err := strconv.ParseUint(v.RefID, 10, 64)
			if err != nil {
				logrus.Errorf("failed to parse refID for label relation %d, %v", v.ID, err)
				continue
			}
			pageReq.IDs = append(pageReq.IDs, id)
		}
	}
	// 分页查询
	iterationModels, total, err := e.iteration.Paging(pageReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	iterationMap := make(map[int64]*apistructs.Iteration, len(iterationModels))
	for _, itr := range iterationModels {
		iteration := itr.Convert(e.getIterationLabelDetails(itr.ID))
		iterationMap[iteration.ID] = &iteration
	}
	if !pageReq.WithoutIssueSummary {
		if err := e.iteration.SetIssueSummaries(pageReq.ProjectID, iterationMap); err != nil {
			return errorresp.ErrResp(err)
		}
	}
	iterations := make([]apistructs.Iteration, 0, len(iterationModels))
	for _, itrModel := range iterationModels {
		if itr, existed := iterationMap[int64(itrModel.ID)]; existed {
			iterations = append(iterations, *itr)
		}
	}
	// userIDs
	var userIDs []string
	for _, itr := range iterations {
		userIDs = append(userIDs, itr.Creator)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	// 返回
	return httpserver.OkResp(apistructs.IterationPagingResponseData{
		Total: total,
		List:  iterations,
	}, userIDs)
}

func (e *Endpoints) getIterationLabelDetails(iterationID uint64) ([]string, []apistructs.ProjectLabel) {
	lrs, _ := e.db.GetLabelRelationsByRef(apistructs.LabelTypeIteration, strconv.FormatUint(iterationID, 10))
	labelIDs := make([]uint64, 0, len(lrs))
	for _, v := range lrs {
		labelIDs = append(labelIDs, v.LabelID)
	}
	var labelNames []string
	var labels []apistructs.ProjectLabel
	labels, _ = e.bdl.ListLabelByIDs(labelIDs)
	labelNames = make([]string, 0, len(labels))
	for _, v := range labels {
		labelNames = append(labelNames, v.Name)
	}
	return labelNames, labels
}
