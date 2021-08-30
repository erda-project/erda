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

package openapiv1

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	openapi "github.com/erda-project/erda/modules/core/openapi-ng"
	"github.com/erda-project/erda/modules/core/openapi-ng/proxy"
	openapiv1 "github.com/erda-project/erda/modules/openapi"
	apiv1 "github.com/erda-project/erda/modules/openapi/api"
	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/modules/openapi/hooks"
	discover "github.com/erda-project/erda/providers/service-discover"
)

// +provider
type provider struct {
	Log      logs.Logger
	Discover discover.Interface `autowired:"discover"`
	Router   openapi.Interface  `autowired:"openapi-router"`
	proxy    proxy.Proxy
	handler  http.Handler
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.proxy.Log = p.Log
	p.proxy.Discover = p.Discover
	hooks.Enable = false
	conf.Load()
	srv, err := openapiv1.NewServer()
	if err != nil {
		return err
	}
	p.handler = srv.Handler
	p.RegisterTo(p.Router)
	return nil
}

func (p *provider) RegisterTo(router transhttp.Router) (err error) {
	methods := make(map[string]struct{})
	for _, api := range apiv1.API {
		methods[api.Method] = struct{}{}
	}
	for method := range methods {
		router.Add(method, "/**", p.handler.ServeHTTP)
	}
	return nil
}

func init() {
	servicehub.Register("openapi-v1-routes", &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
