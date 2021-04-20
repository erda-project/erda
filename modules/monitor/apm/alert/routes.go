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

package alert

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) initRoutes(routes httpserver.Router) error {

	// 微服务告警
	routes.GET("/api/apm/alerts/rules", p.queryAlertRule, p.getPermissionByTenantGroup(permission.ActionList))          // √
	routes.GET("/api/apm/alerts", p.queryAlert, p.getPermissionByTenantGroup(permission.ActionList))                    // √
	routes.GET("/api/apm/alert/:id", p.getAlert, p.getPermissionByTenantGroup(permission.ActionGet))                    // √
	routes.POST("/api/apm/alert", p.createAlert, p.getPermissionByTenantGroup(permission.ActionCreate))                 // √
	routes.PUT("/api/apm/alert/:id", p.updateAlert, p.getPermissionByTenantGroup(permission.ActionUpdate))              // √
	routes.PUT("/api/apm/alert/:id/switch", p.updateAlertEnable, p.getPermissionByTenantGroup(permission.ActionUpdate)) // √
	routes.DELETE("/api/apm/alert/:id", p.deleteAlert, p.getPermissionByTenantGroup(permission.ActionDelete))           // √

	// 微服务自定义告警
	routes.GET("/api/apm/customize/alerts/metrics", p.queryCustomizeMetric, p.getPermissionByTenantGroup(permission.ActionGet))                // V
	routes.GET("/api/apm/customize/alerts/notifies/targets", p.queryCustomizeNotifyTarget, p.getPermissionByTenantGroup(permission.ActionGet)) // √
	routes.GET("/api/apm/customize/alerts", p.queryCustomizeAlerts, p.getPermissionByTenantGroup(permission.ActionList))                       // √
	routes.GET("/api/apm/customize/alert/:id", p.getCustomizeAlert, p.getPermissionByTenantGroup(permission.ActionGet))                        // √
	routes.POST("/api/apm/customize/alert", p.createCustomizeAlert, p.getPermissionByTenantGroup(permission.ActionCreate))                     // √
	routes.PUT("/api/apm/customize/alert/:id", p.updateCustomizeAlert, p.getPermissionByTenantGroup(permission.ActionUpdate))                  // √
	routes.PUT("/api/apm/customize/alert/:id/switch", p.updateCustomizeAlertEnable, p.getPermissionByTenantGroup(permission.ActionUpdate))     // √
	routes.DELETE("/api/apm/customize/alert/:id", p.deleteCustomizeAlert, p.getPermissionByTenantGroup(permission.ActionDelete))               // √

	// 微服务告警记录
	return nil
}
