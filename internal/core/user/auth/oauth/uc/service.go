package uc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (p *provider) ExchangeCode(_ context.Context, code string, _ url.Values) (*domain.OAuthToken, error) {
	formBody := make(url.Values)
	formBody.Set("grant_type", "authorization_code")
	formBody.Set("code", code)
	formBody.Set("redirect_uri", p.Config.RedirectURI)

	return p.doExchange(formBody)
}

func (p *provider) ExchangePassword(_ context.Context, username, password string, _ url.Values) (*domain.OAuthToken, error) {
	formBody := make(url.Values)
	formBody.Set("grant_type", "password")
	formBody.Set("username", username)
	formBody.Set("password", password)
	formBody.Set("scope", "public_profile")

	return p.doExchange(formBody)
}

func (p *provider) ExchangeClientCredentials(_ context.Context, refresh bool, _ url.Values) (*domain.OAuthToken, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.serverToken != nil && p.serverTokenExpireTime.After(time.Now().Add(tokenRefreshMargin)) && !refresh {
		return p.serverToken, nil
	}

	formBody := make(url.Values)
	formBody.Set("grant_type", "client_credentials")

	serverToken, err := p.doExchange(formBody)
	if err != nil {
		return nil, err
	}

	p.serverToken = serverToken
	p.serverTokenExpireTime = time.Now().Add(time.Duration(serverToken.ExpiresIn) * time.Second)
	return serverToken, nil
}

func (p *provider) doExchange(formBody url.Values) (*domain.OAuthToken, error) {
	var (
		body    bytes.Buffer
		reqPath = "/oauth/token"
	)

	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		BasicAuth(p.Config.ClientID, p.Config.ClientSecret).
		Post(p.Config.BackendHost).
		Path(reqPath).
		FormBody(formBody).Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request uc")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var oauthToken domain.OAuthToken
	if err := json.NewDecoder(&body).Decode(&oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

func (p *provider) AuthURL(_ context.Context, referer string) (string, error) {
	q := make(url.Values)
	q.Set("response_type", "code")
	q.Set("client_id", p.Config.ClientID)
	q.Set("redirect_uri", p.Config.RedirectURI)
	q.Set("scope", "public_profile")
	q.Set("referer", referer)

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return "", err
	}

	baseURL.Path = "/oauth/authorize"
	baseURL.RawQuery = q.Encode()
	return baseURL.String(), nil
}
