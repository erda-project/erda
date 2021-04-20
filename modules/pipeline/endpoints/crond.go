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

	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) crondReload(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	logs, err := e.crondSvc.ReloadCrond(e.pipelineSvc.RunCronPipelineFunc)
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "CROND_RELOAD", err.Error())
	}
	return httpserver.OkResp(logs)
}

func (e *Endpoints) crondSnapshot(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	return httpserver.OkResp(http.StatusOK, e.crondSvc.CrondSnapshot())
}
