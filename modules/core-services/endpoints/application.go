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
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// Private 私有
	Private = "private"
	// Public 公开
	Public = "public"
	// All 全部
	All = ""
)

// CreateApplication 创建应用
func (e *Endpoints) CreateApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	identify, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter(err).ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateApplication.MissingParameter("body is nil").ToResp(), nil
	}

	var applicationCreateReq apistructs.ApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&applicationCreateReq); err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter("can't decode body").ToResp(), nil
	}
	if !strutil.IsValidPrjOrAppName(applicationCreateReq.Name) {
		return apierrors.ErrCreateApplication.InvalidParameter(errors.Errorf("app name is invalid %s",
			applicationCreateReq.Name)).ToResp(), nil
	}
	logrus.Infof("request body: %+v", applicationCreateReq)

	// 参数合法性检查
	if err := checkApplicationCreateParam(applicationCreateReq); err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   identify.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  applicationCreateReq.ProjectID,
		Resource: apistructs.AppResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrCreateApplication.AccessDenied().ToResp(), nil
	}

	// 添加应用至DB
	app, err := e.app.CreateWithEvent(identify.UserID, &applicationCreateReq)
	if err != nil {
		return apierrors.ErrCreateApplication.InternalError(err).ToResp(), nil
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  applicationCreateReq.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	applicationDTO := e.convertToApplicationDTO(*app, false, identify.UserID, nil)

	return httpserver.OkResp(applicationDTO)
}

// InitApplication 应用初始化
func (e *Endpoints) InitApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrInitApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strconv.ParseUint(vars["applicationID"], 10, 64)
	if err != nil {
		return apierrors.ErrInitApplication.InvalidParameter(err).ToResp(), nil
	}

	// 检查请求body
	var appInitReq apistructs.ApplicationInitRequest
	if err := json.NewDecoder(r.Body).Decode(&appInitReq); err != nil {
		return apierrors.ErrInitApplication.InvalidParameter(err).ToResp(), nil
	}
	appInitReq.ApplicationID = applicationID
	appInitReq.IdentityInfo = identityInfo

	if !identityInfo.IsInternalClient() {
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.AppScope,
			ScopeID:  applicationID,
			Resource: apistructs.AppResource,
			Action:   apistructs.CreateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrInitApplication.AccessDenied().ToResp(), nil
		}
	}

	// 更新应用信息至DB
	pipelineID, err := e.app.Init(&appInitReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pipelineID)
}

// UpdateApplication 更新应用
func (e *Endpoints) UpdateApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrUpdateApplication.InvalidParameter(err).ToResp(), nil
	}

	// 检查请求body
	if r.Body == nil {
		return apierrors.ErrUpdateApplication.MissingParameter("body").ToResp(), nil
	}
	var applicationUpdateReq apistructs.ApplicationUpdateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&applicationUpdateReq); err != nil {
		return apierrors.ErrUpdateApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(applicationID),
		Resource: apistructs.AppResource,
		Action:   apistructs.UpdateAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrUpdateApplication.AccessDenied().ToResp(), nil
	}

	application, err := e.app.Get(applicationID)
	if err != nil {
		return apierrors.ErrUpdateApplication.InternalError(err).ToResp(), nil
	}
	if application.IsPublic != applicationUpdateReq.IsPublic {
		// 公开应用操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.AppScope,
			ScopeID:  uint64(applicationID),
			Resource: apistructs.AppPublicResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrUpdateApplication.AccessDenied().ToResp(), nil
		}
	}
	// 更新应用信息至DB
	app, err := e.app.UpdateWithEvent(applicationID, &applicationUpdateReq)
	if err != nil {
		return apierrors.ErrUpdateApplication.InternalError(err).ToResp(), nil
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  uint64(app.ProjectID),
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp("update succ")
}

