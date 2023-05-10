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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	provider2 "github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	route2 "github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/reverseproxy"
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
	L        logs.Logger
	Config   *config
	AiServer httpserver.Router `autowired:"http-server@ai"`
	D        *gorm.DB          `autowired:"mysql-gorm.v2-client"`
}

func (p *provider) Init(_ servicehub.Context) error {
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

	// prepare handlers
	for i := 0; i < len(p.Config.Routes); i++ {
		rout := p.Config.Routes[i]

		// validate every route config
		if err := rout.Validate(); err != nil {
			return errors.Wrapf(err, "rout %d is invalid", i)
		}

		// find provider to router to
		to := string(rout.Router.To)
		if !strings.HasPrefix(to, "__") || !strings.HasSuffix(to, "__") {
			prov, ok := p.Config.providers.FindProvider(to, rout.Router.InstanceId)
			if !ok {
				return errors.Errorf("no such provider routes[%d].Route.To: %s", i, p.Config.Routes[i].Router.To)
			}
			rout.With(reverseproxy.ProviderCtxKey{}, prov)
		}

		// prepare reverse proxy handler with contexts
		rout.With(
			reverseproxy.DBCtxKey{}, p.D,
			reverseproxy.LoggerCtxKey{}, p.L,
		)
	}

	// ai-proxy prometheus metrics
	p.AiServer.Any("/metrics", promhttp.Handler())
	// reverse proxy to AI provider's server
	p.AiServer.Any("/**", p)
	return nil
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Config.Routes.FindRoute(r.URL.Path, r.Method, r.Header).
		With(
			reverseproxy.LoggerCtxKey{}, p.L.Sub(r.Header.Get("X-Request-Id")),
			reverseproxy.MutexCtxKey{}, new(sync.Mutex),
		).
		ServeHTTP(w, r)
}

func (p *provider) parseRoutesConfig() error {
	return p.parseConfig(p.Config.RoutesRef, "routes", &p.Config.Routes)
}

func (p *provider) parseProvidersConfig() error {
	return p.parseConfig(p.Config.ProvidersRef, "providers", &p.Config.providers)
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
	RoutesRef    string             `json:"routesRef" yaml:"routesRef"`
	ProvidersRef string             `json:"providersRef" yaml:"providersRef"`
	LogLevel     string             `json:"logLevel" yaml:"logLevel"`
	Exporter     configPromExporter `json:"exporter" yaml:"exporter"`
	providers    provider2.Providers
	Routes       route2.Routes
}

type configPromExporter struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Subsystem string `json:"subsystem" yaml:"subsystem"`
	Name      string `json:"name" yaml:"name"`
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
