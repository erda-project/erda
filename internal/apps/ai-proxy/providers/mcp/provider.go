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

package ai

import (
	"context"
	_ "embed"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	proxyapis "github.com/erda-project/erda/internal/apps/ai-proxy/apis"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/mcp"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	k8sconfig "github.com/erda-project/erda/pkg/k8sclient/config"
)

var (
	_ transport.Register = (*provider)(nil)
)

var (
	name         = "erda.app.mcp-proxy"
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.mcp-proxy.Server"},
		Summary:     "mcp-proxy server",
		Description: "Reverse proxy service between MCP servers and client applications",
		ConfigFunc:  func() interface{} { return new(config.Config) },
		Types:       []reflect.Type{providerType},
		Creator:     func() servicehub.Provider { return new(provider) },
	}
	trySetAuth = func(dao dao.DAO) transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				ctx = ctxhelper.InitCtxMapIfNeed(ctx)
				// check admin key first
				isAdmin, err := akutil.CheckAdmin(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if isAdmin {
					ctxhelper.PutIsAdmin(ctx, true)
					return h(ctx, req)
				}
				// try set clientId by ak
				clientToken, client, err := akutil.CheckAkOrToken(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if clientToken != nil {
					ctxhelper.PutClientToken(ctx, clientToken)
				}
				if client != nil {
					ctxhelper.PutClient(ctx, client)
					ctxhelper.PutClientId(ctx, client.Id)
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
	Config         *config.Config
	L              logs.Logger
	HTTP           mux.Mux                              `autowired:"gorilla-mux@ai"`
	GRPC           grpcserver.Interface                 `autowired:"grpc-server@ai"`
	Dao            dao.DAO                              `autowired:"erda.apps.ai-proxy.dao"`
	ClusterSvc     clusterpb.ClusterServiceServer       `autowired:"erda.core.clustermanager.cluster.ClusterService"`
	DynamicOpenapi dynamic.DynamicOpenapiRegisterServer `autowired:"erda.core.openapi.dynamic_register.DynamicOpenapiRegister"`

	richClientHandler *handler_rich_client.ClientHandler

	cacheManager cachetypes.Manager

	Routes []*router_define.Route
	Router *router_define.Router
}

func (p *provider) Init(ctx servicehub.Context) error {
	// config
	if err := p.Config.DoPost(); err != nil {
		return err
	}

	// load route configs
	yamlFile, err := router_define.LoadRoutesFromEmbeddedDir(p.Config.EmbedRoutesFS)
	if err != nil {
		return err
	}
	// create rout
	p.Routes = yamlFile.Routes
	p.Router = router_define.NewRouter()
	for _, route := range p.Routes {
		expandedRoutes, err := router_define.ExpandRoute(route)
		if err != nil {
			return err
		}
		for _, expandedRoute := range expandedRoutes {
			p.Router.AddRoute(expandedRoute)
			p.L.Infof("register route from yaml: %s", expandedRoute)
		}
	}

	// register gRPC and http handler
	proxyapis.RegisterMcpProxyManageAPI(p, p.Config.McpProxyPublicURL)

	// initialize cache manager
	p.cacheManager = cache.NewCacheManager(p.Dao, p.L, p.Config.IsMcpProxy)

	p.HTTP.Handle("/health", http.MethodGet, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	// reverse proxy to AI provider's server
	proxyapis.ServeAIProxyV2(p, p.HTTP)

	return nil
}

func (p *provider) Add(method, path string, h transhttp.HandlerFunc) {
	p.HTTP.Handle(path, method, http.HandlerFunc(h), mux.SetXRequestId, mux.CORS)
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	p.GRPC.RegisterService(desc, impl)
}

func (p *provider) Run(ctx context.Context) error {
	if !p.Config.IsMcpProxy {
		return nil
	}

	if !p.Config.McpScanConfig.Enable {
		p.L.Infof("mcp server scan is disable")
		return nil
	}

	return p.onLeader(ctx, func(ctx context.Context) {
		handler := handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL)

		clusters := strings.Split(p.Config.McpScanConfig.McpClusters, ",")
		p.L.Infof("listen mcp cluster list: %v", clusters)

		aggregator := mcp.NewAggregator(ctx, p.ClusterSvc, handler, p.L, p.Config.McpScanConfig.SyncClusterConfigInterval, clusters)
		if err := aggregator.Start(ctx); err != nil {
			logrus.Error(err)
			panic(err)
		}
	})
}

func (p *provider) onLeader(ctx context.Context, handle func(ctx context.Context)) error {
	ctx = apis.WithInternalClientContext(ctx, discover.SvcMCPProxy)

	cluster, err := p.ClusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{
		IdOrName: p.Config.DiceInfo.LocalClusterName,
	})
	if err != nil {
		return err
	}

	conf, err := k8sconfig.ParseManageConfigPb(cluster.Data.Name, cluster.Data.ManageConfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return err
	}

	id, _ := os.Hostname()
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "mcp-proxy-leader",
			Namespace: p.Config.DiceInfo.Namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				p.L.Info("I am the mcp proxy leader")
				handle(ctx)
			},
			OnStoppedLeading: func() {
				p.L.Info("stopping the mcp proxy leader")
			},
		},
	})
	return nil
}

func (p *provider) FindBestMatch(method, path string) router_define.RouteMatcher {
	return p.Router.FindBestMatch(method, path)
}

func (p *provider) GetDBClient() dao.DAO {
	return p.Dao
}

func (p *provider) GetRichClientHandler() *handler_rich_client.ClientHandler {
	return p.richClientHandler
}

func (p *provider) SetRichClientHandler(handler *handler_rich_client.ClientHandler) {
	p.richClientHandler = handler
}

func (p *provider) GetCacheManager() cachetypes.Manager {
	return p.cacheManager
}

func (p *provider) IsMcpProxy() bool {
	return true
}
