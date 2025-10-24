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
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy/aiproxytypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
)

const Name = "erda.app.ai-proxy"

type Config struct {
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`
}

type provider struct {
	Config *Config
	L      logs.Logger
	Dao    dao.DAO `autowired:"erda.apps.ai-proxy.dao"`

	ReverseProxy reverseproxy.Interface `autowired:"erda.app.reverse-proxy"`

	cache cachetypes.Manager

	handlers           *aiproxytypes.Handlers
	ctxhelperFunctions []func(context.Context)
}

func (p *provider) Init(ctx servicehub.Context) error {

	// initialize cache manager
	p.cache = cache.NewCacheManager(p.Dao, p.L, false)
	p.ReverseProxy.SetCacheManager(p.cache)

	// initialize token usage collector
	token_usage.InitUsageCollector(p.Dao)

	p.initHandlers()

	p.registerAIProxyManageAPI()
	p.registerMcpProxyManageAPI()

	// custom health check
	p.ReverseProxy.SetHealthCheckAPI(p.HealthCheckAPI())

	p.ReverseProxy.ServeReverseProxyV2(reverseproxy.WithCtxHelperItems(
		func(ctx context.Context) {
			ctxhelper.PutAIProxyHandlers(ctx, p.handlers)
		},
	))

	return nil
}

func init() {
	servicehub.Register(Name, &servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.Server"},
		Summary:     "ai-proxy server",
		Description: "Reverse proxy service between AI vendors and client applications, providing a cut-through for service access",
		ConfigFunc:  func() interface{} { return new(Config) },
		Types:       []reflect.Type{reflect.TypeOf((*provider)(nil))},
		Creator:     func() servicehub.Provider { return new(provider) },
	})
}
