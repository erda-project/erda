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

package legacy

import (
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
)

type config struct {
	Order             int
	OldCookieDomain   string `file:"old_cookie_domain"`
	SessionCookieName string `file:"session_cookie_name"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	if len(p.Cfg.OldCookieDomain) <= 0 {
		return h
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		cs := r.Cookies()
		count := 0
		for _, c := range cs {
			if c.Name == p.Cfg.SessionCookieName {
				count++
			}
		}
		if count > 1 {
			p.Log.Debugf("reset cookie to remove old session")
			http.SetCookie(rw, &http.Cookie{
				Name:     p.Cfg.SessionCookieName,
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				Domain:   p.Cfg.OldCookieDomain,
				HttpOnly: true,
			})
		}
		h(rw, r)
	}
}

func init() {
	servicehub.Register("openapi-interceptor-auth-session-compatibility", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-auth-session-compatibility"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
