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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/kratos"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	Jti          string `json:"jti"`
}

type UCUserAuth struct {
	UCHostFront  string
	UCHost       string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	bdl          *bundle.Bundle
}

const OryCompatibleClientId = "kratos"

func (a *UCUserAuth) oryEnabled() bool {
	// TODO: it's a hack
	return a.ClientID == OryCompatibleClientId
}

func (a *UCUserAuth) oryKratosAddr() string {
	return a.UCHost
}

func NewUCUserAuth(UCHostFront, UCHost, RedirectURI, ClientID, ClientSecret string) *UCUserAuth {
	bdl := bundle.New(bundle.WithErdaServer(), bundle.WithDOP())
	return &UCUserAuth{UCHostFront, UCHost, RedirectURI, ClientID, ClientSecret, bdl}
}

// 返回用户中心的登陆URL, 也就是浏览器请求的地址
// http://uc.terminus.io/oauth/authorize?response_type=code&client_id=dice&redirect_uri=http%3A%2F%2Fopenapi.test.terminus.io%2Flogincb&scope=public_profile
func (a *UCUserAuth) LoginURL(https bool) string {
	proto := "http"
	if https {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=public_profile", proto, a.UCHostFront, a.ClientID, url.QueryEscape(a.RedirectURI))
}

// (登陆) 从uc回调回来，会有uccode，用于得到token
func (a *UCUserAuth) Login(uccode string) (OAuthToken, error) {
	if a.oryEnabled() {
		// TODO:
		return OAuthToken{}, fmt.Errorf("not supported Login")
	}
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(a.ClientID+":"+a.ClientSecret))
	formBody := make(url.Values)
	formBody.Set("grant_type", "authorization_code")
	formBody.Set("code", uccode)
	formBody.Set("redirect_uri", a.RedirectURI)

	var body bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Post(a.UCHost).
		Path("/oauth/token").
		Header("Authorization", basic).
		FormBody(formBody).Do().Body(&body)
	if err != nil {
		err := errors.Wrap(err, "login: post /oauth/token fail")
		return OAuthToken{}, err
	}
	if !r.IsOK() {
		err := fmt.Errorf("login: /oauth/token statuscode: %d, body: %v", r.StatusCode(), body.String())
		return OAuthToken{}, err
	}
	var oauthToken OAuthToken
	if err := json.NewDecoder(&body).Decode(&oauthToken); err != nil {
		err := errors.Wrap(err, "login: decode fail")
		return OAuthToken{}, err
	}
	return oauthToken, nil
}

func (a *UCUserAuth) PwdAuth(username, password string) (OAuthToken, error) {
	if a.oryEnabled() {
		// TODO:
		return OAuthToken{}, fmt.Errorf("not supported PwdAuth")
	}
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(a.ClientID+":"+a.ClientSecret))
	formBody := make(url.Values)
	formBody.Set("grant_type", "password")
	formBody.Set("username", username)
	formBody.Set("password", password)
	formBody.Set("scope", "public_profile")
	var body bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Post(a.UCHost).
		Path("/oauth/token").
		Header("Authorization", basic).
		FormBody(formBody).Do().Body(&body)
	if err != nil {
		err := errors.Wrap(err, "pwdAuth: post /oauth/token fail")
		return OAuthToken{}, err
	}
	if !r.IsOK() {
		err := fmt.Errorf("pwdAuth: /oauth/token statuscode: %d, body: %v", r.StatusCode(), body.String())
		return OAuthToken{}, err
	}
	var oauthToken OAuthToken
	if err := json.NewDecoder(&body).Decode(&oauthToken); err != nil {
		err := errors.Wrap(err, "pwdAuth: decode fail")
		return OAuthToken{}, err
	}
	return oauthToken, nil
}

func (a *UCUserAuth) GetUserInfo(oauthToken OAuthToken) (common.UserInfo, error) {
	if a.oryEnabled() {
		// sessionID as token
		userInfo, err := kratos.Whoami(a.oryKratosAddr(), oauthToken.AccessToken)
		if err != nil {
			return userInfo, err
		}
		ucUserID, err := a.bdl.GetUcUserID(string(userInfo.ID))
		if err != nil {
			return userInfo, err
		}
		if ucUserID != "" {
			userInfo.ID = common.USERID(ucUserID)
		}
		return userInfo, err
	}
	bearer := "Bearer " + oauthToken.AccessToken
	var me bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(a.UCHost).
		Path("/api/oauth/me").
		Header("Authorization", bearer).Do().Body(&me)
	if err != nil {
		err := errors.Wrap(err, "GetUserInfo: /api/oauth/me fail")
		return common.UserInfo{}, err
	}
	if !r.IsOK() {
		err := fmt.Errorf("GetUserInfo: /api/oauth/me statuscode: %d, body: %v", r.StatusCode(), me.String())
		return common.UserInfo{}, err
	}
	var info common.UserInfo
	if err := json.NewDecoder(&me).Decode(&info); err != nil {
		err := errors.Wrap(err, "GetUserInfo: decode fail")
		return common.UserInfo{}, err
	}
	return info, nil
}

