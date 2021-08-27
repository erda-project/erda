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

package token

import (
	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/openapi-ng"
	openapiauth "github.com/erda-project/erda/modules/core/openapi-ng/auth"
	"github.com/erda-project/erda/modules/openapi/oauth2"
)

type config struct {
	Weight int64 `file:"weight"`
}

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	Router       openapi.Interface     `autowired:"openapi-router"`
	Redis        *redis.Client         `autowired:"redis-client"`
	Auth         openapiauth.Interface `autowired:"openapi-auth"`
	oauth2server *oauth2.OAuth2Server
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.oauth2server = oauth2.NewOAuth2Server()
	p.Auth.Register(p)

	router := p.Router
	router.Add("", "/oauth2/token", p.oauth2server.Token)
	router.Add("", "/oauth2/invalidate_token", p.oauth2server.InvalidateToken)
	router.Add("", "/oauth2/validate_token", p.oauth2server.ValidateToken)
	return nil
}

func init() {
	servicehub.Register("openapi-auth-token", &servicehub.Spec{
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
