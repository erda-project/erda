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
	"sort"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

const (
	urlPathAssetID         = "assetID"
	urlPathVersionID       = "versionID"
	urlPathSwaggerVersion  = "swaggerVersion"
	urlPathMinor           = "minor"
	urlPathInstantiationID = "instantiationID"
	urlPathClientID        = "clientID"
	urlPathContractID      = "contractID"
	urlPathAccessID        = "accessID"
	urlPathProjectID       = "projectID"
	urlPathSLAID           = "slaID"
)

// CreateAPIAsset 创建 API 资料
func (e *Endpoints) CreateAPIAsset(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateAPIAsset.NotLogin().ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.CreateAPIAsset.MissingParameter(apierrors.MissingRequestBody).ToResp(), nil
	}
	var req apistructs.APIAssetCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.CreateAPIAsset.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	assetID, err := e.assetSvc.CreateAPIAsset(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(assetID)
}

// GetAPIAsset 查询 API 资料
func (e *Endpoints) GetAPIAsset(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.GetAPIAsset.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.GetAPIAsset.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	req := apistructs.GetAPIAssetReq{
		OrgID:     orgID,
		Identity:  &identityInfo,
		URIParams: &apistructs.GetAPIAssetURIPrams{AssetID: vars[urlPathAssetID]},
	}

	data, err := e.assetSvc.GetAsset(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data, []string{data.Asset.CreatorID})
}

// PagingAPIAssets 分页查询 API 资料
func (e *Endpoints) PagingAPIAssets(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.PagingAPIAssets.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.PagingAPIAssets.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var queryParams apistructs.PagingAPIAssetsQueryParams
	if err := e.queryStringDecoder.Decode(&queryParams, r.URL.Query()); err != nil {
		return apierrors.PagingAPIAssets.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.PagingAPIAssetsReq{
		OrgID:       orgID,
		Identity:    &identityInfo,
		QueryParams: &queryParams,
	}

	result, err := e.assetSvc.PagingAsset(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	sort.Slice(result.List, func(i, j int) bool {
		return result.List[i].Asset.UpdatedAt.After(result.List[j].Asset.UpdatedAt)
	})

	return httpserver.OkResp(result, result.UserIDs)
}

// UpdateAPIAsset 修改 API 资料
func (e *Endpoints) UpdateAPIAsset(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateAPIAsset.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateAPIAsset.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var (
		keys map[string]interface{}
	)
	if err = json.NewDecoder(r.Body).Decode(&keys); err != nil {
		return apierrors.UpdateAPIAsset.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.UpdateAPIAssetReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.UpdateAPIAssetURIParams{AssetID: vars["assetID"]},
		Keys:      keys,
	}

	if apiError := e.assetSvc.UpdateAPIAsset(&req); apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

// DeleteAPIAsset 删除 API 资料
func (e *Endpoints) DeleteAPIAsset(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.DeleteAPIAsset.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.DeleteAPIAsset.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.APIAssetDeleteRequest{
		OrgID:        orgID,
		AssetID:      vars[urlPathAssetID],
		IdentityInfo: identityInfo,
	}

	if err = e.assetSvc.DeleteAssetByAssetID(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}
