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
	"net/textproto"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/conf"
	"github.com/erda-project/erda/internal/core/openapi/legacy/util"
	"github.com/erda-project/erda/internal/core/org"
	identity "github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
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
	sessionID  string
	token      uc.OAuthToken
	info       identity.UserInfo
	scopeInfo  ScopeInfo
	state      GetUserState
	redisCli   *redis.Client
	ucUserAuth *uc.UCUserAuth

	bundle *bundle.Bundle
}

var client = bundle.New(bundle.WithErdaServer(), bundle.WithDOP())

func NewUser(redisCli *redis.Client) *User {
	ucUserAuth := uc.NewUCUserAuth(conf.UCAddrFront(), discover.UC(), "http://"+conf.UCRedirectHost()+"/logincb", conf.UCClientID(), conf.UCClientSecret())
	if conf.OryEnabled() {
		ucUserAuth.ClientID = conf.OryCompatibleClientID()
		ucUserAuth.UCHost = conf.OryKratosAddr()
	}
	return &User{state: GetInit, redisCli: redisCli, ucUserAuth: ucUserAuth, bundle: client}
}

func (u *User) get(req *http.Request, state GetUserState) (interface{}, AuthResult) {
	switch u.state {
	case GetInit:
		session := req.Context().Value("session")
		if session == nil {
			return nil, AuthResult{Unauthed, "User:GetInit"}
		}
		u.sessionID = session.(string)
		u.state = GotSessionID
		fallthrough
	case GotSessionID:
		token, err := u.redisCli.Get(MkSessionKey(u.sessionID)).Result()
		if conf.OryEnabled() {
			// TODO: remove useless `token`
			token = u.sessionID
			err = nil
		}
		if err == redis.Nil {
			return nil, AuthResult{AuthFail,
				errors.Wrap(ErrNotExist, "User:GetInfo:GotSessionID:not exist: "+u.sessionID).Error()}
		} else if err != nil {
			return nil, AuthResult{InternalAuthErr,
				errors.Wrap(err, "User:GetInfo:GotSessionID").Error()}
		}
		u.token = uc.OAuthToken{AccessToken: token}
		u.state = GotToken
		fallthrough
	case GotToken:
		if state == GotToken {
			return nil, AuthResult{AuthSucc, ""}
		}
		cookieExtract := req.Header[textproto.CanonicalMIMEHeaderKey("cookie")]
		var info identity.UserInfo
		var err error
		//useToken, _ := strconv.ParseBool(req.Header.Get("USE-TOKEN"))
		// if len(cookieExtract) > 0 && !useToken {
		cookie := map[string][]string{"cookie": cookieExtract}
		user, err := u.ucUserAuth.GetCurrentUser(cookie)
		// } else {
		info, err = u.ucUserAuth.GetUserInfo(u.token)
		info.LastLoginAt = user.LastLoginAt
		// }
		if err != nil {
			return nil, AuthResult{Unauthed, err.Error()}
		}
		info.Token = u.token.AccessToken
		u.info = info
		u.state = GotInfo
		fallthrough
	case GotInfo:
		if state == GotInfo {
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
			var scopeinfo ScopeInfo
			scopeinfo.OrgID = orgID
			u.scopeInfo = scopeinfo
		}
		u.state = GotScopeInfo
		fallthrough
	case GotScopeInfo:
		if state == GotScopeInfo {
			return u.scopeInfo, AuthResult{AuthSucc, ""}
		}
	}
	panic("unreachable")
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

// 获取用户信息
func (u *User) GetInfo(req *http.Request) (identity.UserInfo, AuthResult) {
	info, authr := u.get(req, GotInfo)
	if authr.Code != AuthSucc {
		return identity.UserInfo{}, authr
	}
	return info.(identity.UserInfo), authr
}

// 获取用户orgID
func (u *User) GetScopeInfo(req *http.Request) (ScopeInfo, AuthResult) {
	scopeinfo, authr := u.get(req, GotScopeInfo)
	if authr.Code != AuthSucc {
		return ScopeInfo{}, authr
	}
	return scopeinfo.(ScopeInfo), authr
}

// return (token, expiredays, err)
func (u *User) Login(uccode string, redirectURI string, expiration time.Duration) (string, int, error) {
	u.ucUserAuth.RedirectURI = redirectURI
	otoken, err := u.ucUserAuth.Login(uccode)
	if err != nil {
		logrus.Error(err)
		return "", int(expiration), err
	}
	u.token = otoken
	userInfo, err := u.ucUserAuth.GetUserInfo(u.token)
	if err != nil {
		return "", int(expiration), err
	}
	u.info = userInfo
	u.state = GotInfo
	if err := u.storeSession(otoken.AccessToken, expiration); err != nil {
		err_ := errors.Wrap(err, "login: storeSession fail")
		logrus.Error(err_)
		return "", int(expiration), err_
	}
	return u.sessionID, int(expiration), nil
}

func (u *User) PwdLogin(username, password string, expiration time.Duration) (string, error) {
	otoken, err := u.ucUserAuth.PwdAuth(username, password)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	u.token = otoken
	userInfo, err := u.ucUserAuth.GetUserInfo(u.token)
	if err != nil {
		return "", err
	}
	u.info = userInfo
	u.state = GotInfo
	if err := u.storeSession(otoken.AccessToken, expiration); err != nil {
		err_ := errors.Wrap(err, "pwdlogin: storeSession fail")
		logrus.Error(err_)
		return "", err_
	}
	return u.sessionID, nil
}

func (u *User) storeSession(token string, expiration time.Duration) error {
	u.sessionID = genSessionID()
	_, err := u.redisCli.Set(MkSessionKey(u.sessionID), token, expiration).Result()
	if err != nil {
		err_ := errors.Wrap(err, "storeSession: store redis fail")
		return err_
	}
	return nil
}

func (u *User) Logout(req *http.Request) error {
	c := req.Context().Value("session")
	if c == nil {
		return fmt.Errorf("not provide session")
	}
	if _, err := u.redisCli.Del(MkSessionKey(c.(string))).Result(); err != nil {
		return err
	}
	return nil
}

func MkSessionKey(sessionID string) string {
	return "openapi:sessionid:" + sessionID
}

func genSessionID() string {
	return uuid.NewV4().String()
}
