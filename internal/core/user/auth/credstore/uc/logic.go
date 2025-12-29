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
	"net/http"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/erda-project/erda/internal/core/user/auth/applier"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

func (p *provider) Load(_ context.Context, r *http.Request) (*domain.PersistedCredential, error) {
	session, err := r.Cookie(p.Config.CookieName)
	if err != nil {
		return nil, err
	}

	token, err := p.Redis.Get(makeSessionKey(session.Value)).Result()
	if err != nil {
		return nil, err
	}

	return &domain.PersistedCredential{
		Authenticator: &applier.BearerTokenAuth{
			Token: token,
		},
		AccessToken: token,
		SessionID:   session.Value,
	}, nil
}

func (p *provider) Persist(_ context.Context, cred *domain.PersistedCredential) (*domain.PersistedCredential, error) {
	if cred.AccessToken == "" {
		return nil, errors.New("credential token is empty")
	}
	sessionID := genSessionID()
	if _, err := p.Redis.Set(makeSessionKey(sessionID), cred.AccessToken, p.Config.Expire).Result(); err != nil {
		return nil, errors.Wrap(err, "failed to store session")
	}
	// TODO: new credential with cookie or session?
	cred.SessionID = sessionID
	return cred, nil
}

func (p *provider) Revoke(_ context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	_, err := p.Redis.Del(makeSessionKey(sessionID)).Result()
	return err
}

//func (p *provider) WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *common.SessionRefresh) error {
//	if refresh == nil || refresh.Token == "" {
//		return nil
//	}
//	c := &http.Cookie{
//		Name:     p.Config.CookieName,
//		Value:    refresh.SessionID,
//		Path:     "/",
//		HttpOnly: true,
//		Secure:   req.TLS != nil,
//		SameSite: http.SameSiteDefaultMode,
//	}
//
//	http.SetCookie(rw, c)
//	return nil
//}

func makeSessionKey(sessionID string) string {
	return fmt.Sprintf("openapi:sessionid:%s", sessionID)
}

func genSessionID() string {
	return uuid.NewV4().String()
}
