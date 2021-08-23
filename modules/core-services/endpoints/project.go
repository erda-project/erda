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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateProject 创建项目
func (e *Endpoints) CreateProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateProject.NotLogin().ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateProject.MissingParameter("body").ToResp(), nil
	}
	var projectCreateReq apistructs.ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&projectCreateReq); err != nil {
		return apierrors.ErrCreateProject.InvalidParameter(err).ToResp(), nil
	}
	if !strutil.IsValidPrjOrAppName(projectCreateReq.Name) {
		return apierrors.ErrCreateProject.InvalidParameter(errors.Errorf("project name is invalid %s",
			projectCreateReq.Name)).ToResp(), nil
	}
	logrus.Infof("request body: %+v", projectCreateReq)

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  projectCreateReq.OrgID,
		Resource: apistructs.ProjectResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrCreateProject.AccessDenied().ToResp(), nil
	}

	projectID, err := e.project.CreateWithEvent(userID.String(), &projectCreateReq)
	if err != nil {
		return apierrors.ErrCreateProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(projectID)
}

// UpdateProject 更新项目
func (e *Endpoints) UpdateProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateProject.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}

	// 检查projectID合法性
	projectID, err := strutil.Atoi64(vars["projectID"])
	if err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}

	// 检查request body合法性
	if r.Body == nil {
		return apierrors.ErrUpdateProject.MissingParameter("body").ToResp(), nil
	}
	var projectUpdateReq apistructs.ProjectUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&projectUpdateReq); err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", projectUpdateReq)

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.UpdateAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		// 若非项目管理员，判断用户是否为企业管理员(数据中心)
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.ProjectResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrUpdateProject.AccessDenied().ToResp(), nil
		}
	}

	oldProject, err := e.project.Get(projectID)
	if err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}
	if oldProject.IsPublic != projectUpdateReq.IsPublic {
		// 只有项目所有者可以更改项目public状态,二次鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(projectID),
			Resource: apistructs.ProjectPublicResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrUpdateProject.AccessDenied().ToResp(), nil
		}
	}

	// 更新项目信息至DB
	if err = e.project.UpdateWithEvent(projectID, &projectUpdateReq); err != nil {
		return apierrors.ErrUpdateProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(projectID)
}

// GetProject 获取项目详情
func (e *Endpoints) GetProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 检查projectID合法性
	projectID, err := strutil.Atoi64(vars["projectID"])
	if err != nil {
		return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetProject.NotLogin().ToResp(), nil
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(projectID),
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			// 若非项目管理员，判断用户是否为企业管理员(数据中心)
			orgIDStr := r.Header.Get(httputil.OrgHeader)
			orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
			if err != nil {
				return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
			}
			req := apistructs.PermissionCheckRequest{
				UserID:   userID.String(),
				Scope:    apistructs.OrgScope,
				ScopeID:  orgID,
				Resource: apistructs.ProjectResource,
				Action:   apistructs.GetAction,
			}
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrGetProject.AccessDenied().ToResp(), nil
			}
		}
	}

	project, err := e.project.Get(projectID)
	if err != nil {
		if err == dao.ErrNotFoundProject {
			return apierrors.ErrGetProject.NotFound().ToResp(), nil
		}
		return apierrors.ErrGetProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(*project, project.Owners)
}

// DeleteProject 删除项目
func (e *Endpoints) DeleteProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}

	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeleteProject.NotLogin().ToResp(), nil
	}

	// 检查projectID合法性
	projectID, err := strutil.Atoi64(vars["projectID"])
	if err != nil {
		return apierrors.ErrDeleteProject.InvalidParameter(err).ToResp(), nil
	}

	// 审计事件需要项目详情，发生错误不应中断业务流程
	project, err := e.project.Get(projectID)
	if err != nil {
		logrus.Errorf("when get project for audit faild %v", err)
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.DeleteAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		// 再鉴一次org下的权限
		req.Scope = apistructs.OrgScope
		req.ScopeID = uint64(orgID)
		if access, err = e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrDeleteProject.AccessDenied().ToResp(), nil
		}
	}

	// 删除项目
	if err = e.project.DeleteWithEvent(projectID); err != nil {
		return apierrors.ErrDeleteProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(project)
}

