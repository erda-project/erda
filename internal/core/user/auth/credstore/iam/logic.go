package iam

import (
	"context"
	"errors"
	"net/http"
	"time"

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
		SameSite: http.SameSiteDefaultMode,
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
		if cfg.MaxAge != 0 {
			c.MaxAge = cfg.MaxAge
			c.Expires = time.Now().Add(time.Duration(cfg.MaxAge) * time.Second)
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
