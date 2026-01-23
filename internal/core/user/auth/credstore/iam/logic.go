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

package iam

import (
	"context"
	"errors"
	"net/http"

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
		Authenticator: &applier.QueryTokenAuth{
			Param: "token",
			Token: cookie.Value,
		},
		AccessToken: cookie.Value,
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
	case cred.JWTToken != "":
		return &domain.PersistedCredential{
			Authenticator: &applier.QueryTokenAuth{
				Param: "token",
				Token: cred.JWTToken,
			},
			AccessToken: cred.JWTToken,
		}, nil
	default:
		return nil, errors.New("unreachable auth credential")
	}
}

func (p *provider) WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *common.SessionRefresh) error {
	if refresh == nil || refresh.Token == "" {
		return nil
	}
	c := &http.Cookie{
		Name:     p.Config.CookieName,
		Value:    refresh.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   req.TLS != nil,
		SameSite: http.SameSite(p.Config.CookieSameSite),
	}

	if cfg := refresh.Cookie; cfg != nil {
		if cfg.Name != "" {
			c.Name = cfg.Name
		}
		if cfg.Path != "" {
			c.Path = cfg.Path
		}
		if cfg.Domain != "" {
			c.Domain = cfg.Domain
		}
		if !cfg.Expires.IsZero() {
			c.Expires = cfg.Expires
		}
		if cfg.MaxAge != 0 {
			c.MaxAge = cfg.MaxAge
		}
		c.HttpOnly = cfg.HttpOnly
		c.Secure = cfg.Secure
		if cfg.SameSite > 0 {
			c.SameSite = cfg.SameSite
		}
	}
	http.SetCookie(rw, c)
	return nil
}
