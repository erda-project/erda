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

// CreateIssuePanel create panel
func (e *Endpoints) CreateIssuePanel(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuePanelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	permissionReq := apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  req.ProjectID,
		Resource: apistructs.IssuePanelResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.bdl.CheckPermission(&permissionReq); err != nil || !access.Access {
		return apierrors.ErrCreateIssuePanel.AccessDenied().ToResp(), nil
	}

	req.IdentityInfo = identityInfo
	panel, err := e.issuePanel.CreatePanel(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(panel.ID)
}

// DeleteIssuePanel delete panel
func (e *Endpoints) DeleteIssuePanel(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var delReq apistructs.IssuePanelRequest
	if err := e.queryStringDecoder.Decode(&delReq, r.URL.Query()); err != nil {
		return apierrors.ErrDeleteIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  delReq.ProjectID,
		Resource: apistructs.IssuePanelResource,
		Action:   apistructs.DeleteAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrDeleteIssuePanel.AccessDenied().ToResp(), nil
	}

	panel, err := e.issuePanel.DeletePanel(&delReq)
	if err != nil {
		return apierrors.ErrDeleteIssuePanel.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(panel)
}

// UpdateIssuePanel update panel
func (e *Endpoints) UpdateIssuePanel(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssuePanelRequest
	if err := e.queryStringDecoder.Decode(&updateReq, r.URL.Query()); err != nil {
		return apierrors.ErrUpdateIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  updateReq.ProjectID,
		Resource: apistructs.IssuePanelResource,
		Action:   apistructs.UpdateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrUpdateIssuePanel.AccessDenied().ToResp(), nil
	}

	panel, err := e.issuePanel.UpdatePanel(&updateReq)
	if err != nil {
		return apierrors.ErrDeleteIssuePanel.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(panel.ID)
}

// UpdateIssuePanelIssue update panelIssue
func (e *Endpoints) UpdateIssuePanelIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var updateReq apistructs.IssuePanelRequest
	if err := e.queryStringDecoder.Decode(&updateReq, r.URL.Query()); err != nil {
		return apierrors.ErrUpdateIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  updateReq.ProjectID,
		Resource: apistructs.IssuePanelResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrUpdateIssuePanel.AccessDenied().ToResp(), nil
	}

	panel, err := e.issuePanel.UpdatePanelIssue(&updateReq)
	if err != nil {
		return apierrors.ErrUpdateIssuePanel.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(panel.ID)
}

// GetIssuePanel get panel
func (e *Endpoints) GetIssuePanel(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var getReq apistructs.IssuePanelRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
	}

	panel, err := e.issuePanel.GetPanelByProjectID(&getReq)
	if err != nil {
		return apierrors.ErrGetIssuePanel.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(panel)
}

// GetIssuePanelIssue get panelIssue
func (e *Endpoints) GetIssuePanelIssue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var getReq apistructs.IssuePanelRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssuePanel.InvalidParameter(err).ToResp(), nil
	}
	// 用户身份
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssuePanel.NotLogin().ToResp(), nil
	}
	// 鉴权
	if !identityInfo.IsInternalClient() {
	}

	var panel apistructs.IssuePanelIssueIDs
	issues, total, err := e.issuePanel.GetPanelIssues(&getReq)
	if err != nil {
		return apierrors.ErrGetIssuePanel.InternalError(err).ToResp(), nil
	}
	panel.Issues = issues
	panel.Total = total

	return httpserver.OkResp(panel)
}
