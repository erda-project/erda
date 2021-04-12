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

package orgapis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric for org center
	routes.GET("/api/orgCenter/metrics/:scope", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		common.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/orgCenter/metrics/:scope/:aggregate", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		common.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/orgCenter/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		common.ResourceOrgCenter, permission.ActionGet,
	))
	routes.POST("/api/orgCenter/metrics/query", p.metricq.Handle, permission.Intercepter(
		permission.ScopeOrg, p.checkOrgMetrics,
		common.ResourceOrgCenter, permission.ActionGet,
	))

	// clusters resources for org center
	checkOrgName := permission.OrgIDByOrgName("orgName")
	routes.GET("/api/resources/types", p.getHostTypes, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/group", p.getGroupHosts, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/:instance_type", p.getContainers, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/group/allocation/:metric_type", p.groupContainerAllocation, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/containers/group/count", p.groupContainerCount, permission.Intercepter(
		permission.ScopeOrg, checkOrgName,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/host-status", p.getHostStatus, permission.Intercepter(
		permission.ScopeOrg, p.getOrgIDNameFromBody,
		common.ResourceOrgCenter, permission.ActionList,
	))
	routes.POST("/api/resources/hosts/actions/offline", p.offlineHost)
	routes.GET("/api/org/clusters/status", p.clusterStatus, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgCenter, permission.ActionList,
	))
	return nil
}
