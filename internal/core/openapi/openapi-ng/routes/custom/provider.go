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

package dynamic

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/proxy"
	discover "github.com/erda-project/erda/internal/pkg/service-discover"
)

// +provider
type provider struct {
	Config   *config
	Log      logs.Logger
	Discover discover.Interface `autowired:"discover"`
	Auth     auth.Interface     `autowired:"openapi-auth"`
	Router   openapi.Interface  `autowired:"openapi-router"`
	proxy    proxy.Proxy
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.proxy.Log = p.Log
	p.proxy.Discover = p.Discover
	return p.RegisterTo(p.Router)
}

func (p *provider) RegisterTo(router transhttp.Router) (err error) {
	// add some custom API
	for _, route := range p.Config.Routes {
		if err = route.Validate(); err != nil {
			return err
		}
		var handler http.HandlerFunc
		var backend string
		if route.BackendServiceName != "" {
			backend = route.BackendServiceName
			handler, err = p.proxy.Wrap(route.Method, route.Path, route.BackendPath, route.BackendServiceName)
		} else {
			backend = route.BackendServiceURL
			handler, err = p.proxy.WrapWithServiceURL(route.Method, route.Path, route.BackendPath, route.BackendServiceURL)
		}
		if err != nil {
			if strings.EqualFold(route.Required, "optional") {
				handler = func(rw http.ResponseWriter, _ *http.Request) {
					rw.Header().Set("Server", "Erda Openapi")
					rw.Header().Set("Backend", backend)
					http.Error(rw, "backend not found", http.StatusNotFound)
				}
			} else {
				return err
			}
		}
		handler = p.Auth.Interceptor(handler, func(_ *http.Request) auth.Options { return route })
		router.Add(route.Method, route.Path, transhttp.HandlerFunc(handler))
	}

	return nil
}

func init() {
	servicehub.Register("openapi-custom-routes", &servicehub.Spec{
		Creator:    func() servicehub.Provider { return &provider{} },
		ConfigFunc: func() interface{} { return new(config) },
	})
}

type config struct {
	Routes []*Route `json:"routes" yaml:"routes"`
}

type Route struct {
	Method             string         `json:"method" yaml:"method"`
	Path               string         `json:"path" yaml:"path"`
	BackendServiceName string         `json:"backendServiceName" yaml:"backendServiceName"`
	BackendServiceURL  string         `json:"backendServiceURL" yaml:"backendServiceURL"`
	BackendPath        string         `json:"backendPath" yaml:"backendPath"`
	Required           string         `json:"required" yaml:"required"`
	Auth               *pb.APIAuth    `json:"auth" yaml:"auth"`
	Opts               map[string]any `json:"opts" yaml:"opts"`
}

func (r *Route) Validate() error {
	if r.Method == "" {
		r.Method = http.MethodGet
	}
	r.Method = strings.ToUpper(r.Method)
	if r.Path == "" {
		return errors.New("empty route.Path")
	}
	if !strings.HasPrefix(r.Path, "/") {
		return errors.Errorf("route.Path must has prefix /, got %s", r.Path)
	}
	if r.BackendServiceName == "" && r.BackendServiceURL == "" {
		return errors.New("must latest specifies one of route.BackendServiceName and route.BackendServiceURL")
	}
	if r.BackendServiceName == "" && r.BackendServiceURL != "" {
		u, err := url.Parse(r.BackendServiceURL)
		if err != nil {
			return errors.Wrap(err, "failed to parse route.BackendServiceURL")
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.Errorf("scheme of route.BackendServiceURL must be http or https, got %s", u.Scheme)
		}
		if len(u.Host) == 0 {
			return errors.New("host of route.BackendServiceURL must not be empty")
		}
	}
	if r.Auth == nil {
		r.Auth = new(pb.APIAuth)
	}
	if r.Opts == nil {
		r.Opts = make(map[string]any)
	}
	return nil
}

func (r *Route) Get(key string) any {
	// for "path", "method"
	field := reflect.ValueOf(r).Elem().FieldByName(key)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}

	// for "CheckLogin", "CheckToken" ...
	if r.Auth == nil {
		r.Auth = new(pb.APIAuth)
	}
	field = reflect.ValueOf(r.Auth).Elem().FieldByName(key)
	if field.IsValid() || field.CanInterface() {
		return field.Interface()
	}

	// for others
	if r.Opts == nil {
		r.Opts = make(map[string]any)
	}
	return r.Opts[key]
}

func (r *Route) Set(key string, val any) {
	if r.Opts == nil {
		r.Opts = make(map[string]any)
	}
	r.Opts[key] = val
}
