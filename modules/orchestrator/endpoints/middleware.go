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
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httputil"
)

// ListMiddlewares 获取 addon 真实实例列表
func (e *Endpoints) ListMiddleware(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListAddon.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrListAddon.NotLogin().ToResp(), nil
	}
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)

	// 鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: "middleware",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}

	params, err := e.getMiddlewareListParams(r)
	if err != nil {
		return apierrors.ErrListAddon.InvalidParameter(err).ToResp(), nil
	}

	// 获取 middleware 列表
	middlewares, err := e.addon.ListMiddleware(orgID, params)
	if err != nil {
		return apierrors.ErrListAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(middlewares)
}

// GetMiddleware 获取 middleware 详情
func (e *Endpoints) GetMiddleware(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrFetchAddon.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchAddon.MissingParameter("orgId").ToResp(), nil
	}
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)
	// 鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: "middleware",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}
	// 查询middleware信息
	middleware, err := e.addon.GetMiddleware(orgID, userID.String(), vars["middlewareID"])
	if err != nil {
		return apierrors.ErrFetchAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(middleware)
}

// GetMiddlewareAddonClassification 获取 middleware addon分类资源占用
func (e *Endpoints) GetMiddlewareAddonClassification(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	logrus.Info("start GetMiddlewareAddonClassification")
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrFetchAddon.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchAddon.MissingParameter("orgId").ToResp(), nil
	}
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)
	// 鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: "middleware",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}

	params, err := e.getMiddlewareListParams(r)
	if err != nil {
		return apierrors.ErrListAddon.InvalidParameter(err).ToResp(), nil
	}

	// 查询middleware信息
	middleware, err := e.addon.GetMiddlewareAddonClassification(orgID, params)
	if err != nil {
		return apierrors.ErrFetchAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(middleware)
}

// GetMiddlewareAddonDaily 获取 middleware 每日addon资源占用
func (e *Endpoints) GetMiddlewareAddonDaily(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrFetchAddon.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchAddon.MissingParameter("orgId").ToResp(), nil
	}
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)
	// 鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: "middleware",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}

	params, err := e.getMiddlewareListParams(r)
	if err != nil {
		return apierrors.ErrListAddon.InvalidParameter(err).ToResp(), nil
	}

	// 查询middleware信息
	middleware, err := e.addon.GetMiddlewareAddonDaily(orgID, params)
	if err != nil {
		return apierrors.ErrFetchAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(middleware)
}

// InnerGetMiddleware 内部获取middleware详情
func (e *Endpoints) InnerGetMiddleware(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 查询middleware信息
	middleware, err := e.addon.InnerGetMiddleware(vars["middlewareID"])
	if err != nil {
		return apierrors.ErrFetchAddon.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(middleware)
}

// GetMiddlewareResource 获取 middleware 资源详情
func (e *Endpoints) GetMiddlewareResource(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrFetchAddon.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchAddon.NotLogin().ToResp(), nil
	}
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)
	// 鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: "middleware",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, err
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}
	// 查询middleware resource信息
	middlewareResource, err := e.addon.GetMiddlewareResource(vars["middlewareID"])
	if err != nil {
		return apierrors.ErrFetchAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(middlewareResource)
}

func (e *Endpoints) getMiddlewareListParams(r *http.Request) (*apistructs.MiddlewareListRequest, error) {
	var (
		projectID uint64
		err       error
	)
	projectIDStr := r.URL.Query().Get("projectId")
	if projectIDStr != "" {
		projectID, err = strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	addonName := r.URL.Query().Get("addonName")
	workspace := r.URL.Query().Get("workspace")
	switch strings.ToUpper(workspace) {
	case "", string(apistructs.DevWorkspace), string(apistructs.TestWorkspace), string(apistructs.StagingWorkspace),
		string(apistructs.ProdWorkspace):
	default:
		return nil, errors.Errorf("invalid workspace")
	}

	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, err
	}

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, err
	}

	instanceID := r.URL.Query().Get("instanceId")
	instanceIP := r.URL.Query().Get("ip")
	return &apistructs.MiddlewareListRequest{
		ProjectID:  projectID,
		AddonName:  addonName,
		Workspace:  workspace,
		PageNo:     pageNo,
		PageSize:   pageSize,
		InstanceID: instanceID,
		InstanceIP: instanceIP,
	}, nil
}
