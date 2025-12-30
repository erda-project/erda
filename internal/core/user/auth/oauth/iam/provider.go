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
	"reflect"

	"github.com/bluele/gcache"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
)

const (
	serverTokenCacheKey = "server-token"
)

type Config struct {
	FrontendURL  string `file:"frontend_url"`
	BackendHost  string `file:"host"`
	ClientID     string `file:"client_id"`
	ClientSecret string `file:"client_secret"`
	// Optional, only needed for authorization_code grant types
	RedirectURI string `file:"redirect_uri"`
}

type provider struct {
	Log    logs.Logger
	Config *Config

	serverTokenCache gcache.Cache
}

func (p *provider) Init(_ servicehub.Context) error {
	legacycontainer.Register[domain.OAuthTokenProvider](p)
	p.serverTokenCache = gcache.New(1).Build()
	return nil
}

func init() {
	servicehub.Register("erda.core.user.oauth.iam", &servicehub.Spec{
		Services:   []string{"erda.core.user.oauth"},
		Types:      []reflect.Type{reflect.TypeOf((*domain.OAuthProvider)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
