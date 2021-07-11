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

package testbackend

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/openapi-ng/testapis"
)

type provider struct {
	Log logs.Logger
}

func (p *provider) Init(ctx servicehub.Context) error {
	routes := ctx.Service("http-server@test").(httpserver.Router)

	fsHandler := http.FileServer(testapis.FileSystemWithPrefix("/static"))
	routes.GET("/static", fsHandler.ServeHTTP)
	routes.GET("/static/**", fsHandler.ServeHTTP)

	routes.GET("/api/hello/:name",
		func(params struct {
			Name string `param:"name"`
		}) interface{} {
			return api.Success(fmt.Sprintf("hello %s", params.Name))
		},
	)

	routes.POST("/api/hello/:name",
		func(params struct {
			Name    string `param:"name"`
			Message string `form:"message" validate:"required"`
		}) interface{} {
			return api.Success(fmt.Sprintf("hello %s, %s", params.Name, params.Message))
		},
	)

	routes.GET("/api/error",
		func() interface{} {
			return api.Errors.Internal("test error")
		},
	)

	routes.GET("/api/websocket", p.handleWebSocket)

	return nil
}

func init() {
	servicehub.Register("test-backend-apis", &servicehub.Spec{
		Dependencies: []string{"http-server@test"},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
