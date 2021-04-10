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
