// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/webcontext"
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
	httpStatus := http.StatusOK
	if status != apistructs.HealthStatusOk {
		httpStatus = http.StatusServiceUnavailable
	}
	ctx.EchoContext.JSON(httpStatus, health)
}

// mysqlHealth check mysql health
func mysqlHealth(ctx *webcontext.Context, modules *[]apistructs.Module) (status apistructs.HealthStatus) {
	if ctx.DBClient == nil || ctx.DBClient.DB == nil {
		status = apistructs.HealthStatusFail
		*modules = append(*modules, apistructs.Module{
			Name:    MysqlName,
			Status:  status,
			Message: "db client not initialized",
		})
		return
	}

	pingCtx, cancel := context.WithTimeout(ctx.EchoContext.Request().Context(), 3*time.Second)
	defer cancel()
	if err := ctx.DBClient.DB.DB().PingContext(pingCtx); err != nil {
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