// GetApplication 获取应用详情
func (e *Endpoints) GetApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrGetApplication.InvalidParameter(err).ToResp(), nil
	}

	var userID string
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		v, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetApplication.NotLogin().ToResp(), nil
		}
		userID = v.String()
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID,
			Scope:    apistructs.AppScope,
			ScopeID:  uint64(applicationID),
			Resource: apistructs.AppResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrGetApplication.AccessDenied().ToResp(), nil
		}
	}

	app, err := e.app.Get(applicationID)
	if err != nil {
		return apierrors.ErrGetApplication.InternalError(err).ToResp(), nil
	}

	orgid_s := r.Header.Get("Org-ID")
	orgID, err := strconv.ParseInt(orgid_s, 10, 64)
	if err != nil {
		orgID = 0
	}
	approves, err := e.db.ListUnblockApplicationApprove(uint64(orgID))
	if err != nil {
		return apierrors.ErrListApplication.InternalError(err).ToResp(), nil
	}
	blockStatusMap := map[uint64]string{}
	for _, approve := range approves {
		extramap := map[string]string{}
		if err := json.Unmarshal([]byte(approve.Extra), &extramap); err != nil {
			continue
		}
		appidsStr := extramap["appIDs"]
		appids := strutil.Split(appidsStr, ",", true)
		status := ""
		if approve.Status == string(apistructs.ApprovalStatusPending) {
			status = "unblocking"
		}
		for _, id := range appids {
			id, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				continue
			}
			blockStatusMap[id] = status
		}
	}
	return httpserver.OkResp(e.convertToApplicationDTO(*app, true, userID, blockStatusMap), nil)
}

// DeleteApplication 删除应用
func (e *Endpoints) DeleteApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeleteApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrDeleteApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(applicationID),
		Resource: apistructs.AppResource,
		Action:   apistructs.DeleteAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrDeleteApplication.AccessDenied().ToResp(), nil
	}
	app, err := e.app.Get(applicationID)
	applicationDTO := e.convertToApplicationDTO(*app, false, userID.String(), nil)

	// 删除应用
	if err = e.app.DeleteWithEvent(applicationID); err != nil {
		return apierrors.ErrDeleteApplication.InternalError(err).ToResp(), nil
	}

	// 更新项目活跃时间
	if err := e.project.UpdateProjectActiveTime(&apistructs.ProjectActiveTimeUpdateRequest{
		ProjectID:  applicationDTO.ProjectID,
		ActiveTime: time.Now(),
	}); err != nil {
		logrus.Errorf("update project active time err: %v", err)
	}

	return httpserver.OkResp(applicationDTO)
}

// ListApplication 所有应用列表
func (e *Endpoints) ListApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return e.listApplications(ctx, r, false)
}

// ListMyApplication 我的应用列表
func (e *Endpoints) ListMyApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return e.listApplications(ctx, r, true)
}

// GetApplicationPublishItemRelationsGroupByENV 根据环境分组应用和发布内容关联
func (e *Endpoints) GetApplicationPublishItemRelationsGroupByENV(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrGetApplicationPublishItemRelation.InvalidParameter(err).ToResp(), nil
	}

	relations, err := e.app.QueryPublishItemRelations(apistructs.QueryAppPublishItemRelationRequest{AppID: applicationID})
	if err != nil {
		return apierrors.ErrGetApplicationPublishItemRelation.InternalError(err).ToResp(), nil
	}

	result := map[string]apistructs.AppPublishItemRelation{}
	for _, relation := range relations {
		var itemNs []string
		itemNs = append(itemNs, e.app.BuildItemMonitorPipelineCmsNs(relation.AppID, relation.Env))
		relation.PublishItemNs = itemNs
		result[relation.Env] = relation
	}

	return httpserver.OkResp(result)
}

// QueryApplicationPublishItemRelations 查询应用和发布内容关联
func (e *Endpoints) QueryApplicationPublishItemRelations(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var req apistructs.QueryAppPublishItemRelationRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssues.InvalidParameter(err).ToResp(), nil
	}

	relations, err := e.app.QueryPublishItemRelations(req)
	if err != nil {
		return apierrors.ErrGetApplicationPublishItemRelation.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(relations)
}

