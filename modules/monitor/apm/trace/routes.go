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

package trace

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) initRoutes(routes httpserver.Router) error {

	// 链路追踪
	routes.GET("/api/apm/traces", p.traces)
	routes.GET("/api/apm/trace/:traceId", p.traceOne)
	routes.GET("/api/apm/trace/debugs", p.traceDebugRecords)
	return nil
}
