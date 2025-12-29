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

package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/core/openapi/legacy/auth"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/common"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

func (p *provider) LoginURL(rw http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	if len(referer) <= 0 {
		referer = p.Cfg.RedirectAfterLogin
	}

	authURL, err := p.OAuthLoginFlowProvider.AuthURL(context.Background(), referer)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	common.ResponseJSON(rw, &struct {
		URL string `json:"url"`
	}{
		URL: authURL,
	})
}

func (p *provider) LoginCallback(rw http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	code := queryValues.Get("code")
	referer := queryValues.Get("referer")
	state := queryValues.Get("state")

	redirectAfterLogin := state
	if redirectAfterLogin == "" {
		redirectAfterLogin = referer
	}

	user := auth.NewUser(p.CredStore)
	if err := user.Login(code, queryValues); err != nil {
		p.Log.Errorf("failed to login: %v", err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	if !p.referMatcher.Match(redirectAfterLogin) {
		http.Error(rw, "invalid referer", http.StatusBadRequest)
		return
	}

	http.Redirect(rw, r, redirectAfterLogin, http.StatusFound)
}

func (p *provider) Logout(rw http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	if len(referer) <= 0 {
		referer = p.Cfg.RedirectAfterLogin
	}

	if cred, err := p.CredStore.Load(context.Background(), r); err == nil && cred != nil && cred.SessionID != "" {
		r = r.WithContext(context.WithValue(r.Context(), "session", cred.SessionID))
		if revoker, ok := p.CredStore.(domain.SessionRevoker); ok {
			if err := revoker.Revoke(context.Background(), cred.SessionID); err != nil {
				err := fmt.Errorf("logout: %v", err)
				p.Log.Error(err)
				http.Error(rw, err.Error(), http.StatusBadGateway)
				return
			}
		}
	}

	scheme := p.getScheme(r)
	http.SetCookie(rw, &http.Cookie{
		Name:     p.Cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Domain:   p.getSessionDomain(r.Host),
		HttpOnly: true,
		Secure:   scheme == "https",
	})

	redirectURL, err := p.OAuthLoginFlowProvider.AuthURL(context.Background(), referer)
	if err != nil {
		p.Log.Errorf("failed to get oauth url, %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	common.ResponseJSON(rw, &struct {
		URL string `json:"url"`
	}{
		URL: fmt.Sprintf("%s/logout?redirectUrl=%s", p.Cfg.IAMPublicURL, url.QueryEscape(redirectURL)),
	})
}

func (p *provider) getScheme(r *http.Request) string {
	// get from standard header first
	proto := firstNonEmpty(r.Header.Get("X-Forwarded-Proto"), r.Header.Get("X-Forwarded-Protocol"), r.URL.Scheme)
	if len(proto) > 0 {
		return proto
	}
	return p.Cfg.PlatformProtocol
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if len(s) > 0 {
			list := strings.Split(s, ",")
			for _, item := range list {
				v := strings.ToLower(strings.TrimSpace(item))
				if len(v) > 0 {
					return v
				}
			}
		}
	}
	return ""
}

func (p *provider) getSessionDomain(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	domains := strings.SplitN(host, ".", -1)
	l := len(domains)
	if l < 2 {
		return ""
	}
	rootDomain := "." + domains[l-2] + "." + domains[l-1]
	for _, domain := range p.Cfg.SessionCookieDomains {
		if strings.Contains(domain, rootDomain) {
			return domain
		}
	}
	return ""
}

func (p *provider) getUCRedirectHost(referer, host string) string {
	for _, addr := range p.Cfg.UCRedirectAddrs {
		domain := addr
		// ignore service port
		for _, v := range []string{addr, host} {
			if h, _, err := net.SplitHostPort(v); err == nil {
				domain = h
			}
		}
		domains := strings.SplitN(domain, ".", -1)
		l := len(domains)
		if l < 2 {
			continue
		}
		rootDomain := domains[l-2] + "." + domains[l-1]
		if strings.Contains(referer, rootDomain) {
			return addr
		}
	}
	return host
}
