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

	"github.com/recallsong/go-utils/logs"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
)

type Config struct {
	CookieName           string   `file:"cookie_name"`
	SessionCookieDomains []string `file:"session_cookie_domain"`
	// CookieSameSite default set to 2, which is `lax`, more options see https://github.com/golang/go/blob/619b419a4b1506bde1aa7e833898f2f67fd0e83e/src/net/http/cookie.go#L52-L57
	CookieSameSite int `file:"cookie_same_site" default:"2" desc:"indicates if cookie is SameSite. optional."`
}

type provider struct {
	Log    logs.Logger
	Config *Config
}

func (p *provider) Init(_ servicehub.Context) error {
	legacycontainer.Register[domain.CredentialStore](p)
	return nil
}

func init() {
	servicehub.Register("erda.core.user.credstore.uc", &servicehub.Spec{
		Services:   []string{"erda.core.user.credstore"},
		Types:      []reflect.Type{reflect.TypeOf((*domain.CredentialStore)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
