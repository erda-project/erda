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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// CreateAutoTestSceneInput 创建场景入参
func (e *Endpoints) CreateAutoTestSceneInput(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrCreateAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestSceneInput.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo
	req.SceneID = id

	// TODO: 鉴权
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: req.SceneID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrCreateAutoTestSceneInput.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateAutoTestSceneInput.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateAutoTestSceneInput.AccessDenied().ToResp(), nil
		}
	}

	sceneID, err := e.autotestV2.CreateAutoTestSceneInput(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.autotestV2.UpdateAutotestSceneUpdateTime(sc.ID); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(sceneID)
}

// UpdateAutoTestSceneInput 更新场景入参
func (e *Endpoints) UpdateAutoTestSceneInput(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneInputUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}
	req.SceneID = id

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneInput.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: req.SceneID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrUpdateAutoTestSceneInput.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateAutoTestSceneInput.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateAutoTestSceneInput.AccessDenied().ToResp(), nil
		}
	}

	req.SpaceID = sp.ID
	sceneID, err := e.autotestV2.UpdateAutoTestSceneInput(req)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneInput.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// ListAutoTestSceneInput 获取场景入参列表
func (e *Endpoints) ListAutoTestSceneInput(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrListAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestSceneInput.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo
	req.SceneID = id

	//TODO 鉴权
	sceneID, err := e.autotestV2.ListAutoTestSceneInput(req.SceneID)
	if err != nil {
		return apierrors.ErrListAutoTestSceneInput.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// DeleteAutoTestSceneInput 删除场景入参列表
func (e *Endpoints) DeleteAutoTestSceneInput(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneInput.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	req.SceneID = id

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneInput.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: req.SceneID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrDeleteAutoTestSceneInput.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return apierrors.ErrDeleteAutoTestSceneInput.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteAutoTestSceneInput.AccessDenied().ToResp(), nil
		}
	}

	id, err = e.autotestV2.DeleteAutoTestSceneInput(id)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneInput.InternalError(err).ToResp(), nil
	}

	if err := e.autotestV2.UpdateAutotestSceneUpdateTime(sc.ID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(id)
}
