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

package fixedusers

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
)

type userSpec struct {
	ID       int64   `file:"id"`
	Password string  `file:"password"`
	OrgIDs   []int64 `file:"org_ids"`
}

type config struct {
	Users map[string]userSpec `file:"users"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) Name() string { return "fixed-users" }

func (p *provider) Match(r *http.Request) (auth.AuthChecker, bool) {
	const prefix = "Basic "
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) <= len(prefix) {
		return nil, false
	}
	if !strings.EqualFold(authHeader[0:len(prefix)], prefix) {
		return nil, false
	}
	authHeader = strings.TrimSpace(authHeader[len(prefix):])
	return func(r *http.Request) (*auth.CheckResult, error) {
		userpwd, err := base64.StdEncoding.DecodeString(authHeader)
		if err != nil {
			return &auth.CheckResult{Success: false}, nil
		}
		splitted := strings.SplitN(string(userpwd), ":", 2)
		if len(splitted) != 2 {
			return &auth.CheckResult{Success: false}, nil
		}
		return p.checkUsernameAndPassword(splitted[0], splitted[1], r)
	}, true
}

func (p *provider) checkUsernameAndPassword(username, password string, r *http.Request) (*auth.CheckResult, error) {
	info, ok := p.Cfg.Users[username]
	if !ok || info.Password != password {
		return &auth.CheckResult{Success: false}, nil
	}

	// TODO: check org

	return &auth.CheckResult{
		Success: true,
		Data: &userInfo{
			id:    strconv.FormatInt(info.ID, 10),
			orgID: 0,
		},
	}, nil
}

func (p *provider) RegisterHandler(add func(method, path string, h http.HandlerFunc)) {
	add(http.MethodGet, "/login", p.Login)
	add(http.MethodPost, "/login", p.Login)
}

func (p *provider) Login(rw http.ResponseWriter, r *http.Request) {
	checker, ok := p.Match(r)
	if ok {
		result, err := checker(r)
		if err == nil && result.Success {
			rw.Write([]byte("OK"))
			return
		}
	}
	// Need to return `401` for browsers to pop-up login box.
	rw.Header().Set("WWW-Authenticate", "basic realm=Restricted")
	http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func init() {
	servicehub.Register("openapi-auth-fixed-users", &servicehub.Spec{
		Services:   []string{"openapi-auth-fixed-users"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
