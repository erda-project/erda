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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) ListAPIGateways(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateContract.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateContract.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.ListAPIGatewaysReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.ListAPIGatewaysURIParams{AssetID: vars[urlPathAssetID]},
	}

	list, apiError := e.assetSvc.ListAPIGateways(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(map[string]interface{}{"total": len(list), "list": list})
}

func (e *Endpoints) ListProjectAPIGateways(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListAPIGateways.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListAPIGateways.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.ListProjectAPIGatewaysReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.ListProjectAPIGatewaysURIParams{ProjectID: vars[urlPathProjectID]},
	}

	list, apiError := e.assetSvc.ListProjectAPIGateways(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(map[string]interface{}{"total": len(list), "list": list})
}
