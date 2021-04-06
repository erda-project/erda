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
	"io/ioutil"
	"net/http"

	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/pkg/httpserver"
)

// 创建service元数据，由调度器来填写
func (e *Endpoints) serviceCreateOrUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		service types.CmService
		body    []byte
		err     error
	)

	if body, err = ioutil.ReadAll(r.Body); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: "read http request body error."}, err
	}

	if err = json.Unmarshal(body, &service); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: "unmarshal failed."}, err
	}

	service.Cluster = vars["cluster"]
	service.DiceProject = vars["project"]
	service.DiceApplication = vars["application"]
	service.DiceRuntime = vars["runtime"]
	service.DiceService = vars["service"]

	if err = e.db.CreateOrUpdateService(ctx, &service); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError, Content: "update service to database failed."}, err
	}

	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: &service,
	}, err
}
