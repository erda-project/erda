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

package runtimeapis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric for project view
	checkProject := permission.QueryValue("filter_project_id")
	routes.GET("/api/project/metrics/:scope", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeProject, checkProject,
		common.ResourceProject, permission.ActionGet,
	))
	routes.GET("/api/project/metrics/:scope/:aggregate", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeProject, checkProject,
		common.ResourceProject, permission.ActionGet,
	))
	routes.GET("/api/project/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeProject, checkProject,
		common.ResourceProject, permission.ActionGet,
	))
	routes.POST("/api/project/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeProject, checkProject,
		common.ResourceProject, permission.ActionGet,
	))
	return nil
}
