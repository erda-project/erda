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
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	mcppb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
	"github.com/erda-project/erda/pkg/common/apis"
)

const Name = "erda.app.mcp-proxy"

type Config struct {
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`
}

type provider struct {
	Config *Config
	L      logs.Logger
	Dao    dao.DAO `autowired:"erda.apps.ai-proxy.dao"`

	reverseproxy.Interface `autowired:"erda.app.reverse-proxy"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.registerMcpProxyManageAPI()

	// initialize cache manager
	p.SetCacheManager(cache.NewCacheManager(p.Dao, p.L, true))

	p.ServeAIProxyV2()
	return nil
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), reverseproxy.TrySetAuth(p.Dao), permission.CheckMCPPerm)
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
