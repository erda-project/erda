package orgapis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric for org center
	routes.GET("/api/orgCenter/metrics/:scope", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/orgCenter/metrics/:scope/:aggregate", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/orgCenter/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))
	routes.POST("/api/orgCenter/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))

	// clusters resources for org center
	checkOrgName := permission.OrgIDByOrgName("orgName")
	routes.GET("/api/resources/types", p.getHostTypes, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/group", p.getGroupHosts, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/:instance_type", p.getContainers, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/group/allocation/:metric_type", p.groupContainerAllocation, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/group/count", p.groupContainerCount, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/host-status", p.getHostStatus, permission.Intercepter(
		permission.ScopeOrg, p.getOrgIDNameFromBody,
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/hosts/actions/offline", p.offlineHost)
	routes.GET("/api/org/clusters/status", p.clusterStatus, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		pkg.ResourceOrgCenter, permission.ActionList,
	))
	return nil
}
