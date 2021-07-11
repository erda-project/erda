// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package kratos

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth/user"
	"github.com/erda-project/erda/pkg/ucauth"
)

type config struct {
	LoginURL      string `file:"login_url"`
	SessionKey    string `file:"session_key" default:"ory_kratos_session"`
	OryKratosAddr string `file:"ory_kratos_addr" default:"kratos:4433" env:"KRATOS_ADDR"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
	bdl *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithCMDB())
	return nil
}

func (p *provider) Name() string { return "kratos" }
func (p *provider) Match(r *http.Request) (auth.AuthChecker, bool) {
	session, err := r.Cookie(p.Cfg.SessionKey)
	if err != nil || len(session.Value) <= 0 {
		return nil, false
	}
	return func(r *http.Request) (*auth.CheckResult, error) {
		// get token from session
		client := p.getClient()
		info, err := client.GetUserInfo(ucauth.OAuthToken{AccessToken: session.Value})
		if err != nil || !info.Enabled || len(info.ID) == 0 {
			return &auth.CheckResult{Success: false}, nil
		}
		info.Token = session.Value

		// check org
		ok, orgID, err := user.CheckOrg(p.bdl, r, string(info.ID))
		if err != nil {
			return &auth.CheckResult{Success: false}, nil
		}
		if !ok {
			return &auth.CheckResult{Success: false}, nil
		}

		return &auth.CheckResult{
			Success: true,
			Data: &userInfo{
				session: session.Value,
				info:    &info,
				orgID:   orgID,
			},
		}, nil
	}, true
}

func (p *provider) getClient() *ucauth.UCUserAuth {
	return ucauth.NewUCUserAuth("", p.Cfg.OryKratosAddr, "", "", "")
}

func (p *provider) RegisterHandler(add func(method, path string, h http.HandlerFunc)) {
	add(http.MethodGet, "/api/openapi/login", p.GetLoginURL)
}

func (p *provider) GetLoginURL(rw http.ResponseWriter, r *http.Request) {
	body, _ := json.Marshal(struct {
		URL string `json:"url"`
	}{
		URL: p.Cfg.LoginURL,
	})
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func init() {
	servicehub.Register("openapi-auth-kratos", &servicehub.Spec{
		Services:   []string{"openapi-auth-kratos"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
