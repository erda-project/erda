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

package uc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
	"github.com/erda-project/erda/pkg/ucauth"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

type config struct {
	SessionKey          string        `file:"session_key" default:"OPENAPISESSION"`
	SessionExpiration   time.Duration `file:"session_expiration" default:"120h"` // 5 days
	SessionCookieDomain []string      `file:"session_cookie_domain" env:"COOKIE_DOMAIN"`
	RedisKeyPrefix      string        `file:"redis_key_prefix" default:"openapi:sessionid:"`
	PublicURL           string        `file:"public_url" env:"SELF_PUBLIC_URL"`
	UCPublicURL         string        `file:"uc_public_url" env:"UC_PUBLIC_URL"`
	UCAddr              string        `file:"uc_addr" env:"UC_ADDR"`
	UCClientID          string        `file:"uc_client_id" env:"UC_CLIENT_ID"`
	UCClientSecret      string        `file:"uc_client_secret" env:"UC_CLIENT_SECRET"`
	DefaultRedirectURL  string        `file:"default_redirect_url" env:"UI_PUBLIC_URL"`
}

// +provider
type provider struct {
	Cfg   *config
	Log   logs.Logger
	Redis *redis.Client `autowired:"redis-client"`
	bdl   *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithCMDB())
	return nil
}

func (p *provider) Authers() []auth.Auther {
	return []auth.Auther{
		&BasicAuther{
			name:              "basic",
			getUCUserAuth:     p.getUCUserAuth,
			makeSessionCookie: p.makeSessionCookie,
			bdl:               p.bdl,
		},
		&LoginAuther{
			name: "session",
			// session
			sessionKey:        p.Cfg.SessionKey,
			getUCUserAuth:     p.getUCUserAuth,
			makeSessionCookie: p.makeSessionCookie,
			getSession:        p.getSession,
			bdl:               p.bdl,
			// urls
			publicURL:          p.Cfg.PublicURL,
			ucPublicURL:        p.Cfg.UCPublicURL,
			clientID:           p.Cfg.UCClientID,
			defaultRedirectURL: p.Cfg.DefaultRedirectURL,
		},
	}
}

func (p *provider) makeSessionCookie(r *http.Request, token string) (*http.Cookie, error) {
	var sessionID string
	session, err := r.Cookie(p.Cfg.SessionKey)
	if err == nil && len(session.Value) > 0 {
		sessionID = session.Value
	} else {
		sessionID = genSessionID()
	}
	expires, err := p.storeSession(sessionID, token)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     p.Cfg.SessionKey,
		Value:    sessionID,
		Domain:   getDomain(r.Host, p.Cfg.SessionCookieDomain),
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
		Expires:  time.Now().Add(expires),
	}, nil
}

func (p *provider) storeSession(sessionID, token string) (time.Duration, error) {
	_, err := p.Redis.Set(p.redisSessionKey(sessionID), token, p.Cfg.SessionExpiration).Result()
	if err != nil {
		err = fmt.Errorf("fail to store session: %w", err)
		p.Log.Error(err)
		return p.Cfg.SessionExpiration, err
	}
	return p.Cfg.SessionExpiration, nil
}

func (p *provider) getSession(session string) (string, bool, error) {
	token, err := p.Redis.Get(p.redisSessionKey(session)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return token, true, nil
}

func (p *provider) getUCUserAuth() *ucauth.UCUserAuth {
	return ucauth.NewUCUserAuth(p.Cfg.UCPublicURL, p.Cfg.UCAddr, "", p.Cfg.UCClientID, p.Cfg.UCClientSecret)
}

func (p *provider) redisSessionKey(sessionID string) string {
	return p.Cfg.RedisKeyPrefix + sessionID
}

func genSessionID() string {
	return uuid.NewV4().String()
}

func getDomain(host string, domains []string) string {
	for _, v := range domains {
		if strings.HasSuffix(host, v) {
			return v
		}
	}
	return ""
}

func init() {
	servicehub.Register("openapi-auth-uc", &servicehub.Spec{
		Services:   []string{"openapi-auth-uc"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
