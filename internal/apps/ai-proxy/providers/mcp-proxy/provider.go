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

package mcp_proxy

import (
	"context"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	mcppb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	mcipb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_config_instance/pb"
	mtpb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_template/pb"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server_config_instance"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/mcp"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	k8sconfig "github.com/erda-project/erda/pkg/k8sclient/config"
)

const Name = "erda.app.mcp-proxy"

type DiceInfo struct {
	LocalClusterName string `file:"local_cluster_name" env:"DICE_CLUSTER_NAME"`
	Namespace        string `file:"namespace" env:"DICE_NAMESPACE"`
}

type McpScanConfig struct {
	Enable                    bool          `file:"enable" env:"MCP_SCAN_ENABLE"`
	McpClusters               string        `file:"mcp_clusters" default:"" env:"MCP_CLUSTERS"`
	SyncClusterConfigInterval time.Duration `file:"sync_cluster_config_interval" default:"10m" env:"SYNC_CLUSTER_CONFIG_INTERVAL"`
}

type Config struct {
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`

	McpScanConfig McpScanConfig `file:"mcp_scan_config"`
	DiceInfo      DiceInfo      `file:"dice_info"`
}

type provider struct {
	Config *Config
	L      logs.Logger
	Dao    dao.DAO `autowired:"erda.apps.ai-proxy.dao"`

	ClusterSvc             clusterpb.ClusterServiceServer `autowired:"erda.core.clustermanager.cluster.ClusterService" optional:"true"`
	reverseproxy.Interface `autowired:"erda.app.reverse-proxy"`

	cache cachetypes.Manager
}

func (p *provider) Init(ctx servicehub.Context) error {
	// initialize cache manager
	p.cache = cache.NewCacheManager(p.Dao, p.L, nil, true)
	p.SetCacheManager(p.cache)

	p.registerMcpProxyManageAPI()

	// initialize cache manager
	p.SetCacheManager(cache.NewCacheManager(p.Dao, p.L, nil, true))

	p.ServeReverseProxyV2(reverseproxy.WithTransport(transports.NewMcpTransport()))
	return nil
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), reverseproxy.TrySetAuth(p.cache), permission.CheckMCPPerm)

	mtpb.RegisterMCPServerTemplateServiceImp(p, handler_mcp_server_template.NewMcpTemplateHandler(p.Dao, p.L), apis.Options(), reverseproxy.TrySetAuth(p.cache), permission.CheckMcpTemplatePerm)

	mcipb.RegisterMCPServerConfigInstanceServiceImp(p, handler_mcp_server_config_instance.NewMCPConfigInstanceHandler(p.Dao, p.L), reverseproxy.TrySetAuth(p.cache), permission.CheckMcpConfigInstancePerm)
}

func (p *provider) Run(ctx context.Context) error {
	if !p.Config.McpScanConfig.Enable {
		p.L.Info("mcp proxy mcp server discovery is disabled")
		return nil
	}
	for {
		err := p.onLeader(ctx, func(ctx context.Context) {
			handler := handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL)

			clusters := strings.Split(p.Config.McpScanConfig.McpClusters, ",")
			p.L.Infof("listen mcp cluster list: %v", clusters)

			aggregator := mcp.NewAggregator(ctx, p.ClusterSvc, handler, p.L, p.Config.McpScanConfig.SyncClusterConfigInterval, clusters, p.cache)
			if err := aggregator.Start(ctx); err != nil {
				logrus.Error(err)
				panic(err)
			}
		})
		if err != nil {
			p.L.Errorf("leader error: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
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

	ctx, cancelFunc := context.WithCancel(ctx)

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
				cancelFunc()
			},
		},
	})
	return nil
}

func init() {
	servicehub.Register(Name, &servicehub.Spec{
		Services:    []string{"erda.app.mcp-proxy.Server"},
		Summary:     "mcp-proxy server",
		Description: "mcp proxy service between mcp servers and client applications",
		ConfigFunc:  func() interface{} { return new(Config) },
		Types:       []reflect.Type{reflect.TypeOf((*provider)(nil))},
		Creator:     func() servicehub.Provider { return new(provider) },
	})
}
