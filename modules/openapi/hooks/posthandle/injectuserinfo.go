// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package posthandle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/pkg/desensitize"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

var (
	once      sync.Once
	tokenAuth *ucauth.UCTokenAuth
	client    *httpclient.HTTPClient

	// 用于 ut
	testHookUC func(*[]User)
)

// USERID user id 可能是 int 或 string
type USERID string

// UnmarshalJSON maybe int or string, unmarshal them to string(USERID)
func (u *USERID) UnmarshalJSON(b []byte) error {
	var intid int
	if err := json.Unmarshal(b, &intid); err != nil {
		var stringid string
		if err := json.Unmarshal(b, &stringid); err != nil {
			return err
		}
		*u = USERID(stringid)
		return nil
	}
	*u = USERID(strconv.Itoa(intid))
	return nil
}

// User 用户中心用户数据结构
type User struct {
	ID        USERID `json:"user_id"`
	Name      string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Phone     string `json:"phone_number"`
	Email     string `json:"email"`
	Nick      string `json:"nickname"`
}

// InjectUserInfo 对 resp 的 body 中注入 userinfo
func InjectUserInfo(resp *http.Response, needDesensitize bool) error {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var bodyjson map[string]interface{}
	if err := json.Unmarshal(content, &bodyjson); err != nil {
		// response body 结构不是 JSON Object, 忽略这种情况即可
		resp.Body = ioutil.NopCloser(bytes.NewReader(content))
		return nil
	}
	userIDs, ok := bodyjson["userIDs"]
	if !ok {
		resp.Body = ioutil.NopCloser(bytes.NewReader(content))
		return nil
	}
	userInfoMap := map[string]apistructs.UserInfo{}
	switch v := userIDs.(type) {
	case []interface{}:
		ids := make([]string, 0, len(v))
		for _, id := range v {
			idstr, ok := id.(string)
			if !ok {
				resp.Body = ioutil.NopCloser(bytes.NewReader(content))
				return fmt.Errorf("failed to inject userinfo, invalid type of userIDs, id: %v", id)
			}
			ids = append(ids, idstr)
		}
		var err error
		if userInfoMap, err = GetUsers(ids, needDesensitize); err != nil {
			resp.Body = ioutil.NopCloser(bytes.NewReader(content))
			return err
		}
	case []string:
		var err error
		if userInfoMap, err = GetUsers(v, needDesensitize); err != nil {
			resp.Body = ioutil.NopCloser(bytes.NewReader(content))
			return err
		}
	}
	// inject to response body
	bodyjson["userInfo"] = userInfoMap
	newbody, err := json.Marshal(bodyjson)
	if err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(content))
		return err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(newbody))
	resp.Header["Content-Length"] = []string{fmt.Sprint(len(newbody))}
	return nil
}

func GetUsers(IDs []string, needDesensitize bool) (map[string]apistructs.UserInfo, error) {
	once.Do(func() {
		if testHookUC == nil {
			var err error
			tokenAuth, err = ucauth.NewUCTokenAuth(discover.UC(), conf.UCClientID(), conf.UCClientSecret())
			if err != nil {
				panic(err)
			}
		}
		client = httpclient.New(httpclient.WithDialerKeepAlive(10 * time.Second))
	})
	var (
		err   error
		token ucauth.OAuthToken
	)
	if testHookUC == nil {
		if token, err = tokenAuth.GetServerToken(false); err != nil {
			return nil, err
		}
	}
	parts := make([]string, len(IDs))
	for i := range IDs {
		parts[i] = strutil.Concat("user_id:", IDs[i])
	}
	query := strutil.Join(parts, " OR ")
	b := []User{}
	var body bytes.Buffer
	if testHookUC != nil {
		testHookUC(&b)
	} else {
		f := func() (*httpclient.Response, error) {
			resp, err := client.Get(discover.UC()).Path("/api/open/v1/users").Param("query", query).
				Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).Do().Body(&body)
			return resp, err
		}
		resp, err := f()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() > 400 {
			// token 过期, 重新申请, 但是只会重试一次
			tokenAuth.ExpireServerToken()
			if token, err = tokenAuth.GetServerToken(true); err != nil {
				return nil, err
			}
			resp2, err := f()
			if err != nil {
				return nil, err
			}
			if !resp2.IsOK() {
				return nil, fmt.Errorf("failed to get users(token refreshed), status code: %d",
					resp.StatusCode())
			}

		} else if !resp.IsOK() {
			return nil, fmt.Errorf("failed to get users, status code: %d", resp.StatusCode())
		}
		if err := json.NewDecoder(&body).Decode(&b); err != nil {
			return nil, err
		}
	}
	users := make(map[string]apistructs.UserInfo, len(b))
	// 是否需要脱敏处理
	if needDesensitize {
		for i := range b {
			users[string(b[i].ID)] = apistructs.UserInfo{
				ID:     "",
				Email:  desensitize.Email(b[i].Email),
				Phone:  desensitize.Mobile(b[i].Phone),
				Avatar: b[i].AvatarURL,
				Name:   desensitize.Name(b[i].Name),
				Nick:   desensitize.Name(b[i].Nick),
			}
		}
	} else {
		for i := range b {
			users[string(b[i].ID)] = apistructs.UserInfo{
				ID:     string(b[i].ID),
				Email:  b[i].Email,
				Phone:  b[i].Phone,
				Avatar: b[i].AvatarURL,
				Name:   b[i].Name,
				Nick:   b[i].Nick,
			}
		}
	}
	for _, userID := range IDs {
		_, exist := users[userID]
		if !exist {
			users[userID] = apistructs.UserInfo{
				ID:     userID,
				Email:  "",
				Phone:  "",
				Avatar: "",
				Name:   "用户已注销",
				Nick:   "用户已注销",
			}
		}
	}
	return users, nil
}
