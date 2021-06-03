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

package testapis

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng"
	"github.com/erda-project/erda/modules/openapi-ng/api"
	"github.com/erda-project/erda/modules/openapi-ng/testapis"
)

// +provider
type provider struct {
	Router openapi.Interface `autowired:"openapi-ng"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	fsHandler := http.FileServer(testapis.FileSystemWithPrefix("/static"))
	p.Router.Add(http.MethodGet, "/static", fsHandler.ServeHTTP)
	p.Router.Add(http.MethodGet, "/static/**", fsHandler.ServeHTTP)

	p.Router.AddAPI(&api.Spec{
		Method:      http.MethodGet,
		Path:        "/api/test/hello/{name}",
		BackendPath: "/api/hello/{name}",
		Service:     "test-apis",
	})

	p.Router.AddAPI(&api.Spec{
		Method:      http.MethodPost,
		Path:        "/api/test/hello/{name}",
		BackendPath: "/api/hello/{name}",
		Service:     "test-apis",
	})

	p.Router.AddAPI(&api.Spec{
		Method:      http.MethodGet,
		Path:        "/api/test/error",
		BackendPath: "/api/error",
		Service:     "test-apis",
	})

	p.Router.AddAPI(&api.Spec{
		Method:      http.MethodGet,
		Path:        "/api/test/websocket",
		BackendPath: "/api/websocket",
		Service:     "test-apis",
	})
	return nil
}

func init() {
	servicehub.Register("openapi-test-apis", &servicehub.Spec{
		Services: []string{"openapi-test-apis"},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
