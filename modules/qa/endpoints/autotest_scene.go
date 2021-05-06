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

// CreateAutoTestScene 新建场景
func (e *Endpoints) CreateAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestScene.NotLogin().ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	set, err := e.autotestV2.GetSceneSet(req.SetID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(set.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrUpdateAutoTestScene.InvalidState("所属测试空间已锁定").ToResp(), nil
	}
	req.SpaceID = sp.ID

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateAutoTestScene.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateAutoTestScene.AccessDenied().ToResp(), nil
		}
	}

	sceneID, err := e.autotestV2.CreateAutotestScene(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(sceneID)
}

// CopyAutoTestScene 复制场景
func (e *Endpoints) CopyAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestScene.NotLogin().ToResp(), nil
	}

	var req apistructs.AutotestSceneCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	set, err := e.autotestV2.GetSceneSet(req.SetID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(set.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateAutoTestScene.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateAutoTestScene.AccessDenied().ToResp(), nil
		}
	}

	sceneID, err := e.autotestV2.CopyAutotestScene(req, false, nil)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(sceneID)
}

// UpdateAutoTestScene 更新场景
func (e *Endpoints) UpdateAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneSceneUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo
	req.SceneID = id

	//TODO 鉴权
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: req.SceneID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.SetID = sc.SetID
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrUpdateAutoTestScene.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateAutoTestScene.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateAutoTestScene.AccessDenied().ToResp(), nil
		}
	}
	sceneID, err := e.autotestV2.UpdateAutotestScene(req)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// MoveAutoTestScene 移动场景
func (e *Endpoints) MoveAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: req.ID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrUpdateAutoTestScene.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrMoveAutoTestScene.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrMoveAutoTestScene.AccessDenied().ToResp(), nil
		}
	}

	req.Name = sc.Name
	sceneID, err := e.autotestV2.MoveAutotestScene(req)
	if err != nil {
		return apierrors.ErrMoveAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// ListAutoTestScene 获取场景列表
func (e *Endpoints) ListAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	var req apistructs.AutotestSceneRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestScene.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权

	total, scenes, err := e.autotestV2.ListAutotestScene(req)
	if err != nil {
		return apierrors.ErrListAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(apistructs.AutoTestSceneList{
		List:  scenes,
		Total: total,
	})
}

func (e *Endpoints) ListAutoTestScenes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	var req apistructs.AutotestScenesRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestScene.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权

	scenes, err := e.autotestV2.ListAutotestScenes(req.SetIDs)
	if err != nil {
		return apierrors.ErrListAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(scenes)
}

// GetAutoTestScene 获取场景
func (e *Endpoints) GetAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	req.SceneID = id

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo

	//TODO 鉴权
	scene, err := e.autotestV2.GetAutotestScene(req)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(scene)
}

// DeleteAutoTestScene 删除场景
func (e *Endpoints) DeleteAutoTestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteAutoTestScene.NotLogin().ToResp(), nil
	}

	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: id})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrDeleteAutoTestScene.InvalidState("所属测试空间已锁定").ToResp(), nil
	}

	// 鉴权
	if !identityInfo.IsInternalClient() {
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  uint64(sp.ProjectID),
			Resource: apistructs.AutotestSceneResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrDeleteAutoTestScene.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteAutoTestScene.AccessDenied().ToResp(), nil
		}
	}

	err = e.autotestV2.DeleteAutotestScene(id)
	if err != nil {
		return apierrors.ErrDeleteAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("delete success")
}

// ExecuteDiceAutotestScene 执行场景
func (e *Endpoints) ExecuteDiceAutotestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AutotestExecuteSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	sceneIDStr := vars["sceneID"]
	sceneID, err := strconv.Atoi(sceneIDStr)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.AutoTestScene.ID = uint64(sceneID)

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestScene.NotLogin().ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	result, err := e.autotestV2.ExecuteDiceAutotestScene(req)
	if err != nil {
		return apierrors.ErrExecuteAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// ExecuteDiceAutotestScene 执行步骤
func (e *Endpoints) ExecuteDiceAutotestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AutotestExecuteSceneStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrExecuteAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	stepIDStr := vars["stepID"]
	sceneID, err := strconv.Atoi(stepIDStr)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.SceneStepID = uint64(sceneID)

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrExecuteAutoTestSceneStep.NotLogin().ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	result, err := e.autotestV2.ExecuteDiceAutotestSceneStep(req)
	if err != nil {
		return apierrors.ErrExecuteAutoTestSceneStep.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// CancelDiceAutotestScene 取消执行场景
func (e *Endpoints) CancelDiceAutotestScene(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var req apistructs.AutotestCancelSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	sceneIDStr := vars["sceneID"]
	sceneID, err := strconv.Atoi(sceneIDStr)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.AutoTestScene.ID = uint64(sceneID)

	err = e.autotestV2.CancelDiceAutotestScene(req)
	if err != nil {
		return apierrors.ErrCancelAutoTestScene.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("success")
}
