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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/util"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	identity "github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

type GetUserState int
type SetUserState int

const (
	GetInit GetUserState = iota
	GotSessionID
	GotToken
	GotInfo
	GotScopeInfo
)

const (
	SetInit SetUserState = iota
	SetSessionID
)

const (
	SessionExpireDays = 5
)

var (
	ErrNotExist = errors.New("session not exist")
)

type ScopeInfo struct {
	OrgID uint64 `json:"orgId"`
	// dont care other fields
}

type User struct {
	credStore          domain.CredentialStore
	authenticator      domain.RequestAuthenticator
	oauthTokenProvider domain.OAuthTokenProvider
	identity           domain.Identity
	info               *identity.UserInfo
	scopeInfo          ScopeInfo
	state              GetUserState

	bundle *bundle.Bundle
}

func NewUser(store domain.CredentialStore) *User {
	return &User{
		state:              GetInit,
		credStore:          store,
		oauthTokenProvider: legacycontainer.Get[domain.OAuthTokenProvider](),
		identity:           legacycontainer.Get[domain.Identity](),
		bundle:             bundle.New(bundle.WithErdaServer(), bundle.WithDOP()),
	}
}

func (u *User) get(req *http.Request, targetState GetUserState) (interface{}, AuthResult) {
	ctx := req.Context()
	for {
		switch u.state {
		case GetInit:
			credential, err := u.credStore.Load(ctx, req)
			if err != nil {
				logrus.WithField("state", u.state).Errorf("failed to load token, %v", err)
				return nil, AuthResult{Unauthed, "User:GetInit"}
			}
			u.authenticator = credential.Authenticator
			u.state = GotToken
		case GotToken:
			if targetState == GotToken {
				return nil, AuthResult{AuthSucc, ""}
			}
			userInfo, err := u.identity.Me(req.Context(), u.authenticator)
			if err != nil {
				return nil, AuthResult{Unauthed, err.Error()}
			}
			u.info = userInfo
			u.state = GotInfo
		case GotInfo:
			if targetState == GotInfo {
				return u.info, AuthResult{AuthSucc, ""}
			}
			orgHeader := req.Header.Get("org")
			domainHeader := req.Header.Get("domain")
			orgID, err := u.GetOrgInfo(orgHeader, domainHeader)
			if err != nil {
				return nil, AuthResult{InternalAuthErr, err.Error()}
			}
			if orgID > 0 {
				role, err := u.bundle.ScopeRoleAccess(string(u.info.ID), &apistructs.ScopeRoleAccessRequest{
					Scope: apistructs.Scope{
						Type: apistructs.OrgScope,
						ID:   strconv.FormatUint(orgID, 10),
					},
				})
				if err != nil {
					return nil, AuthResult{InternalAuthErr, err.Error()}
				}
				if !role.Access {
					return nil, AuthResult{AuthFail, fmt.Sprintf("org access denied: userID: %v, orgID: %v", u.info.ID, orgID)}
				}
				var scopeInfo ScopeInfo
				scopeInfo.OrgID = orgID
				u.scopeInfo = scopeInfo
			}
			u.state = GotScopeInfo
		case GotScopeInfo:
			if targetState == GotScopeInfo {
				return u.scopeInfo, AuthResult{AuthSucc, ""}
			}
		default:
			panic("invalid state")
		}
	}
}

func (u *User) GetOrgInfo(orgHeader, domainHeader string) (orgID uint64, err error) {
	logrus.Debugf("orgHeader: %v, domainHeader: %v", orgHeader, domainHeader)
	var orgName string
	// try to get from org header firstly
	if orgHeader != "" && orgHeader != "-" {
		orgName = orgHeader
	}
	// try to get from domain header
	if orgName == "" {
		orgName, err = util.GetOrgByDomain(domainHeader)
		if err != nil {
			return 0, err
		}
	}
	// if cannot get orgName, just return
	if orgName == "" {
		return 0, nil
	}
	// query org info
	orgResp, err := org.MustGetOrg().GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcOpenapi), &orgpb.GetOrgRequest{
		IdOrName: orgName,
	})
	if err != nil {
		return 0, err
	}
	return orgResp.Data.ID, nil
}

func (u *User) IsLogin(req *http.Request) AuthResult {
	_, authr := u.get(req, GotToken)
	return authr
}

func (u *User) GetInfo(req *http.Request) (identity.UserInfo, AuthResult) {
	info, authr := u.get(req, GotInfo)
	if authr.Code != AuthSucc {
		return identity.UserInfo{}, authr
	}
	userInfo := info.(*identity.UserInfo)
	return *userInfo, authr
}

func (u *User) GetScopeInfo(req *http.Request) (ScopeInfo, AuthResult) {
	scopeInfo, authr := u.get(req, GotScopeInfo)
	if authr.Code != AuthSucc {
		return ScopeInfo{}, authr
	}
	return scopeInfo.(ScopeInfo), authr
}

func (u *User) Login(code string, queryValues url.Values) error {
	oauthToken, err := u.oauthTokenProvider.ExchangeCode(context.Background(), code, queryValues)
	if err != nil {
		logrus.Errorf("failed to login with exchange code, %v", err)
		return err
	}
	persistedCredential, err := u.credStore.Persist(context.Background(), &domain.AuthCredential{
		OAuthToken: oauthToken,
	})
	if err != nil {
		return err
	}
	u.authenticator = persistedCredential.Authenticator
	userInfo, err := u.identity.Me(context.Background(), u.authenticator)
	if err != nil {
		return err
	}
	u.info = userInfo
	u.state = GotInfo
	return nil
}

func (u *User) PwdLogin(username, password string) error {
	oauthToken, err := u.oauthTokenProvider.ExchangePassword(context.Background(), username, password, nil)
	if err != nil {
		logrus.Error(err)
		return err
	}
	persistedCredential, err := u.credStore.Persist(context.Background(), &domain.AuthCredential{
		OAuthToken: oauthToken,
	})
	if err != nil {
		return err
	}
	u.authenticator = persistedCredential.Authenticator
	userInfo, err := u.identity.Me(context.Background(), u.authenticator)
	if err != nil {
		return err
	}
	u.info = userInfo
	u.state = GotInfo
	return nil
}
