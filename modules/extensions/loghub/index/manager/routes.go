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

package manager

import (
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/logs-query/indices", p.inspectIndices)
	routes.POST("/api/logs-manager/indices/:addon", p.createByAddonIndex)
	return nil
}

func (p *provider) inspectIndices(r *http.Request) interface{} {
	return api.Success(p.indices.Load())
}

func (p *provider) createByAddonIndex(params struct {
	Addon string `param:"addon" validate:"required"`
}) interface{} {
	resp, err := p.createIndex(params.Addon)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(resp)
}
