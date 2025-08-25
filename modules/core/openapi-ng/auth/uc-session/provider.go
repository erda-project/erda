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

package ucoauth

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/openapi-ng"
	openapiauth "github.com/erda-project/erda/modules/core/openapi-ng/auth"
	"github.com/erda-project/erda/modules/openapi/settings"
)

type config struct {
	Weight               int64         `file:"weight" default:"100"`
	RedirectAfterLogin   string        `file:"redirect_after_login"`
	ClientID             string        `file:"client_id"`
	UCAddr               string        `file:"uc_addr"`
	UCRedirectAddrs      []string      `file:"uc_redirect_addrs"`
	SessionCookieName    string        `file:"session_cookie_name"`
	SessionCookieDomains []string      `file:"session_cookie_domain"`
	CookieMaxAge         time.Duration `file:"cookie_max_age" default:"24h" desc:"max age of the cookie. optional."`
	// CookieSameSite default set to 2, which is `lax`, more options see https://github.com/golang/go/blob/619b419a4b1506bde1aa7e833898f2f67fd0e83e/src/net/http/cookie.go#L52-L57
	CookieSameSite int `file:"cookie_same_site" default:"2" desc:"indicates if cookie is SameSite. optional."`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Router   openapi.Interface        `autowired:"openapi-router"`
	Redis    *redis.Client            `autowired:"redis-client"`
	Settings settings.OpenapiSettings `autowired:"openapi-settings"`
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.Cfg.RedirectAfterLogin = strings.TrimLeft(p.Cfg.RedirectAfterLogin, "/")

	router := p.Router
	router.Add(http.MethodGet, "/api/openapi/login", p.LoginURL)
	router.Add("", "/api/openapi/logincb", p.LoginCallback)
	router.Add("", "/logincb", p.LoginCallback)
	router.Add(http.MethodPost, "/api/openapi/logout", p.Logout)
	router.Add(http.MethodPost, "/logout", p.Logout)
	p.addUserInfoAPI(router)
	return nil
}

var _ openapiauth.AutherLister = (*provider)(nil)

func (p *provider) Authers() []openapiauth.Auther {
	return []openapiauth.Auther{
		&loginChecker{p: p},
		&tryLoginChecker{p: p},
	}
}

func init() {
	servicehub.Register("openapi-auth-uc", &servicehub.Spec{
		Services:   []string{"openapi-auth-session"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