//	{
//	  "success": true,
//	  "result": {
//	    "id": 1000530,
//	    "tenantId": 1,
//	    "username": "u191-1019703192",
//	    "nickname": "",
//	    "avatar": "",
//	    "prefix": "86",
//	    "mobile": "15950552810",
//	    "email": "",
//	    "pwdExpireAt": null,
//	    "passwordExist": true,
//	    "enabled": true,
//	    "locked": false,
//	    "channel": "",
//	    "channelType": "",
//	    "source": "",
//	    "sourceType": "",
//	    "tag": "",
//	    "extra": null,
//	    "userDetail": null,
//	    "createdAt": "2020-09-21T09:11:26.000+0000",
//	    "updatedAt": "2020-12-15T04:01:02.000+0000",
//	    "lastLoginAt": "2020-12-15T04:01:02.000+0000",
//	    "pk": 1307970680503390208
//	  },
//	  "code": null,
//	  "args": null,
//	  "error": null,
//	  "sourceIp": null,
//	  "sourceStack": null
//	}
type CurrentUser struct {
	Success bool `json:"success"`
	Result  struct {
		ID          common.USERID `json:"id"`
		Email       string        `json:"email"`
		Mobile      string        `json:"mobile"`
		Username    string        `json:"username"`
		Nickname    string        `json:"nickname"`
		LastLoginAt uint64        `json:"lastLoginAt"`
	} `json:"result"`
	Error interface{} `json:"error"`
}

func (a *UCUserAuth) GetCurrentUser(headers http.Header) (common.UserInfo, error) {
	var me bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(a.UCHost).
		Path("/api/user/web/current-user").
		Headers(headers).Do().Body(&me)
	if err != nil {
		err := errors.Wrapf(err, "GetCurrentUser: /api/user/web/current-user: %+v", err)
		return common.UserInfo{}, err
	}
	if !r.IsOK() {
		err := fmt.Errorf("GetCurrentUser: /api/user/web/current-user statuscode: %d, body: %v", r.StatusCode(), me.String())
		return common.UserInfo{}, err
	}
	var info CurrentUser
	d := json.NewDecoder(&me)
	if err := d.Decode(&info); err != nil {
		buffered, _ := io.ReadAll(d.Buffered())
		err := errors.Wrapf(err, "GetCurrentUser: decode fail: %v", string(buffered))
		return common.UserInfo{}, err
	}
	if info.Result.ID == "" {
		return common.UserInfo{}, fmt.Errorf("not login")
	}
	t := time.Unix(int64(info.Result.LastLoginAt/1e3), 0).Format("2006-01-02 15:04:05")
	return common.UserInfo{
		ID:          info.Result.ID,
		Email:       info.Result.Email,
		Phone:       info.Result.Mobile,
		UserName:    info.Result.Username,
		NickName:    info.Result.Nickname,
		LastLoginAt: t,
	}, nil
}

func HandlePagingUsers(req *apistructs.UserPagingRequest, token OAuthToken) (*common.UserPaging, error) {
	if token.TokenType == OryCompatibleClientId {
		return kratos.HandlePagingUsers(req, token.AccessToken)
	}
	v := httpclient.New().Get(discover.UC()).Path("/api/user/admin/paging").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if req.Name != "" {
		v.Param("username", req.Name)
	}
	if req.Nick != "" {
		v.Param("nickname", req.Nick)
	}
	if req.Phone != "" {
		v.Param("mobile", req.Phone)
	}
	if req.Email != "" {
		v.Param("email", req.Email)
	}
	if req.Locked != nil {
		v.Param("locked", strconv.Itoa(*req.Locked))
	}
	if req.Source != "" {
		v.Param("source", req.Source)
	}
	if req.PageNo > 0 {
		v.Param("pageNo", strconv.Itoa(req.PageNo))
	}
	if req.PageSize > 0 {
		v.Param("pageSize", strconv.Itoa(req.PageSize))
	}
	// 批量查询用户
	var resp struct {
		Success bool               `json:"success"`
		Result  *common.UserPaging `json:"result"`
		Error   string             `json:"error"`
	}
	r, err := v.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("internal status code: %v", r.StatusCode())
	}
	if !resp.Success {
		return nil, errors.New(resp.Error)
	}
	return resp.Result, nil
}
