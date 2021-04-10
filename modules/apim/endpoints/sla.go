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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// SLAs 列表
func (e *Endpoints) ListSLAs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListSLAs.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListSLAs.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	req := apistructs.ListSLAsReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.ListSLAsURIs{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
		},
		QueryParams: nil,
	}

	// 检查 query parameters 中是否传入了 clientID
	if clientIDStr := r.URL.Query().Get("clientID"); clientIDStr != "" {
		if clientID, err := strconv.ParseUint(clientIDStr, 10, 64); err == nil && clientID != 0 {
			req.QueryParams = &apistructs.ListSLAsQueries{ClientID: clientID}
		}
	}

	data, apiError := e.assetSvc.ListSLAs(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// 创建 SLA
func (e *Endpoints) CreateSLA(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateSLA.NotLogin().ToResp(), err
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateSLA.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.CreateSLAReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.ListSLAsURIs{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
		},
		Body: new(apistructs.CreateSLABody),
	}

	if err := json.NewDecoder(r.Body).Decode(req.Body); err != nil {
		logrus.Errorf("failed to Decode req.Body, err: %v", err)
		return apierrors.CreateSLA.InvalidParameter("无效的请求体").ToResp(), nil
	}

	if apiError := e.assetSvc.CreateSLA(&req); apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) GetSLA(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.GetSLA.NotLogin().ToResp(), err
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.GetSLA.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	slaID, err := strconv.ParseUint(vars[urlPathSLAID], 10, 64)
	if err != nil {
		return apierrors.GetSLA.InvalidParameter("无效的 SLA ID").ToResp(), nil
	}

	var req = apistructs.GetSLAReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.SLADetailURI{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
			SLAID:          slaID,
		},
	}

	data, apiError := e.assetSvc.GetSLA(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) DeleteSLA(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.DeleteSLA.NotLogin().ToResp(), err
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.DeleteSLA.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	slaID, err := strconv.ParseUint(vars[urlPathSLAID], 10, 64)
	if err != nil {
		return apierrors.DeleteSLA.InvalidParameter("invalid SLA ID").ToResp(), nil
	}
	var req = apistructs.DeleteSLAReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.SLADetailURI{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
			SLAID:          slaID,
		},
	}

	if apiError := e.assetSvc.DeleteSLA(&req); apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(nil)
}

func (e *Endpoints) UpdateSLA(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateSLA.NotLogin().ToResp(), err
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateSLA.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	slaID, err := strconv.ParseUint(vars[urlPathSLAID], 10, 64)
	if err != nil {
		return apierrors.UpdateSLA.InvalidParameter("无效的 slaID").ToResp(), nil
	}

	var req = apistructs.UpdateSLAReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.SLADetailURI{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
			SLAID:          slaID,
		},
		Body: new(apistructs.UpdateSLABody),
	}

	if err := json.NewDecoder(r.Body).Decode(req.Body); err != nil {
		logrus.Errorf("failed to Decode r.Body, err: %v", err)
		return apierrors.UpdateSLA.InternalError(errors.New("无效的请求体")).ToResp(), nil
	}

	if apiError := e.assetSvc.UpdateSLA(&req); apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