// UpdateApplicationPublishItemRelations 更新应用发布内容关联
func (e *Endpoints) UpdateApplicationPublishItemRelations(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrUpdateApplicationPublishItemRelation.InvalidParameter(err).ToResp(), nil
	}

	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateApplicationPublishItemRelation.InvalidParameter(err).ToResp(), nil
	}

	var request apistructs.UpdateAppPublishItemRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrUpdateApplicationPublishItemRelation.InvalidParameter("can't decode body").ToResp(), nil
	}
	request.AppID = applicationID
	request.UserID = userID.String()
	request.AKAIMap = make(map[apistructs.DiceWorkspace]apistructs.MonitorKeys, 0)
	err = e.app.UpdatePublishItemRelations(&request)
	if err != nil {
		return apierrors.ErrUpdateApplicationPublishItemRelation.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}

// RemoveApplicationPublishItemRelations 删除应用发布内容关联
func (e *Endpoints) RemoveApplicationPublishItemRelations(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var request apistructs.RemoveAppPublishItemRelationsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrRemoveApplicationPublishItemRelation.InvalidParameter("can't decode body").ToResp(), nil
	}
	err := e.app.RemovePublishItemRelations(&request)
	if err != nil {
		return apierrors.ErrRemoveApplicationPublishItemRelation.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}

