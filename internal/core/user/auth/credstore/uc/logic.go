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
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/erda-project/erda/internal/core/user/auth/applier"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
)

func (p *provider) Load(_ context.Context, r *http.Request) (*domain.PersistedCredential, error) {
	cookie, err := r.Cookie(p.Config.CookieName)
	if err != nil {
		return nil, err
	}

	return &domain.PersistedCredential{
		Authenticator: &applier.CookieTokenAuth{
			Cookie: cookie,
		},
	}, nil
}

func (p *provider) Persist(_ context.Context, cred *domain.AuthCredential) (*domain.PersistedCredential, error) {
	if cred == nil {
		return nil, errors.New("credential is nil")
	}
	switch {
	case cred.OAuthToken != nil:
		return &domain.PersistedCredential{
			Authenticator: &applier.BearerTokenAuth{
				Token: cred.OAuthToken.AccessToken,
			},
			AccessToken: cred.OAuthToken.AccessToken,
		}, nil
	default:
		return nil, errors.New("unreachable auth credential")
	}
}

func (p *provider) Revoke(_ context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	_, err := p.Redis.Del(makeSessionKey(sessionID)).Result()
	return err
}

func (p *provider) WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *common.SessionRefresh) error {
	if refresh == nil || refresh.Token == "" {
		return nil
	}
	c := &http.Cookie{
		Name:     p.Config.CookieName,
		Value:    refresh.SessionID,
		Path:     "/",
		HttpOnly: true,
		Domain:   p.getSessionDomain(req.Host),
		Secure:   req.TLS != nil,
		SameSite: http.SameSite(p.Config.CookieSameSite),
	}
	http.SetCookie(rw, c)
	return nil
}

func makeSessionKey(sessionID string) string {
	return fmt.Sprintf("openapi:sessionid:%s", sessionID)
}

func genSessionID() string {
	return uuid.NewV4().String()
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
	for _, item := range p.Config.SessionCookieDomains {
		if strings.Contains(item, rootDomain) {
			return item
		}
	}
	return ""
}
