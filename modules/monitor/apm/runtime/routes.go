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

package runtime

import (
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (runtime *provider) initRoutes(routes httpserver.Router) error {

	routes.GET("/api/apm/runtime", runtime.runtime, getRuntimePermission(runtime.db))
	routes.GET("/api/apm/tk", runtime.tkByProjectIdAndWorkspace, getProjectPermission())
	routes.GET("/api/apm/instance", runtime.instanceByTk, getInstancePermission(runtime.db))

	return nil
}

func (runtime *provider) instanceByTk(params struct {
	TerminusKey string `query:"terminusKey" validate:"required"`
}) interface{} {
	instance, err := runtime.getInstanceByTk(params.TerminusKey)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(instance)
}

func (runtime *provider) tkByProjectIdAndWorkspace(params struct {
	ProjectId string `query:"projectId" validate:"required"`
	Workspace string `query:"workspace" validate:"required"`
}) interface{} {
	tk, err := runtime.getTkByProjectIdAndWorkspace(params.ProjectId, params.Workspace)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(tk)
}

func (runtime *provider) runtime(params Vo) interface{} {
	runtimeInfo, err := runtime.getRuntime(params)
	if err != nil {
		return api.Success(nil)
	}

	return api.Success(runtimeInfo)
}