func (e *Endpoints) listApplications(ctx context.Context, r *http.Request, isMine bool) (httpserver.Responser, error) {
	// 获取当前用户
	var (
		total        int
		applications []model.Application
		err          error
	)

	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListApplication.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrListApplication.MissingParameter("org id").ToResp(), nil
	}
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrListApplication.InvalidParameter(err).ToResp(), nil
	}

	// 获取请求参数
	params, err := getListApplicationsParam(r)
	if err != nil {
		return apierrors.ErrListApplication.InvalidParameter(err).ToResp(), nil
	}

	if isMine {
		total, applications, err = e.app.ListMyApplications(orgID, userID.String(), params)
	} else {
		// 获取当前用户
		identityInfo, err := user.GetIdentityInfo(r)
		if err != nil {
			return apierrors.ErrInitApplication.NotLogin().ToResp(), nil
		}
		if !identityInfo.IsInternalClient() {
			// 操作鉴权
			req := apistructs.PermissionCheckRequest{
				UserID:   identityInfo.UserID,
				Scope:    apistructs.ProjectScope,
				ScopeID:  params.ProjectID,
				Resource: apistructs.ProjectResource,
				Action:   apistructs.GetAction,
			}
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrInitApplication.AccessDenied().ToResp(), nil
			}
		}
		total, applications, err = e.app.List(orgID, int64(params.ProjectID), userID.String(), params)
	}

	if err != nil {
		return apierrors.ErrListApplication.InternalError(err).ToResp(), nil
	}

	// 获取项目下成员对应应用的角色
	memberMap := make(map[int64][]string, 0)

	memberID := r.URL.Query().Get("memberID")
	if memberID != "" {
		// 查找成员项目下有权限的列表
		members, err := e.db.GetMembersByParentID(apistructs.AppScope, int64(params.ProjectID), memberID)
		if err != nil {
			return nil, errors.Errorf("failed to get member list under project, (%v)", err)
		}

		for _, m := range members {
			if m.ResourceKey == apistructs.RoleResourceKey {
				memberMap[m.ScopeID] = append(memberMap[m.ScopeID], m.ResourceValue)
			}
		}
	}

	approves, err := e.db.ListUnblockApplicationApprove(uint64(orgID))
	if err != nil {
		return apierrors.ErrListApplication.InternalError(err).ToResp(), nil
	}
	blockStatusMap := map[uint64]string{}
	for _, approve := range approves {
		extramap := map[string]string{}
		if err := json.Unmarshal([]byte(approve.Extra), &extramap); err != nil {
			continue
		}
		appids_str := extramap["appIDs"]
		appids := strutil.Split(appids_str, ",", true)
		status := ""
		if approve.Status == string(apistructs.ApprovalStatusPending) {
			status = "unblocking"
		}
		for _, id_s := range appids {
			id, err := strconv.ParseUint(id_s, 10, 64)
			if err != nil {
				continue
			}
			blockStatusMap[id] = status
		}
	}

	// 转换成所需格式
	pinedAppDTOs := make([]apistructs.ApplicationDTO, 0, 10)
	unpinedAppDTOs := make([]apistructs.ApplicationDTO, 0, len(applications))

	projectIDs := make([]uint64, 0, len(applications))
	for _, app := range applications {
		projectIDs = append(projectIDs, uint64(app.ProjectID))
	}

	projectMap, err := e.project.GetModelProjectsMap(projectIDs)
	if err != nil {
		return apierrors.ErrListApplication.InternalError(err).ToResp(), nil
	}

	for i := range applications {
		var projectName string
		var projectDisplayName string
		if project, ok := projectMap[applications[i].ProjectID]; ok {
			projectName = project.Name
			projectDisplayName = project.DisplayName
		}

		if params.IsSimple {
			appDTO := apistructs.ApplicationDTO{
				Pined:              applications[i].Pined,
				Name:               applications[i].Name,
				Desc:               applications[i].Desc,
				Creator:            applications[i].UserID,
				CreatedAt:          applications[i].CreatedAt,
				UpdatedAt:          applications[i].UpdatedAt,
				ID:                 uint64(applications[i].ID),
				DisplayName:        applications[i].DisplayName,
				IsPublic:           applications[i].IsPublic,
				ProjectID:          uint64(applications[i].ProjectID),
				ProjectName:        projectName,
				ProjectDisplayName: projectDisplayName,
			}
			if appDTO.Pined {
				pinedAppDTOs = append(pinedAppDTOs, appDTO)
			} else {
				unpinedAppDTOs = append(unpinedAppDTOs, appDTO)
			}
			continue
		}
		appDTO := e.convertToApplicationDTO(applications[i], false, userID.String(), blockStatusMap)
		// 填充成员角色
		if roles, ok := memberMap[int64(appDTO.ID)]; ok {
			appDTO.MemberRoles = roles
		}

		appDTO.ProjectName = projectName
		appDTO.ProjectDisplayName = projectDisplayName
		if appDTO.Pined {
			pinedAppDTOs = append(pinedAppDTOs, appDTO)
		} else {
			unpinedAppDTOs = append(unpinedAppDTOs, appDTO)
		}
	}
	applicationDTOs := make([]apistructs.ApplicationDTO, 0, len(applications))
	applicationDTOs = append(pinedAppDTOs, unpinedAppDTOs...)

	return httpserver.OkResp(apistructs.ApplicationListResponseData{Total: total, List: applicationDTOs})
}

// PinApplication  pin应用
func (e Endpoints) PinApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrPinApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrPinApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(applicationID),
		Resource: apistructs.AppResource,
		Action:   apistructs.GetAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrPinApplication.AccessDenied().ToResp(), nil
	}

	if err := e.app.Pin(applicationID, userID.String()); err != nil {
		return apierrors.ErrPinApplication.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("pin succ")
}

// UnPinApplication  unpin应用
func (e Endpoints) UnPinApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUnPinApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrUnPinApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(applicationID),
		Resource: apistructs.AppResource,
		Action:   apistructs.GetAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrUnPinApplication.AccessDenied().ToResp(), nil
	}

	if err := e.app.UnPin(applicationID, userID.String()); err != nil {
		return apierrors.ErrUnPinApplication.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("unpin succ")
}

// ListAppTemplates 列出应用模板
func (e Endpoints) ListAppTemplates(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	appMode := r.URL.Query().Get("mode")

	var templates []string
	switch appMode {
	case string(apistructs.ApplicationModeMobile):
		templates = append(templates, "terminusMobileTemplates")
	case "":
		return apierrors.ErrListAppTemplates.MissingParameter("mode").ToResp(), nil
	}

	return httpserver.OkResp(templates)
}

