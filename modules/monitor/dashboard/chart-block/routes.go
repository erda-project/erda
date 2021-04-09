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