// ListProject 所有项目列表
func (e *Endpoints) ListProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListProject.NotLogin().ToResp(), nil
	}

	// 获取请求参数
	params, err := getListProjectsParam(r)
	if err != nil {
		return apierrors.ErrListProject.InvalidParameter(err).ToResp(), nil
	}

	// 企业管理员和 Support 都可以调用
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.OrgScope,
			ScopeID:  params.OrgID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.ListAction,
		}
		perm, err := e.permission.CheckPermission(&req)
		if err != nil {
			return apierrors.ErrListProject.InternalError(err).ToResp(), nil
		}
		if !perm {
			return apierrors.ErrListProject.AccessDenied().ToResp(), nil
		}
	}

	pagingProjects, err := e.project.ListAllProjects(userID.String(), params)
	if err != nil {
		return apierrors.ErrListProject.InternalError(err).ToResp(), nil
	}

	var userIDs []string
	for _, v := range pagingProjects.List {
		userIDs = append(userIDs, v.Owners...)
	}

	return httpserver.OkResp(*pagingProjects, userIDs)
}

// ListMyProject 我的项目列表
func (e *Endpoints) ListMyProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}

	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListProject.NotLogin().ToResp(), nil
	}

	// 获取请求参数
	params, err := getListProjectsParam(r)
	if err != nil {
		return apierrors.ErrListProject.InvalidParameter(err).ToResp(), nil
	}

	pagingProjects, err := e.project.ListJoinedProjects(orgID, userID.String(), params)
	if err != nil {
		return apierrors.ErrListProject.InternalError(err).ToResp(), nil
	}
	var userIDs []string
	for _, v := range pagingProjects.List {
		userIDs = append(userIDs, v.Owners...)
	}

	return httpserver.OkResp(*pagingProjects, userIDs)
}

// ListPublicProject 公开项目列表
func (e *Endpoints) ListPublicProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListProject.NotLogin().ToResp(), nil
	}

	// 获取请求参数
	params, err := getListProjectsParam(r)
	if err != nil {
		return apierrors.ErrListProject.InvalidParameter(err).ToResp(), nil
	}
	params.IsPublic = true

	pagingProjects, err := e.project.ListPublicProjects(userID.String(), params)
	if err != nil {
		return apierrors.ErrListProject.InternalError(err).ToResp(), nil
	}
	var userIDs []string
	for _, v := range pagingProjects.List {
		userIDs = append(userIDs, v.Owners...)
	}

	return httpserver.OkResp(*pagingProjects, userIDs)
}

// ReferCluster 查看集群是否被项目引用
func (e *Endpoints) ReferCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrReferProject.NotLogin().ToResp(), nil
	}
	// 仅内部使用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrReferProject.AccessDenied().ToResp(), nil
	}

	clusterName := r.URL.Query().Get("cluster")
	if clusterName == "" {
		return apierrors.ErrReferProject.MissingParameter("cluster").ToResp(), nil
	}
	reffered := e.project.ReferCluster(clusterName)

	return httpserver.OkResp(reffered)
}

// GetFunctions 获取项目功能开关配置
func (e *Endpoints) GetFunctions(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	projectIDStr := r.URL.Query().Get("projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetProject.NotLogin().ToResp(), nil
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(projectID),
			Resource: apistructs.ProjectFunctionResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			// 若非项目管理员，判断用户是否为企业管理员(数据中心)
			orgIDStr := r.Header.Get(httputil.OrgHeader)
			orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
			if err != nil {
				return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
			}
			req.Scope = apistructs.OrgScope
			req.ScopeID = orgID
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrGetProject.AccessDenied().ToResp(), nil
			}
		}
	}

	project, err := e.project.GetModelProject(projectID)
	if err != nil {
		if err == dao.ErrNotFoundProject {
			return apierrors.ErrGetProject.NotFound().ToResp(), nil
		}
		return apierrors.ErrGetProject.InternalError(err).ToResp(), nil
	}
	var pf map[apistructs.ProjectFunction]bool
	json.Unmarshal([]byte(project.Functions), &pf)

	return httpserver.OkResp(pf)
}

// SetFunctions 设置项目功能开关
func (e *Endpoints) SetFunctions(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateProject.NotLogin().ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrUpdateProject.MissingParameter("body").ToResp(), nil
	}
	var projectFuncReq apistructs.ProjectFunctionSetRequest
	if err := json.NewDecoder(r.Body).Decode(&projectFuncReq); err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", projectFuncReq)

	// 操作鉴权
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.ProjectScope,
			ScopeID:  projectFuncReq.ProjectID,
			Resource: apistructs.ProjectFunctionResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrUpdateProject.AccessDenied().ToResp(), nil
		}
	}

	projectID, err := e.project.UpdateProjectFunction(&projectFuncReq)
	if err != nil {
		return apierrors.ErrUpdateProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(projectID)
}

