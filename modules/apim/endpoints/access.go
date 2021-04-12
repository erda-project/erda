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
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"

	"github.com/erda-project/erda/modules/apim/services/apierrors"
)

// 创建一个访问管理条目
func (e *Endpoints) CreateAccess(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateAccess.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateAccess.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var body apistructs.CreateAccessBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return apierrors.CreateAccess.InvalidParameter("invalid request body").ToResp(), nil
	}

	var req = apistructs.CreateAccessReq{
		OrgID:    orgID,
		Identity: &identity,
		Body:     &body,
	}

	data, apiError := e.assetSvc.CreateAccess(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// 查询访问条目列表
func (e *Endpoints) ListAccess(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListAccess.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListAccess.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var queryParams apistructs.ListAccessQueryParams
	if err = e.queryStringDecoder.Decode(&queryParams, r.URL.Query()); err != nil {
		return apierrors.ListAccess.InvalidParameter("invalid query parameters").ToResp(), nil
	}

	var req = apistructs.ListAccessReq{
		OrgID:       orgID,
		Identity:    &identity,
		QueryParams: &queryParams,
	}

	data, apiError := e.assetSvc.ListAccess(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) GetAccess(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.GetAccess.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.GetAccess.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.GetAccessReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.GetAccessURIParams{AccessID: vars[urlPathAccessID]},
	}

	data, apiError := e.assetSvc.GetAccess(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

func (e *Endpoints) DeleteAccess(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.DeleteAccess.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.DeleteAccess.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.GetAccessReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.GetAccessURIParams{AccessID: vars[urlPathAccessID]},
	}

	if apiError := e.assetSvc.DeleteAccess(&req); apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) UpdateAccess(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateAccess.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateAccess.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.UpdateAccessReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.GetAccessURIParams{AccessID: vars[urlPathAccessID]},
		Body:      new(apistructs.UpdateAccessBody),
	}

	if err = json.NewDecoder(r.Body).Decode(req.Body); err != nil {
		return apierrors.UpdateAccess.InvalidParameter("invalid body").ToResp(), nil
	}

	access, apiError := e.assetSvc.UpdateAccess(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(map[string]interface{}{"access": access})
}
