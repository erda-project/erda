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

// CreateAutoTestSceneStep 创建场景步骤
func (e *Endpoints) CreateAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrCreateAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestSceneStep.NotLogin().ToResp(), nil
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
		return apierrors.ErrCreateAutoTestSceneStep.InvalidState("所属测试空间已锁定").ToResp(), nil
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
			return apierrors.ErrCreateAutoTestSceneStep.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateAutoTestSceneStep.AccessDenied().ToResp(), nil
		}
	}

	req.SetID = sc.SetID
	req.SpaceID = sc.SpaceID
	sceneID, err := e.autotestV2.CreateAutoTestSceneStep(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.autotestV2.UpdateAutotestSceneUpdateTime(sc.ID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(sceneID)
}

// UpdateAutoTestSceneStep 更新场景步骤
func (e *Endpoints) UpdateAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["stepID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.NotLogin().ToResp(), nil
	}

	req.ID = id
	req.IdentityInfo = identityInfo
	//TODO 鉴权

	step, err := e.autotestV2.GetAutoTestSceneStep(req.ID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sc, err := e.autotestV2.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: step.SceneID})
	if err != nil {
		return errorresp.ErrResp(err)
	}
	sp, err := e.autotestV2.GetSpace(sc.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !sp.IsOpen() {
		return apierrors.ErrUpdateAutoTestSceneStep.InvalidState("所属测试空间已锁定").ToResp(), nil
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
			return apierrors.ErrUpdateAutoTestSceneStep.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateAutoTestSceneStep.AccessDenied().ToResp(), nil
		}
	}

	sceneID, err := e.autotestV2.UpdateAutoTestSceneStep(req)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// MoveAutoTestSceneStep 移动场景步骤
func (e *Endpoints) MoveAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.NotLogin().ToResp(), nil
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
		return apierrors.ErrUpdateAutoTestSceneStep.InvalidState("所属测试空间已锁定").ToResp(), nil
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
			return apierrors.ErrUpdateAutoTestSceneStep.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrUpdateAutoTestSceneStep.AccessDenied().ToResp(), nil
		}
	}

	if err := e.autotestV2.MoveAutoTestSceneStep(req); err != nil {
		return apierrors.ErrUpdateAutoTestSceneStep.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(req.ID)
}

// GetAutoTestSceneStep 获取场景步骤
func (e *Endpoints) GetAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["stepID"], 10, 64)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	step, err := e.autotestV2.GetAutoTestSceneStep(id)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(step.Convert())
}

// ListAutoTestSceneStep 获取场景步骤列表
func (e *Endpoints) ListAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["sceneID"], 10, 64)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStep.NotLogin().ToResp(), nil
	}

	req.IdentityInfo = identityInfo
	req.SceneID = id

	//TODO 鉴权

	sceneID, err := e.autotestV2.ListAutoTestSceneStep(req.SceneID)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStep.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(sceneID)
}

// DeleteAutoTestSceneStep 删除场景步骤
func (e *Endpoints) DeleteAutoTestSceneStep(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	id, err := strconv.ParseUint(vars["stepID"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneStep.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.AutotestSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrListAutoTestSceneStepOutPut.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneStep.NotLogin().ToResp(), nil
	}

	req.ID = id
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
		return apierrors.ErrDeleteAutoTestSceneStep.InvalidState("所属测试空间已锁定").ToResp(), nil
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
			return apierrors.ErrDeleteAutoTestSceneStep.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteAutoTestSceneStep.AccessDenied().ToResp(), nil
		}
	}

	err = e.autotestV2.DeleteAutoTestSceneStep(id)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneStep.InternalError(err).ToResp(), nil
	}

	err = e.autotestV2.UpdateAutotestSceneUpdater(req.SceneID, req.UserID)
	if err != nil {
		return apierrors.ErrDeleteAutoTestScene.InternalError(err).ToResp(), nil
	}

	if err := e.autotestV2.UpdateAutotestSceneUpdateTime(sc.ID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("delete success")
}

// ListAutoTestSceneStepOutPut 获取场景步骤出参
func (e *Endpoints) ListAutoTestSceneStepOutPut(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析请求
	var req apistructs.AutotestListStepOutPutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrListAutoTestSceneStepOutPut.InvalidParameter(err).ToResp(), nil
	}

	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStepOutPut.NotLogin().ToResp(), nil
	}

	//TODO 鉴权

	mp, err := e.autotestV2.AutoTestGetStepOutPut(req.List)
	if err != nil {
		return apierrors.ErrListAutoTestSceneStepOutPut.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(mp)
}
