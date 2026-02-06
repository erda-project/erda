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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/util"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type ResponseMeta struct {
	Success *bool  `json:"success"`
	Error   string `json:"error"`
}

func (p *provider) AuthURL(_ context.Context, r *pb.AuthURLRequest) (*pb.AuthURLResponse, error) {
	redirectUri := fmt.Sprintf("%s?referer=%s", p.Config.RedirectURI, r.Referer)

	q := make(url.Values)
	q.Set("response_type", "code")
	q.Set("client_id", p.Config.ClientID)
	q.Set("redirect_uri", redirectUri)
	q.Set("scope", "public_profile")

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path = "/oauth/authorize"
	baseURL.RawQuery = q.Encode()
	return &pb.AuthURLResponse{Data: baseURL.String()}, nil
}

func (p *provider) LogoutURL(ctx context.Context, r *pb.LogoutURLRequest) (*pb.LogoutURLResponse, error) {
	redirectURL, err := p.AuthURL(ctx, &pb.AuthURLRequest{
		Referer: r.Referer,
	})
	if err != nil {
		return nil, err
	}

	q := make(url.Values)
	q.Set("redirectUrl", redirectURL.Data)

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path = "logout"
	baseURL.RawQuery = q.Encode()
	return &pb.LogoutURLResponse{
		Data: baseURL.String(),
	}, nil
}

func (p *provider) ExchangeCode(_ context.Context, r *pb.ExchangeCodeRequest) (*pb.OAuthToken, error) {
	redirectURI := p.Config.RedirectURI
	referer, ok := r.ExtraParams["referer"]
	if ok {
		redirectURI = fmt.Sprintf("%s?referer=%s", redirectURI, referer.Values[0])
	}

	formBody := make(url.Values)
	formBody.Set("grant_type", "authorization_code")
	formBody.Set("code", r.Code)
	formBody.Set("redirect_uri", redirectURI)

	t, err := p.doExchange(formBody)
	if err != nil {
		return nil, err
	}
	return util.ConvertOAuthDomainToPb(t), nil
}

func (p *provider) ExchangePassword(ctx context.Context, r *pb.ExchangePasswordRequest) (*pb.OAuthToken, error) {
	formBody := make(url.Values)
	formBody.Set("grant_type", "password")
	formBody.Set("username", r.Username)
	formBody.Set("password", r.Password)
	formBody.Set("scope", "public_profile")

	t, err := p.doExchange(formBody)
	if err != nil {
		return nil, err
	}
	return util.ConvertOAuthDomainToPb(t), nil
}

func (p *provider) ExchangeClientCredentials(ctx context.Context, r *pb.ExchangeClientCredentialsRequest) (*pb.OAuthToken, error) {
	if !r.Refresh && p.Config.ServerTokenCacheEnabled {
		cacheToken, err := p.tokenCache.Get(serverTokenCacheKey)
		if err != nil {
			p.Log.Warnf("failed to get server token from cache, %v", err)
		} else {
			oauthToken, ok := cacheToken.(*domain.OAuthToken)
			if ok {
				return util.ConvertOAuthDomainToPb(oauthToken), nil
			}
			p.Log.Warn("server cache token is not *domain.OAuthToken")
		}
	}

	formBody := make(url.Values)
	formBody.Set("grant_type", "client_credentials")

	serverToken, err := p.doExchange(formBody)
	if err != nil {
		return nil, err
	}

	if p.Config.ServerTokenCacheEnabled {
		expireTime := p.convertExpiresIn2Time(serverToken.ExpiresIn)
		if err := p.tokenCache.SetWithExpire(serverTokenCacheKey, serverToken, expireTime); err != nil {
			p.Log.Warnf("failed to set token with expire %s, %v", expireTime.String(), err)
		}
		p.Log.Infof("grant new client_credential token with expire time %s", expireTime.String())
	}

	return util.ConvertOAuthDomainToPb(serverToken), nil
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

	bodyBytes, err := io.ReadAll(&body)
	if err != nil {
		return nil, err
	}

	token, err := DecodeUCFlat[domain.OAuthToken](bodyBytes)
	if err != nil {
		p.Log.Errorf("failed to get exchange token, %v", err)
		return nil, err
	}

	return token, nil
}

func (p *provider) convertExpiresIn2Time(expiresIn int64) time.Duration {
	return time.Duration(float64(expiresIn)*p.Config.TokenCacheEarlyExpireRate) * time.Second
}

func DecodeUCFlat[T any](body []byte) (*T, error) {
	var meta ResponseMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}

	if meta.Success != nil && !*meta.Success {
		if meta.Error != "" {
			return nil, errors.New(meta.Error)
		}
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
