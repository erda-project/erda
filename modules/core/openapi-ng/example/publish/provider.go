// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package static

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	openapi "github.com/erda-project/erda/modules/core/openapi-ng"
	"github.com/erda-project/erda/modules/core/openapi-ng/proxy"
	discover "github.com/erda-project/erda/providers/service-discover"
)

// +provider
type provider struct {
	Log      logs.Logger
	Discover discover.Interface `autowired:"discover"`
	proxy    proxy.Proxy
}

var _ openapi.RouteSource = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.proxy.Log = p.Log
	p.proxy.Discover = p.Discover
	return nil
}

func (p *provider) Name() string { return "route-source-example" }
func (p *provider) RegisterTo(router transhttp.Router) (err error) {
	registerAPIs(func(method, path, backendPath, service string) {
		if err != nil {
			return
		}
		handler, e := p.proxy.Wrap(method, path, backendPath, service)
		if e != nil {
			err = e
			return
		}
		router.Add(method, path, transhttp.HandlerFunc(handler))
	})
	return err
}

func registerAPIs(add func(method, path, backendPath, service string)) {
	add("GET", "/api/example/hello", "/api/hello", "example")
	add("POST", "/api/example/hello", "/api/hello", "example")
	add("GET", "/api/example/hello/{name}", "/api/hello/{name}", "example")
	add("GET", "/api/example/user-info", "/api/user-info", "example")

	// websocket
	add("GET", "/api/websocket", "/api/websocket", "example")
	add("GET", "/static/websocket.html", "/static/websocket.html", "example")
}

func init() {
	servicehub.Register("openapi-example", &servicehub.Spec{
		Services: []string{"openapi-route-example"},
		Creator:  func() servicehub.Provider { return &provider{} },
	})
}
