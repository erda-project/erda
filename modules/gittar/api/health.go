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

package api

import (
	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

const (
	PodName       = "gittar"
	ComponentName = "gittar"
	MysqlName     = "mysql-connection"
)

// Health health check
func Health(ctx *webcontext.Context) {
	status := apistructs.HealthStatusOk
	modules := make([]apistructs.Module, 0)

	// check mysql server
	status = mysqlHealth(ctx, &modules)
	health := apistructs.HealthResponse{
		Name:    ComponentName,
		Status:  status,
		Modules: modules,
		Tags: map[string]string{
			"pod_name": PodName,
			"version":  version.Version,
		},
	}
	ctx.EchoContext.JSON(200, health)
}

// mysqlHealth check mysql health
func mysqlHealth(ctx *webcontext.Context, modules *[]apistructs.Module) (status apistructs.HealthStatus) {
	if err := ctx.DBClient.Ping(); err != nil {
		status = apistructs.HealthStatusFail
		*modules = append(*modules, apistructs.Module{
			Name:    MysqlName,
			Status:  status,
			Message: err.Error(),
		})
	} else {
		status = apistructs.HealthStatusOk
		*modules = append(*modules, apistructs.Module{
			Name:    MysqlName,
			Status:  status,
			Message: "connected",
		})
	}
	return
}
