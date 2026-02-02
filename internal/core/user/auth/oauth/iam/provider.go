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

package iam

import (
	"time"

	"github.com/bluele/gcache"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
)

const (
	serverTokenCacheKey    = "server_token"
	userTokenCachePrefix   = "user:token:"
	defaultEarlyExpireRate = 0.8
)

type Config struct {
	FrontendURL  string `file:"frontend_url"`
	BackendHost  string `file:"host"`
	ClientID     string `file:"client_id"`
	ClientSecret string `file:"client_secret"`
	// Optional, only needed for authorization_code grant types
	RedirectURI string `file:"redirect_uri"`
	// token cache config
	TokenCacheSize            int           `file:"token_cache_size" default:"20000"`
	TokenCacheEarlyExpireRate float64       `file:"token_cache_early_expire_rate" default:"0.8"`
	ServerTokenCacheEnabled   bool          `file:"server_token_cache_enabled" default:"true"`
	UserTokenCacheEnabled     bool          `file:"user_token_cache_enabled" default:"true"`
	UserTokenCacheExpireTime  time.Duration `file:"user_token_cache_expire_time"`
	OAuthTokenCacheSecret     string        `file:"oauth_token_cache_secret" default:"secret"`
}

type provider struct {
	Register transport.Register `autowired:"service-register" required:"true"`
	Log      logs.Logger
	Config   *Config

	tokenCache gcache.Cache
}

func (p *provider) Init(_ servicehub.Context) error {
	// auto fix token cache early expire rate
	rate := p.Config.TokenCacheEarlyExpireRate
	if rate <= 0 || rate >= 1 {
		rate = defaultEarlyExpireRate
		p.Log.Warnf("illegal token cache early expire rate, use default: %v", rate)
	}
	p.Config.TokenCacheEarlyExpireRate = rate
	p.tokenCache = gcache.New(p.Config.TokenCacheSize).LRU().Build()

	if p.Register != nil {
		pb.RegisterUserOAuthServiceImp(p.Register, p)
		pb.RegisterUserOAuthSessionServiceImp(p.Register, p)
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.user.oauth.iam", &servicehub.Spec{
		Services:   pb.ServiceNames(),
		Types:      pb.Types(),
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
