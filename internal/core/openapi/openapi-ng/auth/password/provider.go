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

package password

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng"
	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/common"
	"github.com/erda-project/erda/internal/core/openapi/settings"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	oatuh2TokenStore "github.com/erda-project/erda/pkg/oauth2/clientstore/mysqlclientstore"
)

type config struct {
	Weight               int64  `file:"weight" default:"50"`
	AccessTokenExpiredIn string `file:"access_token_expred_in" default:"1h"`
}

// +provider
type provider struct {
	Cfg       *config
	Log       logs.Logger
	Router    openapi.Interface `autowired:"openapi-router"`
	Redis     *redis.Client     `autowired:"redis-client"`
	Org       org.Interface
	Settings  settings.OpenapiSettings `autowired:"openapi-settings"`
	CredStore domain.CredentialStore   `autowired:"erda.core.user.credstore"`
	Identity  domain.Identity          `autowired:"erda.core.user.identity"`
	UserAuth  domain.UserAuthFacade    `autowired:"erda.core.user.auth.facade"`
	Bdl       *bundle.Bundle
	// openapi token
	openapiToken *oatuh2TokenStore.ClientStoreItem
}

func (p *provider) Init(_ servicehub.Context) error {
	p.Bdl = bundle.New(bundle.WithErdaServer())
	// router
	p.Router.Add(http.MethodPost, "/login", p.Login)

	// get openapi oauth token pair
	repo, err := NewOAuth2Repo()
	if err != nil {
		return err
	}
	p.openapiToken, err = repo.GetOrCreateOpenAPIClient()
	if err != nil {
		return err
	}
	return nil
}

func (p *provider) Login(rw http.ResponseWriter, r *http.Request) {
	var loginParams LoginParams
	switch contentType := strings.ToLower(r.Header.Get("content-type")); {
	case strings.HasPrefix(contentType, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&loginParams); err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}
	default:
		_ = r.ParseForm()
		loginParams.Username = r.FormValue("username")
		loginParams.Password = r.FormValue("password")
	}

	if loginParams.Username == "" || loginParams.Password == "" {
		http.Error(rw, "username or password is required", http.StatusUnauthorized)
		return
	}

	user := p.UserAuth.NewState()
	err := user.PwdLogin(loginParams.Username, loginParams.Password)
	if err != nil {
		err := fmt.Errorf("failed to PwdLogin: %v", err)
		p.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}
	info, authr := user.GetInfo(r)
	if authr.Code != domain.AuthSuccess {
		err := fmt.Errorf("failed to GetInfo: %v", authr.Detail)
		p.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := p.Bdl.GetOAuth2Token(apistructs.OAuth2TokenGetRequest{
		ClientID:     p.openapiToken.ID,
		ClientSecret: p.openapiToken.Secret,
		Payload: apistructs.OAuth2TokenPayload{
			AccessTokenExpiredIn: p.Cfg.AccessTokenExpiredIn,
			AllowAccessAllAPIs:   true,
			Metadata: map[string]string{
				"User-ID": string(info.ID),
			},
		},
	})
	if err != nil {
		err := fmt.Errorf("failed to get oauth2 token with client id %s, %v", p.openapiToken.ID, err)
		p.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	common.ResponseJSON(rw, &LoginResponse{
		User:  &info,
		Token: token,
	})
}

var _ openapiauth.AutherLister = (*provider)(nil)

func (p *provider) Authers() []openapiauth.Auther {
	return []openapiauth.Auther{p}
}

func init() {
	servicehub.Register("openapi-auth-password", &servicehub.Spec{
		Services:     []string{"openapi-auth-password"},
		Dependencies: []string{"openapi-auth-session"}, // to check session
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