func checkApplicationCreateParam(applicationCreateReq apistructs.ApplicationCreateRequest) error {
	if applicationCreateReq.Name == "" {
		return errors.Errorf("invalid request, name is empty")
	}
	if applicationCreateReq.ProjectID == 0 {
		return errors.Errorf("invalid request, projectId is empty")
	}
	err := applicationCreateReq.Mode.CheckAppMode()
	return err
}

// 应用列表时获取请求参数
func getListApplicationsParam(r *http.Request) (*apistructs.ApplicationListRequest, error) {
	mode := r.URL.Query().Get("mode")
	if mode != "" {
		err := apistructs.ApplicationMode(mode).CheckAppMode()
		if err != nil {
			return nil, err
		}
	}

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid request, pageSize is invalid")
	}

	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid request, pageNo is invalid")
	}

	// 获取项目Id
	projectIDStr := r.URL.Query().Get("projectId")
	var projectID int64
	if projectIDStr != "" {
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return nil, errors.Errorf("invalid request, projectId is invalid")
		}
	}

	// 按应用名称搜索
	keyword := r.URL.Query().Get("q")

	public := r.URL.Query().Get("public")
	if public != All && public != Public && public != Private {
		return nil, errors.Errorf("invalid request, public is invalid")
	}

	// 是否只返回简单信息
	isSimple := false
	if r.URL.Query().Get("isSimple") == "true" {
		isSimple = true
	}

	orderBy := r.URL.Query().Get("orderBy")

	req := &apistructs.ApplicationListRequest{
		ProjectID: uint64(projectID),
		Mode:      mode,
		Query:     keyword,
		Name:      r.URL.Query().Get("name"),
		PageNo:    pageNo,
		PageSize:  pageSize,
		Public:    public,
		IsSimple:  isSimple,
		OrderBy:   orderBy,
	}

	return req, nil
}

func (e *Endpoints) checkPermission(r *http.Request, scopeType apistructs.ScopeType, scopeID int64) error {
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient != "" { // 内部服务调用
		return nil
	}

	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return err
	}

	// 操作鉴权
	if e.member.IsAdmin(userID.String()) { // 系统管理员
		return nil
	}
	if permission, err := e.getPermission(userID.String(), scopeType, scopeID); err != nil ||
		!permission.Access {
		return errors.Errorf("failed to get permission")
	}

	return nil
}

