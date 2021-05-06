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
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateIteration 创建迭代
func (e *Endpoints) CreateIteration(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IterationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
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
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  createReq.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateIteration.InternalError(err).ToResp(), nil
		}
		if !access {
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

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
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
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateIteration.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrUpdateIteration.AccessDenied().ToResp(), nil
		}
	}
	// 更新 iteration
	if err := e.iteration.Update(iterationID, updateReq); err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
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
	if err != nil {
		return apierrors.ErrDeleteIteration.InvalidParameter("id").ToResp(), nil
	}
	iteration, err := e.iteration.Get(iterationID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return apierrors.ErrDeleteIteration.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrDeleteIteration.AccessDenied().ToResp(), nil
		}
	}

	// 删除 iteration
	if err := e.iteration.Delete(iterationID); err != nil {
		return errorresp.ErrResp(err)
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
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
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  iteration.ProjectID,
			Resource: apistructs.IterationResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrGetIteration.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrGetIteration.AccessDenied().ToResp(), nil
		}
	}

	userIDs := []string{iteration.Creator}
	return httpserver.OkResp(iteration.Convert(), userIDs)
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
		access, err := e.permission.CheckPermission(per)
		if err != nil {
			return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
		}
		if !access {
			// 再次校验企业的权限
			per.Scope = apistructs.OrgScope
			per.ScopeID = orgID
			access, err := e.permission.CheckPermission(per)
			if err != nil {
				return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
			}
			if !access {
				return apierrors.ErrPagingIterations.AccessDenied().ToResp(), nil
			}
		}
	}
	// 分页查询
	iterationModels, total, err := e.iteration.Paging(pageReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	iterations := make([]apistructs.Iteration, 0, len(iterationModels))
	for _, itr := range iterationModels {
		iteration := itr.Convert()
		// 默认不查询 summary
		// TODO 这里面的逻辑需要优化，调用了太多次数据库 @周子曰
		if !pageReq.WithoutIssueSummary {
			iteration.IssueSummary, err = e.iteration.GetIssueSummary(iteration.ID, pageReq.ProjectID)
			if err != nil {
				return apierrors.ErrPagingIterations.InternalError(err).ToResp(), nil
			}
		}
		iterations = append(iterations, iteration)
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
