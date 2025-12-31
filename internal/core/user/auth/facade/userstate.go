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

package facade

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/util"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
	uutil "github.com/erda-project/erda/internal/core/user/util"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

type GetUserState int
type SetUserState int

const (
	GetInit GetUserState = iota
	GotToken
	GotInfo
	GotScopeInfo
)

type userState struct {
	info          *common.UserInfo
	scopeInfo     common.UserScopeInfo
	state         GetUserState
	authenticator domain.RequestAuthenticator

	// dependencies
	UserOAuthService pb.UserOAuthServiceServer
	identity         domain.Identity
	bundle           *bundle.Bundle
	credStore        domain.CredentialStore
}

func (u *userState) get(req *http.Request, targetState GetUserState) (interface{}, domain.UserAuthResult) {
	ctx := req.Context()
	for {
		switch u.state {
		case GetInit:
			credential, err := u.credStore.Load(ctx, req)
			if err != nil {
				logrus.WithField("state", u.state).Errorf("failed to load token, %v", err)
				return nil, domain.UserAuthResult{Code: domain.Unauthed, Detail: "User:State:GetInit"}
			}
			u.authenticator = credential.Authenticator
			u.state = GotToken
		case GotToken:
			if targetState == GotToken {
				return nil, domain.UserAuthResult{Code: domain.AuthSuccess}
			}
			userInfo, err := u.identity.Me(req.Context(), u.authenticator)
			if err != nil {
				return nil, domain.UserAuthResult{Code: domain.Unauthed, Detail: err.Error()}
			}
			u.info = userInfo
			u.state = GotInfo
		case GotInfo:
			if targetState == GotInfo {
				return u.info, domain.UserAuthResult{Code: domain.AuthSuccess}
			}
			orgHeader := req.Header.Get("org")
			domainHeader := req.Header.Get("domain")
			orgID, err := u.GetOrgInfo(orgHeader, domainHeader)
			if err != nil {
				return nil, domain.UserAuthResult{Code: domain.InternalAuthErr, Detail: err.Error()}
			}
			if orgID > 0 {
				role, err := u.bundle.ScopeRoleAccess(string(u.info.ID), &apistructs.ScopeRoleAccessRequest{
					Scope: apistructs.Scope{
						Type: apistructs.OrgScope,
						ID:   strconv.FormatUint(orgID, 10),
					},
				})
				if err != nil {
					return nil, domain.UserAuthResult{Code: domain.InternalAuthErr, Detail: err.Error()}
				}
				if !role.Access {
					return nil, domain.UserAuthResult{
						Code:   domain.AuthFail,
						Detail: fmt.Sprintf("org access denied: userID: %v, orgID: %v", u.info.ID, orgID),
					}
				}
				var scopeInfo common.UserScopeInfo
				scopeInfo.OrgID = orgID
				u.scopeInfo = scopeInfo
			}
			u.state = GotScopeInfo
		case GotScopeInfo:
			if targetState == GotScopeInfo {
				return u.scopeInfo, domain.UserAuthResult{Code: domain.AuthSuccess}
			}
		default:
			panic("invalid state")
		}
	}
}

func (u *userState) GetOrgInfo(orgHeader, domainHeader string) (orgID uint64, err error) {
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

func (u *userState) IsLogin(req *http.Request) domain.UserAuthResult {
	_, authr := u.get(req, GotToken)
	return authr
}

func (u *userState) GetInfo(req *http.Request) (common.UserInfo, domain.UserAuthResult) {
	info, authr := u.get(req, GotInfo)
	if authr.Code != domain.AuthSuccess {
		return common.UserInfo{}, authr
	}
	userInfo := info.(*common.UserInfo)
	return *userInfo, authr
}

func (u *userState) GetScopeInfo(req *http.Request) (common.UserScopeInfo, domain.UserAuthResult) {
	scopeInfo, authr := u.get(req, GotScopeInfo)
	if authr.Code != domain.AuthSuccess {
		return common.UserScopeInfo{}, authr
	}
	return scopeInfo.(common.UserScopeInfo), authr
}

func (u *userState) Login(code string, queryValues url.Values) error {
	oauthToken, err := u.UserOAuthService.ExchangeCode(context.Background(), &pb.ExchangeCodeRequest{
		Code:        code,
		ExtraParams: uutil.ConvertURLValuesToPb(queryValues),
	})
	if err != nil {
		logrus.Errorf("failed to login with exchange code, %v", err)
		return err
	}
	persistedCredential, err := u.credStore.Persist(context.Background(), &domain.AuthCredential{
		OAuthToken: uutil.ConvertPbToOAuthDomain(oauthToken),
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

func (u *userState) PwdLogin(username, password string) error {
	oauthToken, err := u.UserOAuthService.ExchangePassword(context.Background(), &pb.ExchangePasswordRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}
	persistedCredential, err := u.credStore.Persist(context.Background(), &domain.AuthCredential{
		OAuthToken: uutil.ConvertPbToOAuthDomain(oauthToken),
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
