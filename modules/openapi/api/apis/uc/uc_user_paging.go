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
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/ucauth"
)

var UC_USER_PAGING = apis.ApiSpec{
	Path:         "/api/users/actions/paging",
	Scheme:       "http",
	Method:       "GET",
	Custom:       pagingUsers,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserPagingRequest{},
	ResponseType: apistructs.UserPagingResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 用户分页",
}

func pagingUsers(w http.ResponseWriter, r *http.Request) {
	operatorID, err := user.GetUserID(r)
	if err != nil {
		apierrors.ErrAdminUser.NotLogin().Write(w)
		return
	}

	if err := checkPermission(operatorID, apistructs.GetAction); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	// TODO: nothing to do when oryEnabled
	token, err := auth.GetDiceClientToken()
	if err != nil {
		logrus.Errorf("failed to get token: %v", err)
		apierrors.ErrListUser.InternalError(err).
			Write(w)
		return
	}

	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrListUser)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	req, err := getPagingUsersReq(r)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	data, err := ucauth.HandlePagingUsers(req, token)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	httpserver.WriteData(w, ucauth.ConvertToUserInfoExt(data))
}

func getPagingUsersReq(r *http.Request) (*apistructs.UserPagingRequest, error) {
	req := apistructs.UserPagingRequest{
		Name:  r.URL.Query().Get("name"),
		Nick:  r.URL.Query().Get("nick"),
		Phone: r.URL.Query().Get("phone"),
		Email: r.URL.Query().Get("email"),
	}
	v := r.URL.Query().Get("locked")
	if v != "" {
		var locked int
		if v == "true" {
			locked = 1
		} else if v == "false" {
			locked = 0
		} else {
			return nil, apierrors.ErrListUser.InvalidParameter("invalid parameter locked")
		}
		req.Locked = &locked
	}
	v = r.URL.Query().Get("source")
	if v != "" {
		req.Source = v
	}
	v = r.URL.Query().Get("pageNo")
	if v != "" {
		pageNo, err := strconv.Atoi(v)
		if err != nil {
			return nil, apierrors.ErrListUser.InvalidParameter(err)
		}
		req.PageNo = pageNo
	}
	v = r.URL.Query().Get("pageSize")
	if v != "" {
		pageSize, err := strconv.Atoi(v)
		if err != nil {
			return nil, apierrors.ErrListUser.InvalidParameter(err)
		}
		req.PageSize = pageSize
	}
	return &req, nil
}

func mustManageUsersPerm(r *http.Request, errBuilder *errorresp.APIError) (string, error) {
	// check login
	userID, err := user.GetUserID(r)
	if err != nil {
		logrus.Errorf("failed to get userID, (%v)", err)
		return "", errBuilder.NotLogin()
	}
	// check permission
	if !isManageUsersPerm(userID) {
		return "", errBuilder.AccessDenied()
	}
	return userID.String(), nil
}

func isManageUsersPerm(userID user.ID) bool {
	// TODO: check permission
	return true
}
