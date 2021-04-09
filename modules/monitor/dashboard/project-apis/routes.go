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
