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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) CreateIssuePropertyInstance(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IssuePropertyRelationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 事件详情
	issueModel, err := e.DBClient().GetIssue(createReq.IssueID)
	if err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueProperty.NotLogin().ToResp(), nil
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
			ScopeID:  issueModel.ProjectID,
			Resource: issueModel.Type.GetCorrespondingResource(),
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateIssue.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrCreateIssue.AccessDenied().ToResp(), nil
		}
	}
	if err := e.issueProperty.CreatePropertyRelation(&createReq); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(createReq.IssueID)
}

func (e *Endpoints) UpdateIssuePropertyInstance(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var updateReq apistructs.IssuePropertyRelationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrCreateIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 事件详情
	issueModel, err := e.DBClient().GetIssue(updateReq.IssueID)
	if err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueProperty.NotLogin().ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo
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
	if err := e.issueProperty.UpdatePropertyRelation(&updateReq); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp("update success")
}

func (e *Endpoints) GetIssuePropertyInstance(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var getReq apistructs.IssuePropertyRelationGetRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 事件详情
	issueModel, err := e.DBClient().GetIssue(getReq.IssueID)
	if err != nil {
		return apierrors.ErrUpdateIssue.InvalidParameter(err).ToResp(), nil
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueProperty.NotLogin().ToResp(), nil
	}
	getReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// issue 分页查询 校验用户在 当前 project 下是否拥有 GET ${project} 权限
		if issueModel.ProjectID == 0 {
			return apierrors.ErrPagingIssues.MissingParameter("projectID").ToResp(), nil
		}
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  issueModel.ProjectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrPagingIssues.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrPagingIssues.AccessDenied().ToResp(), nil
		}
	}
	instance, err := e.issueProperty.GetPropertyRelation(&getReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(instance)
}
