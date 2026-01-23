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
	"net/http"

	"github.com/pkg/errors"

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

func (p *provider) WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *common.SessionRefresh) error {
	if refresh == nil {
		return errors.New("refresh cookie is nil")
	}
	c := &http.Cookie{
		Name:     p.Config.CookieName,
		Value:    refresh.Cookie.Value,
		Path:     refresh.Cookie.Path,
		HttpOnly: refresh.Cookie.HttpOnly,
		Expires:  refresh.Cookie.Expires,
		Domain:   refresh.Cookie.Domain,
		Secure:   req.TLS != nil,
		SameSite: http.SameSite(p.Config.CookieSameSite),
	}
	http.SetCookie(rw, c)
	return nil
}
