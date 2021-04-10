package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// 查询 swaggerVersions (即版本树)
func (e *Endpoints) ListSwaggerVersions(ctx context.Context, r *http.Request,
	vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.PagingSwaggerVersion.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.PagingSwaggerVersion.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.ListSwaggerVersionsReq{
		OrgID:       orgID,
		Identity:    &identity,
		URIParams:   &apistructs.ListSwaggerVersionsURIParams{AssetID: vars[urlPathAssetID]},
		QueryParams: new(apistructs.ListSwaggerVersionsQueryParams),
	}

	if err := e.queryStringDecoder.Decode(req.QueryParams, r.URL.Query()); err != nil {
		return apierrors.PagingSwaggerVersion.InvalidParameter(err).ToResp(), nil
	}

	data, err := e.assetSvc.ListSwaggerVersions(&req)
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "", err.Error())
	}

	return httpserver.OkResp(data)
}