// UpdateProjectActiveTime 更新项目活跃时间
func (e *Endpoints) UpdateProjectActiveTime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser,
	error) {
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrUpdateProject.AccessDenied().ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrUpdateProject.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.ProjectActiveTimeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateProject.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	if err := e.project.UpdateProjectActiveTime(&req); err != nil {
		return apierrors.ErrUpdateProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("update project active time succ")
}

// 项目列表时获取请求参数
func getListProjectsParam(r *http.Request) (*apistructs.ProjectListRequest, error) {
	// 获取企业Id
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("orgId")
		if orgIDStr == "" {
			return nil, errors.Errorf("invalid param, orgId is empty")
		}
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	// 按项目名称搜索
	keyword := r.URL.Query().Get("q")

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageSize is invalid")
	}
	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageNo is invalid")
	}
	// 获取isPublic
	var isPublic bool
	isPublicStr := r.URL.Query().Get("is_public")
	if isPublicStr == "true" {
		isPublic = true
	}
	var asc bool
	ascStr := r.URL.Query().Get("asc")
	if ascStr == "true" {
		asc = true
	}
	orderBy := r.URL.Query().Get("orderBy")
	switch orderBy {
	case "cpuQuota":
		orderBy = "cpu_quota"
	case "memQuota":
		orderBy = "mem_quota"
	case "activeTime":
		orderBy = "active_time"
	case "name":
		orderBy = "name"
	default:
		orderBy = ""
	}

	return &apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		Query:    keyword,
		Name:     r.URL.Query().Get("name"),
		PageNo:   pageNo,
		PageSize: pageSize,
		OrderBy:  orderBy,
		Asc:      asc,
		IsPublic: isPublic,
	}, nil
}

// ListProjectResourceUsage 项目的 CPU/Memory 使用率的历史图表
func (e *Endpoints) ListProjectResourceUsage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListProjectResourceUsage.NotLogin().ToResp(), nil
	}

	// 获取请求的项目ID参数，用于鉴权
	projectID := r.URL.Query().Get("filter_project_id")
	if projectID == "" {
		return nil, apierrors.ErrListProjectResourceUsage.AccessDenied()
	}

	projectIDInt, err := strconv.Atoi(projectID)
	access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectIDInt),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, apierrors.ErrListProjectResourceUsage.AccessDenied()
	}

	metricsResp, err := e.bdl.GetProjectMetric(r.URL.Query())
	if err != nil {
		return nil, apierrors.ErrListProjectResourceUsage.InternalError(err)
	}
	return httpserver.OkResp(metricsResp["data"])
}

// GetNSInfo 获取项目级命名空间信息
func (e *Endpoints) GetNSInfo(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 检查projectID合法性
	projectID, err := strutil.Atoi64(vars["projectID"])
	if err != nil {
		return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetProject.NotLogin().ToResp(), nil
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(projectID),
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			// 若非项目管理员，判断用户是否为企业管理员(数据中心)
			orgIDStr := r.Header.Get(httputil.OrgHeader)
			orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
			if err != nil {
				return apierrors.ErrGetProject.InvalidParameter(err).ToResp(), nil
			}
			req := apistructs.PermissionCheckRequest{
				UserID:   userID.String(),
				Scope:    apistructs.OrgScope,
				ScopeID:  orgID,
				Resource: apistructs.ProjectResource,
				Action:   apistructs.GetAction,
			}
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrGetProject.AccessDenied().ToResp(), nil
			}
		}
	}

	prjNsInfo, err := e.project.GetProjectNSInfo(projectID)
	if err != nil {
		return nil, err
	}

	return httpserver.OkResp(*prjNsInfo)
}

// ListMyProjectIDs list my projectIDs
func (e *Endpoints) ListMyProjectIDs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListProject.NotLogin().ToResp(), nil
	}

	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrListProject.InvalidParameter(err).ToResp(), nil
	}

	ids, err := e.project.GetMyProjectIDList(orgID, identityInfo.UserID)
	if err != nil {
		return apierrors.ErrListProjectID.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(ids)
}

// GetProjectListByStates list projects by states
func (e *Endpoints) GetProjectListByStates(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrListProjectByStates.AccessDenied().ToResp(), nil
	}
	if r.Body == nil {
		return apierrors.ErrListProjectByStates.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.GetProjectIDListByStatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrListProjectByStates.InvalidParameter(err).ToResp(), nil
	}

	total, list, err := e.project.GetProjectIDListByStates(req.StateReq, req.ProIDs)
	if err != nil {
		return apierrors.ErrListProjectByStates.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(apistructs.GetProjectIDListByStatesData{
		Total: total,
		List:  list,
	})
}

// GetAllProjects get all projects
func (e *Endpoints) GetAllProjects(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrListAllProject.AccessDenied().ToResp(), nil
	}
	projects, err := e.project.GetAllProjects()
	if err != nil {
		return apierrors.ErrListAllProject.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(projects)
}
