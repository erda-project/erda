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
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

var UC_PWD_SECURITY_CONFIG_GET = apis.ApiSpec{
	Path:         "/api/users/actions/get-pwd-security-config",
	Scheme:       "http",
	Method:       "GET",
	Custom:       getPwdSecurityConfig,
	CheckLogin:   true,
	CheckToken:   true,
	ResponseType: apistructs.PwdSecurityConfigGetResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 查询密码安全配置",
}

func getPwdSecurityConfig(w http.ResponseWriter, r *http.Request) {
	operatorID, err := user.GetUserID(r)
	if err != nil {
		apierrors.ErrAdminUser.NotLogin().Write(w)
		return
	}

	if err := checkPermission(operatorID, apistructs.GetAction); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	token, err := auth.GetDiceClientToken()
	if err != nil {
		logrus.Errorf("failed to get token: %v", err)
		apierrors.ErrGetPwdSecurityConfig.InternalError(err).
			Write(w)
		return
	}

	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrGetPwdSecurityConfig)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	config, err := handleGetPwdSecurityConfig(token)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}
	httpserver.WriteData(w, config)
}

func handleGetPwdSecurityConfig(token ucauth.OAuthToken) (*apistructs.PwdSecurityConfig, error) {
	var resp struct {
		Success bool                          `json:"success"`
		Result  *apistructs.PwdSecurityConfig `json:"result"`
		Error   string                        `json:"error"`
	}
	r, err := httpclient.New().Get(discover.UC()).
		Path("/api/user/admin/pwd-security-config").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(err)
	}
	if !r.IsOK() {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(errors.New(resp.Error))
	}
	return resp.Result, nil
}

// checkPermission 检查权限
func checkPermission(userID user.ID, action string) error {
	permissionResult, err := bundle.New(bundle.WithCMDB()).CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.SysScope,
		ScopeID:  1,
		Resource: apistructs.OrgResource,
		Action:   action,
	})
	if err != nil {
		return err
	}
	if !permissionResult.Access {
		return apierrors.ErrAdminUser.AccessDenied()
	}
	return nil
}
