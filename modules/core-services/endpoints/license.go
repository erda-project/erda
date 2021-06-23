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
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
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
	var hostCount uint64
	//hostCount, err := e.host.GetHostNumber()
	//if err != nil {
	//	return apierrors.ErrGetLicense.InternalError(err).ToResp(), nil
	//}
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
