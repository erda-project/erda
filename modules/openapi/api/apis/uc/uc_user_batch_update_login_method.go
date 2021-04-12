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
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var UC_USER_BATCH_UPDATE_LOGIN_METHOD = apis.ApiSpec{
	Path:         "/api/users/actions/batch-update-login-method",
	Scheme:       "http",
	Method:       "POST",
	Custom:       batchUpdateLoginMethod,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserBatchUpdateLoginMethodRequest{},
	ResponseType: apistructs.UserBatchUpdateLoginMethodResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 批量修改用户登录方式",
}

func batchUpdateLoginMethod(w http.ResponseWriter, r *http.Request) {
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
	var req apistructs.UserBatchUpdateLoginMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	source := req.Source

	operatorIDStr := operatorID.String()
	for _, v := range req.UserIDs {
		r := apistructs.UserUpdateLoginMethodRequest{
			ID:     v,
			Source: source,
		}
		// handle
		if err := handleUpdateLoginMethod(r, operatorIDStr, token); err != nil {
			errorresp.ErrWrite(err, w)
			return
		}
	}

	httpserver.WriteData(w, nil)
}
