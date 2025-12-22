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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy/aiproxytypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	etcdstore "github.com/erda-project/erda/pkg/jsonstore/etcd"
)

const Name = "erda.app.ai-proxy"

type Config struct {
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`

	LBStateStoreType      string        `file:"lb_state_store_type" env:"LB_STATE_STORE_TYPE" default:"memory"`
	LBStateStoreStickyTTL time.Duration `file:"lb_state_store_sticky_ttl" env:"LB_STATE_STORE_STICKY_TTL" default:"10m"`

	// Redis settings (standalone or sentinel via redis.UniversalOptions)
	RedisAddr          string `file:"redis_addr" env:"REDIS_ADDR"`
	RedisSentinelsAddr string `file:"redis_sentinels_addr" env:"REDIS_SENTINELS_ADDR"`
	RedisMasterName    string `file:"redis_master_name" env:"REDIS_MASTER_NAME"`
	RedisPassword      string `file:"redis_password" env:"REDIS_PASSWORD"`
	RedisDB            int    `file:"redis_db" env:"REDIS_DB" default:"0"`

	// Etcd settings
	EtcdEndpoints string `file:"etcd_endpoints" env:"ETCD_ENDPOINTS"`
	EtcdUsername  string `file:"etcd_username" env:"ETCD_USERNAME"`
	EtcdPassword  string `file:"etcd_password" env:"ETCD_PASSWORD"`
}

type provider struct {
	Config *Config
	L      logs.Logger
	Dao    dao.DAO `autowired:"erda.apps.ai-proxy.dao"`

	ReverseProxy reverseproxy.Interface `autowired:"erda.app.reverse-proxy"`

	cache cachetypes.Manager

	handlers           *aiproxytypes.Handlers
	ctxhelperFunctions []func(context.Context)

	etcdStore *etcdstore.Store
}

func (p *provider) Init(ctx servicehub.Context) error {
	// load templates
	templatesByType, err := template.LoadTemplatesFromEmbeddedFS(p.L, config.EmbedTemplatesFS)
	if err != nil {
		return err
	}

	// init lb state store
	if _, _, err := p.initPolicyGroupStateStore(); err != nil {
		return fmt.Errorf("init policy-group state store failed: %w", err)
	}
	// init policy group engine
	engine.SetEngine(engine.NewEngine(state_store.GetStore(), engine.WithStickyTTL(p.Config.LBStateStoreStickyTTL)))

	// initialize cache manager
	p.cache = cache.NewCacheManager(p.Dao, p.L, templatesByType, false)
	p.ReverseProxy.SetCacheManager(p.cache)

	// initialize token usage collector
	token_usage.InitUsageCollector(p.Dao)

	p.initHandlers(templatesByType)

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

func (p *provider) initPolicyGroupStateStore() (store state_store.LBStateStore, desc string, err error) {
	defer func() {
		if store != nil {
			state_store.SetStore(store)
			p.L.Infof("policy-group state store: %s", desc)
		}
	}()
	typ := strings.ToLower(strings.TrimSpace(p.Config.LBStateStoreType))
	switch typ {
	case "redis":
		return p.buildRedisStateStore()
	case "etcd":
		return p.buildEtcdStateStore()
	default:
		return state_store.NewMemoryStateStore(), "memory", nil
	}
}

func (p *provider) buildRedisStateStore() (state_store.LBStateStore, string, error) {
	opt := p.buildRedisOptions()
	if opt == nil {
		return nil, "", fmt.Errorf("redis options not configured")
	}
	store := state_store.NewRedisStateStoreUniversal(opt, "ai-proxy:lb")
	if store == nil {
		return nil, "", fmt.Errorf("redis client is nil")
	}
	desc := "redis"
	if len(opt.MasterName) > 0 && len(opt.Addrs) > 0 {
		desc = fmt.Sprintf("redis(sentinel=%s master=%s)", strings.Join(opt.Addrs, ","), opt.MasterName)
	} else if len(opt.Addrs) > 0 {
		desc = fmt.Sprintf("redis(%s)", strings.Join(opt.Addrs, ","))
	}
	return store, desc, nil
}

func (p *provider) buildEtcdStateStore() (state_store.LBStateStore, string, error) {
	opts := []etcdstore.OpOption{}
	if eps := splitAndTrim(p.Config.EtcdEndpoints); len(eps) > 0 {
		opts = append(opts, etcdstore.WithEndpoints(eps))
	}
	store, err := etcdstore.New(opts...)
	if err != nil {
		return nil, "", err
	}
	p.etcdStore = store
	cli := store.GetClient()
	desc := "etcd"
	if len(opts) > 0 {
		desc = fmt.Sprintf("etcd(%s)", strings.Join(splitAndTrim(p.Config.EtcdEndpoints), ","))
	}
	return state_store.NewEtcdStateStore(cli, "ai-proxy:lb"), desc, nil
}

func (p *provider) buildRedisOptions() *redis.UniversalOptions {
	sentinels := splitAndTrim(p.Config.RedisSentinelsAddr)
	addrs := splitAndTrim(p.Config.RedisAddr)
	// prefer sentinel when provided
	if len(sentinels) > 0 {
		return &redis.UniversalOptions{
			MasterName: p.Config.RedisMasterName,
			Addrs:      sentinels,
			Password:   p.Config.RedisPassword,
			DB:         p.Config.RedisDB,
		}
	}
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:6379"}
	}
	return &redis.UniversalOptions{
		Addrs:    addrs,
		Password: p.Config.RedisPassword,
		DB:       p.Config.RedisDB,
	}
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
	})
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
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
