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

package ai_proxy

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	swagger_ui "github.com/erda-project/erda/internal/apps/swagger-ui"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	reverse_proxy "github.com/erda-project/erda/internal/pkg/ai-proxy/filter/reverse-proxy"
	provider2 "github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	route2 "github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	name         = "erda.app.ai-proxy"
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.Server"},
		Summary:     "ai-proxy server",
		Description: "Reverse proxy service between AI vendors and client applications, providing a cut-through for service access",
		ConfigFunc: func() interface{} {
			return new(config) // todo: 启动时支持从初始配置文件或配置中心(Nacos, ETCD, MySQL)获取配置
		},
		Types: []reflect.Type{providerType},
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	L          logs.Logger
	Config     *config
	HttpServer httpserver.Router    `autowired:"http-server"`
	SwaggerUI  swagger_ui.Interface `autowired:"erda.app.swagger-ui.Server"`
	D          *gorm.DB             `autowired:"mysql-gorm.v2-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	if err := p.L.SetLevel(p.Config.GetLogLevel()); err != nil {
		return errors.Wrapf(err, "failed to %T.SetLevel, logLevel: %s", p.L, p.Config.GetLogLevel())
	} else {
		p.L.Infof("logLevel: %s", p.Config.GetLogLevel())
	}
	if err := p.parseRoutesConfig(); err != nil {
		return errors.Wrap(err, "failed to parseRoutesConfig")
	}
	p.L.Infof("providers config:\n%s", strutil.TryGetYamlStr(p.Config.providers))
	if err := p.parseProvidersConfig(); err != nil {
		return errors.Wrap(err, "failed to parseProvidersConfig")
	}
	p.L.Infof("routes config:\n%s", strutil.TryGetYamlStr(p.Config.Routes))

	// register http api
	// redirect to swagger page from root
	p.HttpServer.Any("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/swagger")
		w.WriteHeader(http.StatusPermanentRedirect)
	})
	// swagger ui page rendered from a html template
	p.HttpServer.Any("/swagger", p.SwaggerUI)
	// statics swagger files
	p.HttpServer.Static("/swaggers", "swaggers", http.FileServer(http.Dir("swaggers")))
	// reverse proxy to AI provider's server
	p.HttpServer.Any("/**", p.ServeAI)
	return nil
}

func (p *provider) ServeAI(w http.ResponseWriter, r *http.Request) {
	rout, ok := p.Config.Routes.FindRoute(r.URL.Path, r.Method)
	if !ok {
		p.responseNoSuchRoute(w, r.URL.Path, r.Method)
		return
	}
	var filters []filter.Filter
	for i := 0; i < len(rout.Filters); i++ {
		var name, config = rout.Filters[i].Name, rout.Filters[i].Config
		factory, ok := filter.GetFilterFactory(name)
		if !ok {
			p.L.Errorf("failed to GetFilterFactory, filter name: %s", name)
			p.responseNoSuchFilter(w, name)
			return
		}
		f, err := factory(config)
		if err != nil {
			p.L.Errorf("failed to instantiate filter, filter name: %s, config: %s", name, config)
			p.responseInstantiateFilterError(w, name)
			return
		}
		filters = append(filters, f)
	}
	var ctx = filter.NewContext(map[any]any{
		filter.RouteCtxKey{}:     rout,
		filter.ProvidersCtxKey{}: p.Config.providers,
		filter.FiltersCtxKey{}:   filters, // todo: 风险: 将 filters 通过 context 传入, 后续的 filter 都能拿到和调用其他 filter
		filter.DBCtxKey{}:        p.D,
		filter.LoggerCtxKey{}:    p.L.Sub(r.Header.Get("X-Request-Id")),
		filter.MutexCtxKey{}:     new(sync.Mutex),
	})
	for i := 0; i < len(filters); i++ {
		if reflect.TypeOf(filters[i]) == reverse_proxy.Type {
			_, _ = filters[i].(filter.RequestFilter).OnHttpRequest(ctx, w, r)
			return
		}
	}
}

func (p *provider) parseRoutesConfig() error {
	return p.parseConfig(p.Config.RoutesRef, "routes", &p.Config.Routes)
}

func (p *provider) parseProvidersConfig() error {
	if err := p.parseConfig(p.Config.ProvidersRef, "providers", &p.Config.providers); err != nil {
		return err
	}
	for i := 0; i < len(p.Config.providers); i++ {
		if err := p.Config.providers[i].LoadOpenapiSpec(); err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) parseConfig(ref, key string, i interface{}) error {
	data, err := os.ReadFile(ref)
	if err != nil {
		return err
	}
	var m = make(map[string]json.RawMessage)
	if err := yaml.Unmarshal(data, &m); err != nil {
		return err
	}
	data, ok := m[key]
	if !ok {
		return nil
	}
	return yaml.Unmarshal(data, i)
}

func (p *provider) responseNoSuchRoute(w http.ResponseWriter, path, method string) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "no such route",
		"path":   path,
		"method": method,
	})
}

func (p *provider) responseNoSuchFilter(w http.ResponseWriter, filterName string) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "no such filter",
		"filter": filterName,
	})
}

func (p *provider) responseInstantiateFilterError(w http.ResponseWriter, filterName string) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "failed to instantiate filter",
		"filter": filterName,
	})
}

type config struct {
	RoutesRef    string `json:"routesRef" yaml:"routesRef"`
	ProvidersRef string `json:"providersRef" yaml:"providersRef"`
	LogLevel     string `json:"logLevel" yaml:"logLevel"`
	providers    provider2.Providers
	Routes       route2.Routes
}

func (c *config) GetLogLevel() string {
	expr, start, end, err := strutil.FirstCustomExpression(c.LogLevel, "${", "}", func(s string) bool {
		return strings.HasPrefix(strings.TrimSpace(s), "env.")
	})
	if err != nil || start == end {
		return c.LogLevel
	}
	key := strings.TrimPrefix(expr, "env.")
	keys := strings.Split(key, ":")
	if len(keys) > 0 {
		key = keys[0]
	}
	env, ok := os.LookupEnv(key)
	if !ok {
		if len(keys) > 1 {
			return strings.Join(keys[1:], ":")
		}
		return "info"
	}
	return env
}
