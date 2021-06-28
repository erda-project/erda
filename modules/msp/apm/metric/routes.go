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

package metric

import "github.com/erda-project/erda-infra/providers/httpserver"

const PATH_PREFIX = "/api/tmc"

func (p *provider) initRoutes(routes httpserver.Router) error {
	routes.GET(PATH_PREFIX+"/metrics-query", p.metricQueryByQL)
	routes.POST(PATH_PREFIX+"/metrics-query", p.metricQueryByQL)
	routes.GET(PATH_PREFIX+"/metrics/:scope", p.metricQuery)
	routes.GET(PATH_PREFIX+"/metrics/:scope/histogram", p.metricQueryHistogram)
	routes.GET(PATH_PREFIX+"/metrics/:scope/range", p.metricQueryRange)
	routes.GET(PATH_PREFIX+"/metrics/:scope/apdex", p.metricQueryApdex)
	routes.GET(PATH_PREFIX+"/metric/groups", p.listGroups)
	routes.GET(PATH_PREFIX+"/metric/groups/:id", p.getGroup)
	return nil
}
