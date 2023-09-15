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
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/common/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_session"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	provider2 "github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	route2 "github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
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
	trySetAuth = func(dao dao.DAO) transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				// check admin key first
				adminKey := vars.TrimBearer(apis.GetHeader(ctx, httputil.HeaderKeyAuthorization))
				if len(adminKey) > 0 && adminKey == os.Getenv(vars.EnvAIProxyAdminAuthKey) {
					ctx = context.WithValue(ctx, vars.CtxKeyIsAdmin{}, true)
					return h(ctx, req)
				}
				// try set clientId by ak
				clientId, err := akutil.CheckAkOrToken(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if clientId != "" {
					ctx = context.WithValue(ctx, vars.CtxKeyClientId{}, clientId)
				}
				return h(ctx, req)
			}
		})
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Config         *config
	L              logs.Logger
	HTTP           mux.Mux                              `autowired:"gorilla-mux@ai"`
	GRPC           grpcserver.Interface                 `autowired:"grpc-server@ai"`
	Dao            dao.DAO                              `autowired:"erda.apps.ai-proxy.dao"`
	DynamicOpenapi dynamic.DynamicOpenapiRegisterServer `autowired:"erda.core.openapi.dynamic_register.DynamicOpenapiRegister"`
	ErdaOpenapis   map[string]*url.URL
}

func (p *provider) Init(ctx servicehub.Context) error {
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

	// register gRPC and http handler
	encoderOpts := mux.InfraEncoderOpt(mux.InfraCORS)
	clientpb.RegisterClientServiceImp(p, &handler_client.ClientHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientPerm)
	modelproviderpb.RegisterModelProviderServiceImp(p, &handler_model_provider.ModelProviderHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckModelProviderPerm)
	modelpb.RegisterModelServiceImp(p, &handler_model.ModelHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckModelPerm)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(p, &handler_client_model_relation.ClientModelRelationHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientModelRelationPerm)
	promptpb.RegisterPromptServiceImp(p, &handler_prompt.PromptHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckPromptPerm)
	sessionpb.RegisterSessionServiceImp(p, &handler_session.SessionHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckSessionPerm)
	clienttokenpb.RegisterClientTokenServiceImp(p, &handler_client_token.ClientTokenHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientTokenPerm)

	// ai-proxy prometheus metrics
	p.HTTP.Handle("/metrics", http.MethodGet, promhttp.Handler())
	p.HTTP.Handle("/health", http.MethodGet, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	// reverse proxy to AI provider's server
	p.ServeAIProxy()

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

func (p *provider) ServeAIProxy() {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		p.Config.Routes.FindRoute(r).HandlerWith(
			context.Background(),
			reverseproxy.LoggerCtxKey{}, p.L.Sub(r.Header.Get("X-Request-Id")),
			reverseproxy.MutexCtxKey{}, new(sync.Mutex),
			reverseproxy.CtxKeyMap{}, new(sync.Map),
			vars.CtxKeyDAO{}, p.Dao,
			vars.CtxKeyErdaOpenapi{}, p.ErdaOpenapis,
		).ServeHTTP(w, r)
	}
	p.HTTP.HandlePrefix("/", "*", f, mux.SetXRequestId, mux.CORS)
}

func (p *provider) Add(method, path string, h transhttp.HandlerFunc) {
	p.HTTP.Handle(path, method, http.HandlerFunc(h))
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	p.GRPC.RegisterService(desc, impl)
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

type Platform struct {
	Name        string `json:"name" yaml:"name"`
	Openapi     string `json:"openapi" yaml:"openapi"`
	Description string `json:"description" yaml:"description"`
}
