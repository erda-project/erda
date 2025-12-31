package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/util"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (p *provider) ExchangeCode(ctx context.Context, r *pb.ExchangeCodeRequest) (*pb.OAuthToken, error) {
	formBody := make(url.Values)
	formBody.Set("grant_type", "authorization_code")
	formBody.Set("code", r.Code)
	formBody.Set("redirect_uri", p.Config.RedirectURI)

	oauthToken, err := p.doExchange(ctx, formBody)
	if err != nil {
		return nil, err
	}
	return util.ConvertOAuthDomainToPb(oauthToken), nil
}

func (p *provider) ExchangePassword(ctx context.Context, r *pb.ExchangePasswordRequest) (*pb.OAuthToken, error) {
	if p.Config.UserTokenCacheEnabled {
		cacheTokenAny, err := p.tokenCache.Get(userTokenCacheKey(r.Username, r.Password))
		if err != nil {
			p.Log.Warnf("failed to get user token from cache (username: %s), %v", r.Username, err)
		} else {
			cacheToken, ok := cacheTokenAny.(*domain.OAuthToken)
			if ok {
				p.Log.Infof("cached get user token: %s, %s", r.Username, cacheToken.AccessToken)
				return util.ConvertOAuthDomainToPb(cacheToken), nil
			}
			p.Log.Warn("user cache token is not *domain.OAuthToken")
		}
	}

	formBody := make(url.Values)
	formBody.Set("grant_type", "password")
	formBody.Set("username", r.Username)
	formBody.Set("password", r.Password)
	// fixed scope user_info
	formBody.Set("scope", "user_info")

	oauthToken, err := p.doExchange(ctx, formBody)
	if err != nil {
		return nil, err
	}

	if p.Config.UserTokenCacheEnabled {
		expireTime := p.convertExpiresIn2Time(oauthToken.ExpiresIn)
		if err := p.tokenCache.SetWithExpire(userTokenCacheKey(r.Username, r.Password), oauthToken, expireTime); err != nil {
			p.Log.Warnf("failed to set token with expire %s (username: %s), %v", expireTime.String(), r.Username, err)
		}
		p.Log.Infof("grant new password token with expire time %s (username: %s)", expireTime.String(), r.Username)
	}

	return util.ConvertOAuthDomainToPb(oauthToken), nil
}

func (p *provider) ExchangeClientCredentials(ctx context.Context, r *pb.ExchangeClientCredentialsRequest) (*pb.OAuthToken, error) {
	// load from cache
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

	serverToken, err := p.doExchange(ctx, formBody)
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

func (p *provider) doExchange(_ context.Context, formBody url.Values) (*domain.OAuthToken, error) {
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

func (p *provider) AuthURL(_ context.Context, r *pb.AuthURLRequest) (*pb.AuthURLResponse, error) {
	q := make(url.Values)
	q.Set("state", r.Referer)
	q.Set("response_type", "code")
	q.Set("client_id", p.Config.ClientID)
	q.Set("redirect_uri", p.Config.RedirectURI)
	q.Set("scope", "api")

	baseURL, err := url.Parse(p.Config.FrontendURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path = "/iam/oauth2/server/authorize"
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

func (p *provider) convertExpiresIn2Time(expiresIn int64) time.Duration {
	return time.Duration(float64(expiresIn)*p.Config.TokenCacheEarlyExpireRate) * time.Second
}

func userTokenCacheKey(username, password string) string {
	return userTokenCachePrefix + username + ":" + password
}
