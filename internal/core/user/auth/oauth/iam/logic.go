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
	"bytes"
	"context"
	"encoding/json"
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
	// fixed scope user_info
	formBody.Set("scope", "user_info")

	return p.doExchange(formBody)
}

func (p *provider) ExchangeClientCredentials(_ context.Context, refresh bool, _ url.Values) (*domain.OAuthToken, error) {
	// load from cache
	if !refresh {
		cacheToken, err := p.serverTokenCache.Get(serverTokenCacheKey)
		if err != nil {
			p.Log.Warnf("failed to get server token from cache, %v", err)
		} else {
			oauthToken, ok := cacheToken.(*domain.OAuthToken)
			if ok {
				return oauthToken, nil
			}
			p.Log.Warn("cache token is not *domain.OAuthToken")
		}
	}

	formBody := make(url.Values)
	formBody.Set("grant_type", "client_credentials")

	serverToken, err := p.doExchange(formBody)
	if err != nil {
		return nil, err
	}

	expireTime := time.Duration(serverToken.ExpiresIn/10*8) * time.Second
	if err := p.serverTokenCache.SetWithExpire(serverTokenCacheKey, serverToken, expireTime); err != nil {
		p.Log.Warnf("failed to set token with expire %s, %v", expireTime.String(), err)
	}

	p.Log.Infof("new server token with expire time %s", expireTime.String())
	return serverToken, nil
}

func (p *provider) doExchange(formBody url.Values) (*domain.OAuthToken, error) {
	var (
		body    bytes.Buffer
		reqPath = "/iam/oauth2/server/token"
	)

	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		BasicAuth(p.Config.ClientID, p.Config.ClientSecret).
		Post(p.Config.BackendHost).
		Path(reqPath).
		FormBody(formBody).Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request iam")
	}
	if !r.IsOK() {
		p.Log.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
		return nil, errors.New("Unauthorized")
	}

	var oauthToken domain.OAuthToken
	if err := json.NewDecoder(&body).Decode(&oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

func (p *provider) AuthURL(_ context.Context, referer string) (string, error) {
	q := make(url.Values)
	q.Set("state", referer)
	q.Set("response_type", "code")
	q.Set("client_id", p.Config.ClientID)
	q.Set("redirect_uri", p.Config.RedirectURI)
	q.Set("scope", "api")

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return "", err
	}

	baseURL.Path = "/iam/oauth2/server/authorize"
	baseURL.RawQuery = q.Encode()
	return baseURL.String(), nil
}

func (p *provider) LogoutURL(ctx context.Context, referer string) (string, error) {
	redirectURL, err := p.AuthURL(ctx, referer)
	if err != nil {
		return "", err
	}

	q := make(url.Values)
	q.Set("redirectUrl", redirectURL)

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return "", err
	}

	baseURL.Path = "logout"
	baseURL.RawQuery = q.Encode()
	return baseURL.String(), nil
}
