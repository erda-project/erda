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
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
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
		ConfigFunc:  func() interface{} { return new(config) }, // todo: 启动时支持从初始配置文件或配置中心(Nacos, ETCD, MySQL)获取配置
		Types:       []reflect.Type{providerType},
		Creator:     func() servicehub.Provider { return new(provider) },
	}
	rootKeyAuth = transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if auth := transport.ContextHeader(ctx).Get("Authorization"); len(auth) == 0 || auth[0] != vars.ConcatBearer(os.Getenv(vars.EnvAIProxyRootKey)) {
				return nil, errors.New("Access denied to the admin API")
			}
			return h(ctx, req)
		}
	})
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Config         *config
	L              logs.Logger
	HTTP           httpserver.Router                    `autowired:"http-server@ai"`
	GRPC           grpcserver.Interface                 `autowired:"grpc-server@ai"`
	Dao            dao.DAO                              `autowired:"erda.apps.ai-proxy.dao"`
	DynamicOpenapi dynamic.DynamicOpenapiRegisterServer `autowired:"erda.core.openapi.dynamic_register.DynamicOpenapiRegister"`
	ErdaOpenapis   map[string]*url.URL
}

func (p *provider) Init(_ servicehub.Context) error {
	p.initLogger()
	if err := p.parseRoutesConfig(); err != nil {
		return errors.Wrap(err, "failed to parseRoutesConfig")
	}
	p.L.Infof("routes config:\n%s", strutil.TryGetYamlStr(p.Config.Routes))
	if err := p.parseProvidersConfig(); err != nil {
		return errors.Wrap(err, "failed to parseProvidersConfig")
	}
	p.L.Infof("providers config:\n%s", strutil.TryGetYamlStr(p.Config.Providers))
	if err := p.parsePlatformsConfig(); err != nil {
		return errors.Wrap(err, "failed to parsePlatformsConfig")
	}
	p.L.Infof("platforms config:\n%s", strutil.TryGetYamlStr(p.Config.Platforms))

	if p.Config.SelfURL == "" {
		p.Config.SelfURL = "http://ai-proxy:8081"
	}
	if selfURL, ok := os.LookupEnv("SELF_URL"); ok && len(selfURL) > 0 {
		p.Config.SelfURL = selfURL
	}
	if len(p.ErdaOpenapis) == 0 {
		p.ErdaOpenapis = make(map[string]*url.URL)
	}
	for i, plat := range p.Config.Platforms {
		if plat.Name == "" {
			return errors.Errorf("invalid platforms[%d] config, name is empty", i)
		}
		if plat.Openapi == "" {
			return errors.Errorf("the platform %s's openapi is invalid", plat.Name)
		}
		openapi, err := url.Parse(plat.Openapi)
		if err != nil {
			return errors.Wrapf(err, "faield to parse openapi, name: %s, openapi: %s", plat.Name, plat.Openapi)
		}
		p.ErdaOpenapis[plat.Name] = openapi
	}

	// 将静态文件中的 providers 同步到数据库
	ph := &handlers.ProviderHandler{Dao: p.Dao, Log: p.L.Sub("ProviderHandler")}
	if err := ph.Sync(context.Background(), p.Config.Providers); err != nil {
		return errors.Wrap(err, "failed to sync providers from static config")
	}

	// register gRPC and http handler
	pb.RegisterAccessImp(p, &handlers.AccessHandler{Dao: p.Dao, Log: p.L.Sub("AccessHandler")}, apis.Options())
	pb.RegisterChatLogsImp(p, &handlers.ChatLogsHandler{Dao: p.Dao, Log: p.L.Sub("ChatLogsHandler")}, apis.Options())
	pb.RegisterCredentialsImp(p, &handlers.CredentialsHandler{Dao: p.Dao, Log: p.L.Sub("CredentialHandler")}, apis.Options(), rootKeyAuth)
	pb.RegisterModelsImp(p, &handlers.ModelsHandler{Dao: p.Dao, Log: p.L.Sub("ModelsHandler")}, apis.Options())
	pb.RegisterAIProviderImp(p, ph, apis.Options(), rootKeyAuth)
	pb.RegisterSessionsImp(p, &handlers.SessionsHandler{Dao: p.Dao, Log: p.L.Sub("SessionsHandler")}, apis.Options())

	// ai-proxy prometheus metrics
	p.HTTP.Any("/metrics", promhttp.Handler())
	// reverse proxy to AI provider's server
	p.HTTP.Any("/**", p)

	// open APIs on Erda
	if p.Config.OpenOnErda {
		if err := p.openAPIsOnErda(); err != nil {
			p.L.Errorf("failed to open APIs on Erda, err: %v", err)
			// TODO: return err in next PR
			//return err
		}
	}

	return nil
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Config.Routes.FindRoute(WrapRequest(r, SetXRequestId)).
		HandlerWith(
			context.Background(),
			reverseproxy.LoggerCtxKey{}, p.L.Sub(r.Header.Get("X-Request-Id")),
			reverseproxy.MutexCtxKey{}, new(sync.Mutex),
			reverseproxy.CtxKeyMap{}, new(sync.Map),
			vars.CtxKeyDAO{}, p.Dao,
			vars.CtxKeyErdaOpenapi{}, p.ErdaOpenapis,
		).
		ServeHTTP(w, r)
}

