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

var UC_USER_BATCH_UNFREEZE = apis.ApiSpec{
	Path:         "/api/users/actions/batch-unfreeze",
	Scheme:       "http",
	Method:       "PUT",
	Custom:       batchUnFreezeUser,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserBatchUnFreezeRequest{},
	ResponseType: apistructs.UserBatchUnFreezeResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 批量用户解冻",
}

func batchUnFreezeUser(w http.ResponseWriter, r *http.Request) {
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
		apierrors.ErrBatchUnFreezeUser.InternalError(err).
			Write(w)
		return
	}
	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrBatchUnFreezeUser)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	// get req
	var req apistructs.UserBatchUnFreezeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	userIDs := req.UserIDs
	logrus.Debugf("to unfreeze userID: %v", userIDs)
	// handle
	operatorIDStr := operatorID.String()
	for _, userID := range userIDs {
		if err := handleUnfreezeUser(userID, operatorIDStr, token); err != nil {
			errorresp.ErrWrite(err, w)
			return
		}
	}
	httpserver.WriteData(w, nil)
}
