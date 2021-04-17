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

	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) reloadActionExecutorConfig(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	_, cfgChan, err := e.dbClient.ListPipelineConfigsOfActionExecutor()
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "RELOAD PIPENGINE CONFIG: LIST", err.Error())
	}

	if err := actionexecutor.GetManager().Initialize(cfgChan); err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "RELOAD PIPENGINE CONFIG: RELOAD", err.Error())
	}
	return httpserver.OkResp(nil)
}

func (e *Endpoints) throttlerSnapshot(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	snapshot := e.reconciler.TaskThrottler.Export()
	w.Write(snapshot)
	return nil
}
