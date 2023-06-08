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
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes"
	provider2 "github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	route2 "github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

// issue: 创建 session 时没有校验 org

var (
	_ transport.Register = (*provider)(nil)
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
	Config  *config
	L       logs.Logger
	HTTP    httpserver.Router      `autowired:"http-server@ai"`
	GRPC    grpcserver.Interface   `autowired:"grpc-server@ai"`
	Dao     dao.DAO                `autowired:"erda.apps.ai-proxy.dao"`
	OrgSvc  orgpb.OrgServiceServer `autowired:"erda.core.org.OrgService"`
	Openapi routes.Register        `autowired:"openapi-dynamic-register.client"`
}

func (p *provider) Init(_ servicehub.Context) error {
	p.initLogger()
	if err := p.parseRoutesConfig(); err != nil {
		return errors.Wrap(err, "failed to parseRoutesConfig")
	}
	p.L.Infof("providers config:\n%s", strutil.TryGetYamlStr(p.Config.providers))
	if err := p.parseProvidersConfig(); err != nil {
		return errors.Wrap(err, "failed to parseProvidersConfig")
	}
	p.L.Infof("routes config:\n%s", strutil.TryGetYamlStr(p.Config.Routes))

	if p.Config.SelfURL == "" {
		p.Config.SelfURL = "http://ai-proxy:8081"
	}
	if selfURL, ok := os.LookupEnv("SELF_URL"); ok && len(selfURL) > 0 {
		p.Config.SelfURL = selfURL
	}

	// prepare handlers
	for i := 0; i < len(p.Config.Routes); i++ {
		rout := p.Config.Routes[i]

		// validate every route config
		if err := rout.Validate(); err != nil {
			return errors.Wrapf(err, "rout %d is invalid", i)
		}

		// find provider to router to
		if to := string(rout.Router.To); !strings.HasPrefix(to, "__") || !strings.HasSuffix(to, "__") {
			prov, ok := p.Config.providers.FindProvider(to, rout.Router.InstanceId)
			if !ok {
				return errors.Errorf("no such provider routes[%d].Route.To: %s", i, p.Config.Routes[i].Router.To)
			}
			rout.Provider = prov
		}

		// register to erda openapi
		if err := p.Openapi.Register(&routes.APIProxy{
			Method:      rout.Method,
			Path:        path.Join("/api/ai-proxy", rout.Path),
			ServiceURL:  p.Config.SelfURL,
			BackendPath: rout.Path,
			Auth: &common.APIAuth{
				CheckLogin: true,
				CheckToken: true,
			},
		}); err != nil {
			return err
		}
	}

	// register gRPC and http handler
	pb.RegisterChatLogsImp(p, &handlers.ChatLogsHandler{Dao: p.Dao, Log: p.L.Sub("ChatLogsHandler")}, apis.Options())
	pb.RegisterModelsImp(p, &handlers.ModelsHandler{Dao: p.Dao, Log: p.L.Sub("ModelsHandler")}, apis.Options())
	pb.RegisterSessionsImp(p, &handlers.SessionsHandler{Dao: p.Dao, Log: p.L.Sub("SessionsHandler")}, apis.Options())

	// ai-proxy prometheus metrics
	p.HTTP.Any("/metrics", promhttp.Handler())
	// reverse proxy to AI provider's server
	p.HTTP.Any("/**", p)
	return nil
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Config.Routes.FindRoute(r.URL.Path, r.Method, r.Header).
		HandlerWith(
			context.Background(),
			reverseproxy.LoggerCtxKey{}, p.L.Sub(r.Header.Get("X-Request-Id")),
			reverseproxy.MutexCtxKey{}, new(sync.Mutex),
			vars.CtxKeyOrgSvc{}, p.OrgSvc,
			vars.CtxKeyDAO{}, p.Dao,
		).
		ServeHTTP(w, r)
}

func (p *provider) Add(method, path string, handler transhttp.HandlerFunc) {
	if p.HTTP != nil {
		if err := p.HTTP.Add(method, path, handler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs)); err != nil {
			p.L.Fatalf("failed to %T.Add(%s, %s, %T), err: %v", p, method, path, handler, err)
		}
	}
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	if p.GRPC != nil {
		p.GRPC.RegisterService(desc, impl)
	}
}

func (p *provider) initLogger() {
	if l, ok := p.L.(*logrusx.Logger); ok {
		var logger = logrus.New()
		var formatter = &logrus.TextFormatter{
			ForceColors:      true,
			DisableQuote:     true,
			TimestampFormat:  time.RFC3339,
			DisableSorting:   true,
			QuoteEmptyFields: true,
		}
		p.L.Infof("logger formatter: %+v", formatter)
		logger.SetFormatter(formatter)
		if level, err := logrus.ParseLevel(p.Config.GetLogLevel()); err == nil {
			p.L.Infof("logger level: %s", p.Config.GetLogLevel())
			logger.SetLevel(level)
		} else {
			p.L.Infof("failed to parse logger level from config, set it as %s", logrus.InfoLevel.String())
			logger.SetLevel(logrus.InfoLevel)
		}
		l.Entry = logrus.NewEntry(logger)
		return
	}
	p.L.Infof("logger level: %s", p.Config.GetLogLevel())
	if err := p.L.SetLevel(p.Config.GetLogLevel()); err != nil {
		p.L.Infof("failed to set logger level from config, set it as %s", logrus.InfoLevel.String())
		_ = p.L.SetLevel(logrus.InfoLevel.String())
	}
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
	w.Header().Set("Server", "AI Service on Erda")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "no such route",
		"path":   path,
		"method": method,
	})
}

func (p *provider) responseNoSuchFilter(w http.ResponseWriter, filterName string) {
	w.Header().Set("Server", "AI Service on Erda")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "no such filter",
		"filter": filterName,
	})
}

func (p *provider) responseInstantiateFilterError(w http.ResponseWriter, filterName string) {
	w.Header().Set("Server", "AI Service on Erda")
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
	SelfURL      string             `json:"selfURL" yaml:"selfURL"`
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
		return logrus.InfoLevel.String()
	}
	return env
}
