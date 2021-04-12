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

package admin_tools

import (
	"fmt"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/rakyll/statik/fs"
)

//go:generate statik -src=./web/public -ns "monitor/admin-tools" -include=*.jpg,*.png,*ico,*.txt,*.html,*.css,*.js,*.js.map
func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/admin/envs", p.showEnvs)
	routes.GET("/api/admin/version", p.showVersionInfo)

	routes.GET("/api/admin/es/indices", p.showIndicesByDate)
	routes.GET("/api/admin/es/indices/:wildcard", p.showIndicesByDate)
	routes.DELETE("/api/admin/es/indices/:wildcard", p.deleteIndices)

	routes.GET("/api/admin/kafka/channel", p.showKafkaChannelSize)
	routes.POST("/api/admin/kafka/push", p.pushKafkaData)

	routes.Any("/api/admin/proxy", p.proxy, interceptors.CORS())
	routes.Any("/api/admin/proxy/*", p.proxy, interceptors.CORS())

	assets, err := fs.NewWithNamespace("monitor/admin-tools")
	if err != nil {
		return fmt.Errorf("fail to init file system: %s", err)
	}
	fs := assets
	routes.Static("/", "/", httpserver.WithFileSystem(fs))
	return nil
}