func (p *provider) Add(method, path string, handler transhttp.HandlerFunc) {
	if p.HTTP != nil {
		if err := p.HTTP.Add(method, path, handler,
			httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs),
			interceptors.CORS(true),
		); err != nil {
			p.L.Fatalf("failed to %T.Add(%s, %s, %T), err: %v", p, method, path, handler, err)
		}
	}
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	if p.GRPC != nil {
		p.GRPC.RegisterService(desc, impl)
	}
}

// openAPIsOnErda opens the ai-proxy APIs on Erda
func (p *provider) openAPIsOnErda() error {
	auth := &common.APIAuth{
		CheckLogin: true,
		CheckToken: true,
	}

	// register openai APis
	for _, rout := range p.Config.Routes {
		if _, err := p.DynamicOpenapi.Register(context.Background(), &dynamic.API{
			Upstream:    p.Config.SelfURL,
			Method:      rout.Method,
			Path:        path.Join("/api/ai-proxy/openai", rout.Path),
			BackendPath: rout.Path,
			Auth:        auth,
		}); err != nil {
			return err
		}
	}

	// register admin APIs
	for _, api := range handlers.APIs {
		api.Upstream = p.Config.SelfURL
		api.Path = api.BackendPath
		api.Auth = auth
		if _, err := p.DynamicOpenapi.Register(context.Background(), api); err != nil {
			return err
		}
	}

	return nil
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
		if level, err := logrus.ParseLevel(p.Config.LogLevel); err == nil {
			p.L.Infof("logger level: %s", p.Config.LogLevel)
			logger.SetLevel(level)
		} else {
			p.L.Infof("failed to parse logger level from config, set it as %s", logrus.InfoLevel.String())
			logger.SetLevel(logrus.InfoLevel)
		}
		l.Entry = logrus.NewEntry(logger)
		return
	}
	p.L.Infof("logger level: %s", p.Config.LogLevel)
	if err := p.L.SetLevel(p.Config.LogLevel); err != nil {
		p.L.Infof("failed to set logger level from config, set it as %s", logrus.InfoLevel.String())
		_ = p.L.SetLevel(logrus.InfoLevel.String())
	}
}

func (p *provider) parseRoutesConfig() error {
	if err := p.parseConfig(p.Config.RoutesRef, "routes", &p.Config.Routes); err != nil {
		return err
	}
	return p.Config.Routes.Validate()
}

func (p *provider) parseProvidersConfig() error {
	return p.parseConfig(p.Config.ProvidersRef, "providers", &p.Config.Providers)
}

func (p *provider) parsePlatformsConfig() error {
	return p.parseConfig(p.Config.PlatformsRef, "platforms", &p.Config.Platforms)
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
	RoutesRef    string              `json:"routesRef" yaml:"routesRef"`
	ProvidersRef string              `json:"providersRef" yaml:"providersRef"`
	PlatformsRef string              `json:"platformsRef" yaml:"platformsRef"`
	LogLevel     string              `json:"logLevel" yaml:"logLevel"`
	Exporter     configPromExporter  `json:"exporter" yaml:"exporter"`
	SelfURL      string              `json:"selfURL" yaml:"selfURL"`
	OpenOnErda   bool                `json:"openOnErda" yaml:"openOnErda"`
	Routes       route2.Routes       `json:"-" yaml:"-"`
	Providers    provider2.Providers `json:"-" yaml:"-"`
	Platforms    []*Platform         `json:"-" yaml:"-"`
}

type configPromExporter struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Subsystem string `json:"subsystem" yaml:"subsystem"`
	Name      string `json:"name" yaml:"name"`
}

func WrapRequest(r *http.Request, wraps ...func(*http.Request)) *http.Request {
	for _, wrap := range wraps {
		wrap(r)
	}
	return r
}

func SetXRequestId(r *http.Request) {
	if id := r.Header.Get("X-Request-Id"); id == "" {
		r.Header.Set("X-Request-Id", strings.ReplaceAll(uuid.NewString(), "-", ""))
	}
}

type Platform struct {
	Name        string `json:"name" yaml:"name"`
	Openapi     string `json:"openapi" yaml:"openapi"`
	Description string `json:"description" yaml:"description"`
}
