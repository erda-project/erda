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

package uc

import (
	"reflect"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
	"github.com/erda-project/erda/pkg/jsonstore"
)

const (
	jsonStoreCacheTimeout = 60
	tokenRefreshMargin    = 60 * time.Second
)

type Config struct {
	FrontendURL  string `file:"frontend_url"`
	BackendHost  string `file:"host"`
	ClientID     string `file:"oauth_client_id"`
	ClientSecret string `file:"oauth_client_secret"`
	// Optional, only needed for authorization_code grant types
	RedirectURI string `file:"redirect_uri"`
}

type provider struct {
	Log    logs.Logger
	Config *Config
	mu     sync.Mutex

	// server token
	serverToken           *domain.OAuthToken
	serverTokenExpireTime time.Time

	// client token cache
	clientTokenCache jsonstore.JsonStore
}

func (p *provider) Init(_ servicehub.Context) error {
	clientTokenCache, err := jsonstore.New(
		jsonstore.UseMemStore(),
		jsonstore.UseTimeoutStore(jsonStoreCacheTimeout),
	)
	if err != nil {
		return err
	}
	p.clientTokenCache = clientTokenCache
	legacycontainer.Register[domain.OAuthTokenProvider](p)
	return nil
}

func init() {
	servicehub.Register("erda.core.user.oauth.uc", &servicehub.Spec{
		Services:   []string{"erda.core.user.oauth"},
		Types:      []reflect.Type{reflect.TypeOf((*domain.OAuthProvider)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
