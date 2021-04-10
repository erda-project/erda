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

func (e *Endpoints) GetSceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["setID"], 10, 64)
	var req apistructs.SceneSetRequest
	req.IdentityInfo = identityInfo
	res, err := e.sceneset.GetSceneSet(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(res)
}

func (e *Endpoints) GetSceneSets(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// spaceID := r.URL.Query().Get("spaceID")
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	var req apistructs.SceneSetRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListAutoTestScene.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	// if req.SpaceID == nil {
	// 	return apierrors.ErrListAutoTestSceneSet.MissingParameter("spaceID").ToResp(), nil
	// }

	scenesets, err := e.sceneset.GetSceneSetsBySpaceID(req.SpaceID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(scenesets)
}

func (e *Endpoints) CreateSceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestSet.MissingParameter("request body").ToResp(), nil
	}

	var req apistructs.SceneSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAutoTestSceneSet.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectId,
			Resource: apistructs.SceneSetResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrCreateAutoTestSceneSet.AccessDenied()
		}
	}

	res, err := e.sceneset.CreateSceneSet(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(res)
}

func (e *Endpoints) UpdateSceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["setID"], 10, 64)

	var req apistructs.SceneSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestSceneSet.InvalidParameter("SceneSetID").ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectId,
			Resource: apistructs.SceneSetResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrUpdateAutoTestSceneSet.AccessDenied()
		}
	}
	res, err := e.sceneset.UpdateSceneSet(id, req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(res)
}

func (e *Endpoints) DeleteSceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["setID"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.InvalidParameter("SceneSetID").ToResp(), nil
	}
	var req apistructs.SceneSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.InvalidParameter("SceneSetID").ToResp(), nil
	}
	req.SetID = id
	req.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectId,
			Resource: apistructs.SceneSetResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrDeleteAutoTestSceneSet.AccessDenied()
		}
	}
	if err = e.sceneset.DeleteSceneSet(req); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(id)
}

func (e *Endpoints) DragSceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDragAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	var req apistructs.SceneSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDragAutoTestSceneSet.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectId,
			Resource: apistructs.SceneSetResource,
			Action:   apistructs.OperateAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrDragAutoTestSceneSet.AccessDenied()
		}
	}
	err = e.sceneset.DragSceneSet(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(req.SetID)
}

func (e *Endpoints) CopySceneSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDragAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	var req apistructs.SceneSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDragAutoTestSceneSet.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	setId, err := e.sceneset.CopySceneSet(req, false)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(setId)
}
