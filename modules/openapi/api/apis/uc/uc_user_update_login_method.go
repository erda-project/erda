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

var UC_USER_UPDATE_LOGIN_METHOD = apis.ApiSpec{
	Path:         "/api/users/<userID>/actions/update-login-method",
	Scheme:       "http",
	Method:       "POST",
	Custom:       updateLoginMethod,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserUpdateLoginMethodRequest{},
	ResponseType: apistructs.UserUpdateLoginMethodResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 修改用户登录方式",
}

func updateLoginMethod(w http.ResponseWriter, r *http.Request) {
	operatorID, err := user.GetUserID(r)
	if err != nil {
		apierrors.ErrAdminUser.NotLogin().Write(w)
		return
	}

	if err := checkPermission(operatorID, apistructs.UpdateAction); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	token, err := auth.GetDiceClientToken()
	if err != nil {
		logrus.Errorf("failed to get token: %v", err)
		apierrors.ErrBatchFreezeUser.InternalError(err).
			Write(w)
		return
	}
	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrBatchFreezeUser)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	// get req
	var req apistructs.UserUpdateLoginMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	userID := strutil.Split(r.URL.Path, "/", true)[2]
	req.ID = userID

	// handle
	if err := handleUpdateLoginMethod(req, operatorID.String(), token); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	httpserver.WriteData(w, nil)
}

func handleUpdateLoginMethod(req apistructs.UserUpdateLoginMethodRequest, operatorID string, token ucauth.OAuthToken) error {
	var resp struct {
		Success bool   `json:"success"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
	}
	r, err := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/change-full-info").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Header("operatorID", operatorID).
		JSONBody(&req).
		Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke change user login method, (%v)", err)
		return apierrors.ErrUpdateLoginMethod.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to change user login method, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrUpdateLoginMethod.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to change user login method: %+v", resp.Error)
		return apierrors.ErrUpdateLoginMethod.InternalError(errors.New(resp.Error))
	}
	return nil
}
