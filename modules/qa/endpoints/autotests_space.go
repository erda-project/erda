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
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// CreateAutoTestSpace 创建测试空间
func (e *Endpoints) CreateAutoTestSpace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestSpace.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrCreateAutoTestSpace.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.AutoTestSpaceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	space, err := e.autotestV2.CreateSpace(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(space)
}

// UpdateAutoTestSpace 更新测试空间
func (e *Endpoints) UpdateAutoTestSpace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSpace.NotLogin().ToResp(), nil
	}
	fmt.Println("userID: ", identityInfo.UserID)
	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateAutoTestSpace.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.AutoTestSpace
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}

	res, err := e.autotestV2.GetSpace(req.ID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// TODO: 鉴权
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(res.ProjectID),
			Resource: apistructs.TestSpaceResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrCreateTestPlan.AccessDenied()
		}
	}

	space, err := e.autotestV2.UpdateAutoTestSpace(req, identityInfo.UserID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(space)
}

// DeleteAutoTestSpace 删除测试空间
func (e *Endpoints) DeleteAutoTestSpace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSpace.NotLogin().ToResp(), nil
	}

	var req apistructs.AutoTestSpace
	req.ID, err = strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}

	res, err := e.autotestV2.GetSpace(req.ID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// TODO: 鉴权
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(res.ProjectID),
			Resource: apistructs.TestSpaceResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrCreateTestPlan.AccessDenied()
		}
	}

	space, err := e.autotestV2.DeleteAutoTestSpace(req, identityInfo)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(space)
}

// GetAutoTestSpaceList 获取测试空间列表
func (e *Endpoints) GetAutoTestSpaceList(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	projectID, err := strconv.ParseInt(r.URL.Query().Get("projectId"), 10, 64)
	if err != nil {
		return apierrors.ErrListAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("pageNo"))
	if err != nil {
		return apierrors.ErrListAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		return apierrors.ErrListAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}
	space, err := e.autotestV2.GetSpaceList(projectID, pageNo, pageSize)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(space)
}

// GetAutoTestSpace 获取测试空间
func (e *Endpoints) GetAutoTestSpace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrGetAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}
	// TODO: 鉴权

	space, err := e.autotestV2.GetSpace(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(space)
}

// CopyAutoTestSpace 复制测试空间
func (e *Endpoints) CopyAutoTestSpace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetAutoTestSpace.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateAutoTestSpace.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.AutoTestSpace
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSpace.InvalidParameter(err).ToResp(), nil
	}

	res, err := e.autotestV2.GetSpace(req.ID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// TODO: 鉴权
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(res.ProjectID),
			Resource: apistructs.TestSpaceResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrCreateTestPlan.AccessDenied()
		}
	}
	space, err := e.autotestV2.CopyAutoTestSpace(e.sceneset, *res, identityInfo)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(space)
}
