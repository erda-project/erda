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

package auth

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi-ng/api"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type config struct {
	Order              int    `default:"10"`
	RedirectAfterLogin string `redirect_after_login:""`
	DefaultAuther      string `default:"session"`
}

// +provider
type provider struct {
	Cfg  *config
	Log  logs.Logger
	HTTP httpserver.Router `autowired:"http-server"`

	autherNames  []string
	authers      map[string]Auther
	interceptors interceptors.Interceptors

	bundle *bundle.Bundle
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func (p *provider) Init(ctx servicehub.Context) error {
	ctx.Hub().ForeachServices(func(service string) bool {
		if strings.HasPrefix(service, "openapi-auth-") {
			svr := ctx.Service(service)
			if ap, ok := svr.(AutherProvider); ok {
				authers := ap.Authers()
				for _, auth := range authers {
					p.authers[auth.Name()] = auth
					p.autherNames = append(p.autherNames, auth.Name())
				}

			} else if auth, ok := svr.(Auther); ok {
				p.authers[auth.Name()] = auth
				p.autherNames = append(p.autherNames, auth.Name())
			} else {
				panic(fmt.Errorf("service %q is not AutherProvider or Auther", service))
			}
		} else if strings.HasPrefix(service, "openapi-interceptor-") && service != "openapi-interceptor-auth" {
			inters, ok := ctx.Service(service).(interceptors.Interface)
			if ok {
				for _, inter := range inters.List() {
					if inter.Order <= p.Cfg.Order {
						p.interceptors = append(p.interceptors, inter)
					}
				}
			}
		}
		return true
	})
	sort.Strings(p.autherNames)
	sort.Sort(p.interceptors)
	for _, name := range p.autherNames {
		auther := p.authers[name]
		if ah, ok := auther.(AuthHandler); ok {
			ah.RegisterHandler(p.Add)
		}
	}
	return nil
}

func (p *provider) Add(method, path string, handler http.HandlerFunc) {
	for i := len(p.interceptors) - 1; i >= 0; i-- {
		handler = p.interceptors[i].Wrapper(handler)
	}
	p.HTTP.Add(method, path, handler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r.Header.Del(httputil.UserHeader)
		r.Header.Del(httputil.OrgHeader)

		apictx := api.GetContext(r.Context())
		if apictx == nil {
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		spec := apictx.Spec()
		authers, ok := spec.Attributes["authers"]
		var names []string
		if ok {
			switch v := authers.(type) {
			case string:
				names = strings.Split(v, ",")
			case []string:
				names = v
			default:
				names = p.autherNames
			}
		} else {
			names = p.autherNames
		}
		for _, name := range names {
			auther := p.authers[name]
			if auther != nil {
				if checker, ok := auther.Match(r); ok {
					result, err := checker(r)
					if err != nil {
						err := fmt.Errorf("fail to auth: %s", err)
						p.Log.Error(err)
						http.Error(rw, err.Error(), http.StatusInternalServerError)
						return
					}
					if result.Success {
						r = r.WithContext(WithAuthInfo(r.Context(), &AuthInfo{
							Type: name,
							Data: result.Data,
						}))
						if setter, ok := result.Data.(AuthInfoSetter); ok {
							r = setter.SetAuthInfo(r)
						}
						h(rw, r)
						return
					}
				}
			}
		}
		try, _ := spec.Attributes["try_auth"].(bool)
		if try {
			h(rw, r)
			return
		}
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func init() {
	servicehub.Register("openapi-interceptor-auth", &servicehub.Spec{
		Services: []string{"openapi-interceptor-auth"},
		DependenciesFunc: func(hub *servicehub.Hub) (list []string) {
			hub.ForeachServices(func(service string) bool {
				if strings.HasPrefix(service, "openapi-auth-") ||
					(strings.HasPrefix(service, "openapi-interceptor-") && service != "openapi-interceptor-auth") {
					list = append(list, service)
				}
				return true
			})
			return list
		},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{
				authers: make(map[string]Auther),
			}
		},
	})
}
