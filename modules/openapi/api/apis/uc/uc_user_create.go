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

package uc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

var UC_USER_CREATE = apis.ApiSpec{
	Path:         "/api/users",
	Scheme:       "http",
	Method:       "POST",
	Custom:       createUsers,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserCreateRequest{},
	ResponseType: apistructs.UserCreateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 用户创建",
}

func createUsers(w http.ResponseWriter, r *http.Request) {
	operatorID, err := user.GetUserID(r)
	if err != nil {
		apierrors.ErrAdminUser.NotLogin().Write(w)
		return
	}

	if err := checkPermission(operatorID, apistructs.CreateAction); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	token, err := auth.GetDiceClientToken()
	if err != nil {
		logrus.Errorf("failed to get token: %v", err)
		apierrors.ErrCreateUser.InternalError(err).
			Write(w)
		return
	}

	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrCreateUser)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	// get req
	var items []apistructs.UserCreateItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		apierrors.ErrCreateUser.InvalidParameter(err).
			Write(w)
		return
	}
	if len(items) == 0 {
		apierrors.ErrCreateUser.InvalidParameter("no users to create").
			Write(w)
	}
	req := apistructs.UserCreateRequest{Users: items}

	// handle
	if err := handleCreateUsers(&req, operatorID.String(), token); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	httpserver.WriteData(w, nil)
}

func handleCreateUsers(req *apistructs.UserCreateRequest, operatorID string, token ucauth.OAuthToken) error {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	users := make([]createUserItem, len(req.Users))
	for i, u := range req.Users {
		users[i] = convertCreateUserItem(&u)
	}
	reqBody := createUser{Users: users}
	r, err := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/batch-create-user").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Header("operatorId", operatorID).
		JSONBody(&reqBody).
		Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke create user, (%v)", err)
		return apierrors.ErrCreateUser.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to create user, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrCreateUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to create user: %+v", resp.Error)
		return apierrors.ErrCreateUser.InternalError(errors.New(resp.Error))
	}
	return nil
}

type createUser struct {
	Users []createUserItem `json:"users"`
}

type createUserItem struct {
	Username    string      `json:"username,omitempty"`    // 用户名
	Nickname    string      `json:"nickname,omitempty"`    // 昵称
	Mobile      string      `json:"mobile,omitempty"`      //
	Email       string      `json:"email,omitempty"`       // 邮箱
	Password    string      `json:"password"`              // 密码
	Avatar      string      `json:"avatar,omitempty"`      // 头像
	Channel     string      `json:"channel,omitempty"`     // 注册渠道
	ChannelType string      `json:"channelType,omitempty"` // 渠道类型
	Extra       interface{} `json:"extra,omitempty"`       //
	Source      string      `json:"source,omitempty"`      // 用户来源
	SourceType  string      `json:"sourceType,omitempty"`  // 来源类型
	Tag         string      `json:"tag,omitempty"`         // 标签
	UserDetail  interface{} `json:"userDetail,omitempty"`  // 用户详情
}

func convertCreateUserItem(item *apistructs.UserCreateItem) createUserItem {
	return createUserItem{
		Username: item.Name,
		Nickname: item.Nick,
		Mobile:   item.Phone,
		Email:    item.Email,
		Password: item.Password,
	}
}
