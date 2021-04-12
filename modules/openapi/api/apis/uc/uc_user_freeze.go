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

var UC_USER_FREEZE = apis.ApiSpec{
	Path:         "/api/users/<userID>/actions/freeze",
	Scheme:       "http",
	Method:       "PUT",
	Custom:       freezeUser,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserFreezeRequest{},
	ResponseType: apistructs.UserFreezeResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 用户冻结",
}

func freezeUser(w http.ResponseWriter, r *http.Request) {
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
		apierrors.ErrFreezeUser.InternalError(err).
			Write(w)
		return
	}

	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrFreezeUser)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	// get req
	userID := strutil.Split(r.URL.Path, "/", true)[2]
	logrus.Debugf("to freeze userID: %v", userID)

	// handle
	if err := handleFreezeUser(userID, operatorID.String(), token); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	httpserver.WriteData(w, nil)
}

func handleFreezeUser(userID, operatorID string, token ucauth.OAuthToken) error {
	var resp struct {
		Success bool   `json:"success"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
	}
	r, err := httpclient.New().Put(discover.UC()).
		Path("/api/user/admin/freeze/"+userID).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Header("operatorId", operatorID).
		Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrFreezeUser.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrFreezeUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return apierrors.ErrFreezeUser.InternalError(errors.New(resp.Error))
	}
	return nil
}
