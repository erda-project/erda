package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

// GetLicense 获取授权情况
func (e *Endpoints) GetLicense(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	if e.license == nil {
		return apierrors.ErrGetLicense.InvalidState("license is empty").ToResp(), nil
	}

	resp := apistructs.LicenseResponse{
		License: e.license,
		Valid:   true,
	}
	hostCount, err := e.host.GetHostNumber()
	if err != nil {
		return apierrors.ErrGetLicense.InternalError(err).ToResp(), nil
	}
	resp.CurrentHostCount = hostCount

	if e.license.IsExpired() {
		resp.Valid = false
		resp.Message = "已过期"
		return httpserver.OkResp(resp)
	}
	if e.license.Data.MaxHostCount < hostCount {
		resp.Valid = false
		resp.Message = "超过最大host数"
		return httpserver.OkResp(resp)
	}
	return httpserver.OkResp(resp)
}
