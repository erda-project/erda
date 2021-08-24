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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

// CreateIssueState 创建事件状态
func (e *Endpoints) CreateIssueState(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var Req apistructs.IssueStateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&Req); err != nil {
		return apierrors.ErrCreateIssueState.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueState.NotLogin().ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(Req.ProjectID),
			Resource: apistructs.IssueStateResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateIssueState.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateIssueState.AccessDenied().ToResp(), nil
		}
	}
	state, err := e.issueState.CreateIssueState(&Req)
	if err != nil {
		return apierrors.ErrCreateIssueState.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(state.ID)
}

// DeleteIssueState 删除事件状态
func (e *Endpoints) DeleteIssueState(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var delReq apistructs.IssueStateDeleteRequest
	if err := e.queryStringDecoder.Decode(&delReq, r.URL.Query()); err != nil {
		return apierrors.ErrDeleteIssueState.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteIssueState.NotLogin().ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(delReq.ProjectID),
			Resource: apistructs.IssueStateResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return apierrors.ErrDeleteIssueState.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteIssueState.AccessDenied().ToResp(), nil
		}
	}
	state, err := e.issueState.GetIssuesStatesByID(delReq.ID)
	if err != nil {
		return apierrors.ErrDeleteIssueState.InternalError(err).ToResp(), nil
	}
	err = e.issueState.DeleteIssueState(delReq.ID)
	if err != nil {
		return apierrors.ErrDeleteIssueState.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(state)
}

// UpdateIssueStateRelation 更新工作流
func (e *Endpoints) UpdateIssueStateRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssueStateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrUpdateIssueStateRelation.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssueStateRelation.NotLogin().ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(updateReq.ProjectID),
			Resource: apistructs.IssueStateResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateIssueState.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateIssueState.AccessDenied().ToResp(), nil
		}
	}

	issueType, err := e.issueState.UpdateIssueStates(&updateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// 工作流详情
	issueStateRelations, err := e.issueState.GetIssueStatesRelations(apistructs.IssueStateRelationGetRequest{ProjectID: uint64(updateReq.ProjectID), IssueType: issueType})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issueStateRelations)
}

// GetIssueStateRelation 获取工作流
func (e *Endpoints) GetIssueStateRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var getReq apistructs.IssueStateRelationGetRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssues.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssueStateRelation.NotLogin().ToResp(), nil
	}

	// 工作流详情
	issueStateRelations, err := e.issueState.GetIssueStatesRelations(getReq)
	if err != nil {
		return apierrors.ErrGetIssueStateRelation.InternalError(err).ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}
	return httpserver.OkResp(issueStateRelations)
}

// GetIssueStates 获取事件状态列表
func (e *Endpoints) GetIssueStates(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var getReq apistructs.IssueStatesGetRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueState.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssueState.NotLogin().ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		// TODO
	}

	state, err := e.issueState.GetIssueStates(&getReq)
	if err != nil {
		return apierrors.ErrGetIssueState.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(state)
}

// GetIssueStatesBelong 根据事件类型获取事件状态列表
func (e *Endpoints) GetIssueStatesBelong(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var getReq apistructs.IssueStateRelationGetRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueState.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssueState.NotLogin().ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		// TODO
	}

	state, err := e.issueState.GetIssueStatesBelong(&getReq)
	if err != nil {
		return apierrors.ErrGetIssueState.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(state)
}

func (e *Endpoints) GetIssueStatesByIDs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var getReq []int64
	if err := json.NewDecoder(r.Body).Decode(&getReq); err != nil {
		return apierrors.ErrGetIssueState.InvalidParameter(err).ToResp(), nil
	}
	states, err := e.issueState.GetIssuesStatesNameByID(getReq)
	if err != nil {
		return apierrors.ErrGetIssueState.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(states)
}
