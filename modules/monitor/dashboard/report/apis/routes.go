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

package apis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

// permission resources
const (
	ResourceOrgCenter = "monitor_org_center"
	ResourceRuntime   = "monitor_runtime"
	ResourceProject   = "monitor_project"
	ResourceOrgAlert  = "monitor_org_alert"
	ResourceNotify    = "notify"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// report task
	checkOrgID := permission.QueryValue("scopeId")
	routes.GET("/api/org/report/tasks", p.listOrgReportTasks, permission.Intercepter(
		permission.ScopeOrg, checkOrgID,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/org/report/tasks", p.creatOrgReportTask)
	routes.PUT("/api/org/report/tasks/:id", p.updateOrgReportTask)
	routes.PUT("/api/org/report/tasks/:id/switch", p.switchOrgReportTask)
	routes.GET("/api/org/report/tasks/:id", p.getOrgReportTask)
	routes.DELETE("/api/org/report/tasks/:id", p.delOrgReportTask)
	routes.POST("/api/org/report/tasks/:id/run-now", p.runReportTaskAtOnce) // for test

	// report type
	routes.GET("/api/report/types", p.listReportType)

	// report history
	routes.GET("/api/report/histories", p.listReportHistories)
	routes.POST("/api/report/histories", p.createReportHistory)
	routes.GET("/api/report/histories/:id", p.getReportHistory)
	routes.DELETE("/api/report/histories/:id", p.delReportHistory)
	return nil
}
