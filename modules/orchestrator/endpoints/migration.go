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

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// MigrationLog migration log接口
func (e *Endpoints) MigrationLog(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//鉴权
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateAddon.NotLogin().ToResp(), nil
	}
	// 获取日志
	var logReq apistructs.DashboardSpotLogRequest
	if err := queryStringDecoder.Decode(&logReq, r.URL.Query()); err != nil {
		return apierrors.ErrMigrationLog.InvalidParameter(err).ToResp(), nil
	}
	logReq.ID = "migration-task-" + vars["migrationId"]
	logReq.Source = apistructs.DashboardSpotLogSourceJob
	// 查询日志信息
	log, err := e.bdl.GetLog(logReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if len(log.Lines) == 0 {
		log.Lines = []apistructs.DashboardSpotLogLine{}
	}
	return httpserver.OkResp(log)
}

// 清理migration namespace
func (e *Endpoints) CleanUnusedMigrationNs() (bool, error) {
	return e.migration.CleanUnusedMigrationNs()
}

var queryStringDecoder *schema.Decoder

func init() {
	queryStringDecoder = schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
}