func (e *Endpoints) convertToApplicationDTO(application model.Application, withProjectToken bool, userID string, blockStatusMap map[uint64]string) apistructs.ApplicationDTO {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(application.Config), &config); err != nil {
		config = make(map[string]interface{})
	}

	var extra map[string]string
	if err := json.Unmarshal([]byte(application.Extra), &extra); err != nil {
		extra = make(map[string]string)
	}

	token := ""
	if withProjectToken {
		members, err := e.db.GetMemberByScopeAndUserID(userID, apistructs.OrgScope, application.OrgID)
		if err == nil && members != nil && len(members) > 0 {
			token = members[0].Token
		}
	}

	// TODO ApplicationDTO 去除clusterName, 暂时兼容添加
	project, err := e.project.Get(application.ProjectID)
	if err != nil {
		logrus.Error(err)
	}

	workspaces := make([]apistructs.ApplicationWorkspace, 0, len(extra))
	for k, v := range extra {
		if strings.Contains(k, ".") {
			env := strings.Split(k, ".")[0]
			workspace := apistructs.ApplicationWorkspace{
				Workspace:       env,
				ConfigNamespace: v,
			}
			if project != nil {
				workspace.ClusterName = project.ClusterConfig[env]
			}
			workspaces = append(workspaces, workspace)
		}
	}

	// TODO ApplicationDTO 去除orgName，暂时兼容添加
	var orgName string
	var orgDisplayName string
	org, err := e.org.Get(application.OrgID)
	if err == nil {
		orgName = org.Name
		orgDisplayName = org.DisplayName
	} else {
		logrus.Error(err)
	}

	// TODO 应用表新增reference字段，提供API供orchestrator调用
	var runtimeCount uint
	e.db.Table("ps_v2_project_runtimes").Where("application_id = ?", application.ID).Count(&runtimeCount)

	gitRepo := strutil.Concat(conf.GittarOutterURL(), "/", application.GitRepoAbbrev)

	gitRepoNew := strutil.Concat(conf.UIDomain(), "/", orgName, "/dop/", application.ProjectName, "/", application.Name)

	var repoConfig apistructs.GitRepoConfig
	if application.IsExternalRepo {
		json.Unmarshal([]byte(application.RepoConfig), &repoConfig)
		repoConfig.Password = ""
		repoConfig.Username = ""
	}

	isOrgBlocked := false
	blockStatus := ""
	if org != nil && (org.BlockoutConfig.BlockDEV ||
		org.BlockoutConfig.BlockTEST ||
		org.BlockoutConfig.BlockStage ||
		org.BlockoutConfig.BlockProd) {
		isOrgBlocked = true
		blockStatus = "blocked"
	}
	now := time.Now()
	if application.UnblockStart != nil && application.UnblockEnd != nil &&
		now.Before(*application.UnblockEnd) && now.After(*application.UnblockStart) {
		blockStatus = "unblocked"
	} else if len(blockStatusMap) > 0 && blockStatusMap[uint64(application.ID)] != "" {
		blockStatus = blockStatusMap[uint64(application.ID)]
	}
	var unblockStart *time.Time
	var unblockEnd *time.Time

	if application.UnblockStart != nil && application.UnblockEnd != nil &&
		now.Before(*application.UnblockEnd) {
		// 过期之后的情况不返回时间
		unblockStart = application.UnblockStart
		unblockEnd = application.UnblockEnd
	}

	return apistructs.ApplicationDTO{
		ID:             uint64(application.ID),
		Name:           application.Name,
		DisplayName:    application.DisplayName,
		Desc:           application.Desc,
		Logo:           filehelper.APIFileUrlRetriever(application.Logo),
		Config:         config,
		UnBlockStart:   map[bool]*time.Time{true: unblockStart, false: nil}[isOrgBlocked],
		UnBlockEnd:     map[bool]*time.Time{true: unblockEnd, false: nil}[isOrgBlocked],
		BlockStatus:    map[bool]string{true: blockStatus, false: ""}[isOrgBlocked],
		ProjectID:      uint64(application.ProjectID),
		ProjectName:    application.ProjectName,
		OrgID:          uint64(application.OrgID),
		OrgName:        orgName,
		OrgDisplayName: orgDisplayName,
		Mode:           application.Mode,
		Pined:          application.Pined,
		IsPublic:       application.IsPublic,
		Creator:        application.UserID,
		GitRepo:        gitRepo,
		GitRepoAbbrev:  application.GitRepoAbbrev,
		GitRepoNew:     gitRepoNew,
		Token:          token,
		Workspaces:     workspaces,
		Stats: apistructs.ApplicationStats{
			CountRuntimes: runtimeCount,
		},
		IsExternalRepo: application.IsExternalRepo,
		RepoConfig:     &repoConfig,
		CreatedAt:      application.CreatedAt,
		UpdatedAt:      application.UpdatedAt,
		Extra:          application.Extra,
	}
}

// CountAppByProID count app by proID
func (e Endpoints) CountAppByProID(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCountApplication.NotLogin().ToResp(), nil
	}
	projectIDStr := r.URL.Query().Get("projectID")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrCountApplication.InvalidParameter(err).ToResp(), nil
	}
	count, err := e.db.GetApplicationCountByProjectID(projectID)
	if err != nil {
		return apierrors.ErrCountApplication.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(count)
}
