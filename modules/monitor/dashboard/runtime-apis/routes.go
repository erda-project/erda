package runtimeapis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric for runtime view
	checkApplication := permission.QueryValue("filter_application_id")
	routes.GET("/api/runtime/metrics/:scope", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeApp, checkApplication,
		common.ResourceRuntime, permission.ActionGet,
	))
	routes.GET("/api/runtime/metrics/:scope/:aggregate", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeApp, checkApplication,
		common.ResourceRuntime, permission.ActionGet,
	))
	routes.GET("/api/runtime/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeApp, checkApplication,
		common.ResourceRuntime, permission.ActionGet,
	))
	routes.POST("/api/runtime/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeApp, checkApplication,
		common.ResourceRuntime, permission.ActionGet,
	))
	return nil
}
