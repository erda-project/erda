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

package block

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// system dashboard block
	routes.GET("/api/dashboard/system/blocks", p.listBlockSystem)
	routes.GET("/api/dashboard/system/blocks/:id", p.getSystemBlock)
	// routes.POST("/api/dashboard/system/blocks", p.createBlockSystem)
	// routes.DELETE("/api/dashboard/system/blocks/:id", p.delSystemBlock)

	// user dashboard block
	routes.POST("/api/dashboard/blocks", p.createUserBlock)
	routes.GET("/api/dashboard/blocks", p.listUserBlock)
	routes.GET("/api/dashboard/blocks/:id", p.getUserBlock)
	routes.PUT("/api/dashboard/blocks/:id", p.updateUserBlock)
	routes.DELETE("/api/dashboard/blocks/:id", p.delUserBlock)

	//routes.POST("/api/dashboard/blocks", p.createUserBlock, p.getPermissionByScopeId(permission.ActionCreate))
	//routes.GET("/api/dashboard/blocks", p.listUserBlock, p.getPermissionByScopeId(permission.ActionList))
	//routes.GET("/api/dashboard/blocks/:id", p.getUserBlock, p.getPermissionByScopeId(permission.ActionGet))
	//routes.PUT("/api/dashboard/blocks/:id", p.updateUserBlock, p.getPermissionByScopeId(permission.ActionUpdate))
	//routes.DELETE("/api/dashboard/blocks/:id", p.delUserBlock, p.getPermissionByScopeId(permission.ActionDelete))
	return nil
}
