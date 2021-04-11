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

package template

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.POST("/api/dashboard/template", p.createTemplate)
	routes.GET("/api/dashboard/templates", p.getListTemplates)
	routes.GET("/api/dashboard/template/:id", p.getTemplate)
	routes.PUT("/api/dashboard/template/:id", p.updateTemplate)
	routes.DELETE("/api/dashboard/template/:id", p.deleteTemplate)

	// routes.POST("/api/dashboard/template", p.createTemplate, p.getPermissionByScopeId(permission.ActionCreate))
	// routes.GET("/api/dashboard/templates", p.listTemplate, p.getPermissionByScopeId(permission.ActionList))
	// routes.GET("/api/dashboard/template/:id", p.getTemplate, p.getPermissionByScopeId(permission.ActionGet))
	// routes.PUT("/api/dashboard/template/:id", p.updateTemplate, p.getPermissionByScopeId(permission.ActionUpdate))
	// routes.DELETE("/api/dashboard/template/:id", p.deleteTemplate, p.getPermissionByScopeId(permission.ActionDelete))
	return nil
}

func (p *provider) getPermissionByScopeId(action permission.Action) httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.ScopeIdFromParams(p.authDb),
		common.ResourceMicroService, action,
	)
}
