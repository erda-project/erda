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

package facade

import (
	"reflect"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
)

type Config struct {
}

type provider struct {
	Cfg *Config

	OAuthTokenProvider domain.OAuthTokenProvider `autowired:"erda.core.user.oauth"`
	CredStore          domain.CredentialStore    `autowired:"erda.core.user.credstore"`
	Identity           domain.Identity           `autowired:"erda.core.user.identity"`
	Bundle             *bundle.Bundle
}

func (p *provider) Init(_ servicehub.Context) error {
	p.Bundle = bundle.New(bundle.WithErdaServer(), bundle.WithDOP())
	legacycontainer.Register[domain.UserAuthFacade](p)
	return nil
}

func (p *provider) NewUserState() domain.UserAuthState {
	return &userState{
		state:              GetInit,
		oauthTokenProvider: p.OAuthTokenProvider,
		identity:           p.Identity,
		bundle:             p.Bundle,
		credStore:          p.CredStore,
	}
}

func init() {
	servicehub.Register("erda.core.user.auth.facade", &servicehub.Spec{
		Services:   []string{"erda.core.user.auth.facade"},
		Types:      []reflect.Type{reflect.TypeOf((*domain.UserAuthFacade)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
