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
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda/modules/core/openapi-ng"
	"github.com/erda-project/erda/modules/core/openapi-ng/proxy"
	openapiv1 "github.com/erda-project/erda/modules/openapi"
	apiv1 "github.com/erda-project/erda/modules/openapi/api"
	"github.com/erda-project/erda/modules/openapi/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/modules/openapi/hooks"
	"github.com/erda-project/erda/modules/openapi/settings"
	discover "github.com/erda-project/erda/providers/service-discover"
)

type config struct {
	CP types.ComponentProtocolConfigs `file:"component-protocol"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Discover discover.Interface `autowired:"discover"`
	Router   openapi.Interface  `autowired:"openapi-router"`
	proxy    proxy.Proxy
	handler  http.Handler
	Settings settings.OpenapiSettings `autowired:"openapi-settings"`
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.proxy.Log = p.Log
	p.proxy.Discover = p.Discover
	hooks.Enable = false
	conf.Load()
	srv, err := openapiv1.NewServer(p.Settings)
	if err != nil {
		return err
	}
	p.handler = srv.Handler
	types.CPConfigs = p.Cfg.CP
	return p.RegisterTo(p.Router)
}

func (p *provider) RegisterTo(router transhttp.Router) (err error) {
	for _, api := range apiv1.API {
		newPath := replaceOpenapiV1Path(api.Path.String())
		router.Add(api.Method, newPath, p.handler.ServeHTTP)
	}
	// v1 router add backport methods routes
	router.Add("", "/**", p.handler.ServeHTTP)
	return nil
}

func replaceOpenapiV1Path(path string) string {
	path = strings.ReplaceAll(path, "<*>", "**")
	newPath := strings.NewReplacer("<", "{", ">", "}", " ", "_").Replace(path)
	return newPath
}

func init() {
	servicehub.Register("openapi-v1-routes", &servicehub.Spec{
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
