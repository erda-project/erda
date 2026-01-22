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

package auth

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/spec"
	"github.com/erda-project/erda/internal/core/openapi/legacy/monitor"
	"github.com/erda-project/erda/internal/core/openapi/settings"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	identity "github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
	"github.com/erda-project/erda/pkg/oauth2"
)

type TokenClient struct {
	ClientID   string
	ClientName string
}

const (
	HeaderAuthorization                = "Authorization"
	HeaderAuthorizationBearerPrefix    = "Bearer "
	HeaderAuthorizationBasicAuthPrefix = "Basic "
)

type Auth struct {
	UserAuth     domain.UserAuthFacade
	OAuth2Server *oauth2.OAuth2Server
	TokenService tokenpb.TokenServiceServer
	Settings     settings.OpenapiSettings
	CredStore    domain.CredentialStore
}

func NewAuth(oauth2server *oauth2.OAuth2Server, token tokenpb.TokenServiceServer,
	settings settings.OpenapiSettings) (*Auth, error) {
	return &Auth{
		UserAuth:     legacycontainer.Get[domain.UserAuthFacade](),
		OAuth2Server: oauth2server,
		TokenService: token,
		Settings:     settings,
		CredStore:    legacycontainer.Get[domain.CredentialStore](),
	}, nil
}

func (a *Auth) Auth(spec *spec.Spec, req *http.Request) domain.UserAuthResult {
	r := domain.UserAuthResult{Code: domain.AuthSuccess}
	defer func() {
		if r.Code != domain.AuthSuccess {
			monitor.Notify(monitor.Info{
				Tp: monitor.AuthFail, Detail: r.Detail,
			})
		} else {
			monitor.Notify(monitor.Info{
				Tp: monitor.AuthSucc, Detail: "",
			})
		}
	}()

	var t checkType
	t, err := a.whichCheck(req, spec)
	if err != nil {
		return domain.UserAuthResult{Code: domain.Unauthed, Detail: err.Error()}
	}
	switch t {
	case NONE:
		break
	case LOGIN:
		user := a.UserAuth.NewState()
		if r = user.IsLogin(req); r.Code != domain.AuthSuccess {
			return r
		}
		if r = applyUserIdentity(req, user); r.Code != domain.AuthSuccess {
			return r
		}
	case TRY_LOGIN:
		user := a.UserAuth.NewState()
		if r := user.IsLogin(req); r.Code != domain.AuthSuccess {
			break
		}
		if r := applyUserIdentity(req, user); r.Code != domain.AuthSuccess {
			break
		}
	case BASICAUTH:
		user := a.UserAuth.NewState()
		if r = a.checkBasicAuth(req, user); r.Code != domain.AuthSuccess {
			return r
		}
		if r = applyUserIdentity(req, user); r.Code != domain.AuthSuccess {
			return r
		}
	case TOKEN:
		var tokenClient TokenClient
		tokenClient, r = a.checkToken(spec, req)
		if r.Code != domain.AuthSuccess {
			return r
		}
		req.Header.Set("Client-ID", tokenClient.ClientID)
		req.Header.Set("Client-Name", tokenClient.ClientName)
	}
	return r
}

func applyUserIdentity(req *http.Request, user domain.UserAuthState) domain.UserAuthResult {
	var (
		userinfo identity.UserInfo
		r        = domain.UserAuthResult{
			Code: domain.AuthSuccess,
		}
	)

	userinfo, r = user.GetInfo(req)
	if r.Code != domain.AuthSuccess {
		return r
	}

	// set User-ID
	req.Header.Set("User-ID", string(userinfo.ID))

	// with session refresh context
	if newCtx := WithSessionRefresh(req.Context(), userinfo.SessionRefresh); newCtx != req.Context() {
		*req = *req.WithContext(newCtx)
	}

	var scopeInfo identity.UserScopeInfo
	scopeInfo, r = user.GetScopeInfo(req)
	if r.Code != domain.AuthSuccess {
		return r
	}

	// set Org-ID
	if scopeInfo.OrgID != 0 {
		req.Header.Set("Org-ID", strconv.FormatUint(scopeInfo.OrgID, 10))
	}
	return r
}

type checkType int

const (
	LOGIN checkType = iota
	TRY_LOGIN
	BASICAUTH
	TOKEN
	NONE
)

func (a *Auth) whichCheck(req *http.Request, spec *spec.Spec) (checkType, error) {
	cred, _ := a.CredStore.Load(req.Context(), req)
	if spec.CheckLogin && cred != nil {
		return LOGIN, nil
	}
	auth := req.Header.Get(HeaderAuthorization)
	if spec.CheckBasicAuth && strings.HasPrefix(auth, HeaderAuthorizationBasicAuthPrefix) {
		return BASICAUTH, nil
	}
	if spec.CheckToken && auth != "" {
		return TOKEN, nil
	}
	if spec.TryCheckLogin {
		return TRY_LOGIN, nil
	}
	if !spec.CheckToken && !spec.CheckLogin {
		return NONE, nil
	}
	return NONE, errors.New("lack of required auth header")
}

// checkToken try:
// 1. openapi oauth2 token
// 2. access key
func (a *Auth) checkToken(spec *spec.Spec, req *http.Request) (TokenClient, domain.UserAuthResult) {
	// 1. openapi oauth2 token
	oauth2TC, err := VerifyOpenapiOAuth2Token(a.OAuth2Server, &OpenapiSpec{Spec: spec}, req)
	if err != nil {
		logrus.Errorf("failed to verify openapi oauth token, %v", err)
	} else {
		return oauth2TC, domain.UserAuthResult{Code: domain.AuthSuccess}
	}
	// 2. access key
	ak, err := VerifyAccessKey(a.TokenService, req)
	if err != nil {
		return TokenClient{}, domain.UserAuthResult{Code: domain.AuthFail, Detail: err.Error()}
	}
	return ak, domain.UserAuthResult{Code: domain.AuthSuccess}
}

func (a *Auth) checkBasicAuth(req *http.Request, user domain.UserAuthState) domain.UserAuthResult {
	auth := req.Header.Get(HeaderAuthorization)
	if auth == "" {
		return domain.UserAuthResult{Code: domain.AuthFail, Detail: "missing Authorization header"}
	}

	if !strings.HasPrefix(auth, HeaderAuthorizationBasicAuthPrefix) {
		return domain.UserAuthResult{Code: domain.AuthFail, Detail: "Authorization is not Basic"}
	}

	auth = strings.TrimPrefix(auth, HeaderAuthorizationBasicAuthPrefix)

	raw, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return domain.UserAuthResult{Code: domain.AuthFail, Detail: "invalid base64 basic auth"}
	}

	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return domain.UserAuthResult{Code: domain.AuthFail, Detail: "invalid basic auth format"}
	}

	if err := user.PwdLogin(parts[0], parts[1]); err != nil {
		return domain.UserAuthResult{Code: domain.Unauthed, Detail: err.Error()}
	}

	return domain.UserAuthResult{Code: domain.AuthSuccess}
}
