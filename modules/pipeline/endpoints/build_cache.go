package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) reportBuildCache(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.BuildCacheImageReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrReportBuildCache.InvalidParameter(err).ToResp(), nil
	}

	cacheImage := spec.CIV3BuildCache{
		Name:        req.Name,
		ClusterName: req.ClusterName,
	}

	if err := e.buildCacheSvc.Report(&req, &cacheImage); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}
