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

package details_apis

import (
	"net/http"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric for alert
	routes.GET("/api/alert/metrics/:scope", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.getOrgIDByClusters,
		common.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/alert/metrics/:scope/:aggregate", p.metricq.HandleV1, permission.Intercepter(
		permission.ScopeOrg, p.getOrgIDByClusters,
		common.ResourceOrgCenter, permission.ActionGet,
	))

	// metrics for system
	routes.GET("/api/system/addon/metrics/:scope/:aggregate", p.systemAddonMetrics, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDByCluster("filter_cluster_name"),
		common.ResourceOrgCenter, permission.ActionGet,
	))
	return nil
}

func (p *provider) systemAddonMetrics(r *http.Request, params *struct {
	metricq.QueryParams
	AddonID string `query:"filter_addon_id" validate:"required"`
}) interface{} {
	return p.metricq.HandleV1(r, &params.QueryParams)
}
