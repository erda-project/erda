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

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

//Health Component health check interface
func (e *Endpoints) Health(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	res := &apistructs.HealthResponse{
		Name:   "dop",
		Status: apistructs.HealthStatusFail,
		Tags: map[string]string{
			"pod_name": "dop",
			"version":  version.Version,
		},
	}
	res.Modules = make([]apistructs.Module, 0)
	err := e.db.DB.DB().Ping()
	if err != nil {
		sql := apistructs.Module{
			Name:    "mysql-connection",
			Status:  apistructs.HealthStatusFail,
			Message: err.Error(),
		}
		res.Modules = append(res.Modules, sql)
		return httpserver.OkResp(res)
	}

	res.Modules = append(res.Modules, apistructs.Module{
		Name:    "mysql-connection",
		Status:  apistructs.HealthStatusOk,
		Message: "connected",
	})
	res.Status = apistructs.HealthStatusOk
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: res,
	}, nil
}
