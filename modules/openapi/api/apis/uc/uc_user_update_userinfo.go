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
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

var UC_USER_UPDATE_USERINFO = apis.ApiSpec{
	Path:         "/api/user/admin/update-userinfo",
	Scheme:       "http",
	Method:       "PUT",
	Custom:       updateUserInfo,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.UserUpdateInfoRequset{},
	ResponseType: apistructs.UserUpdateInfoResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 更新用户信息",
}

var bdl = bundle.New(bundle.WithCMDB())

func updateUserInfo(w http.ResponseWriter, r *http.Request) {
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
		apierrors.ErrUpdateUserInfo.InternalError(err).
			Write(w)
		return
	}

	// check login & permission
	_, err = mustManageUsersPerm(r, apierrors.ErrUpdateUserInfo)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	// get req
	var req apistructs.UserUpdateInfoRequset
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.ErrUpdateUserInfo.InvalidParameter(err).
			Write(w)
		return
	}

	ctx := &spec.AuditContext{
		UserID:    operatorID.String(),
		Bundle:    bdl,
		BeginTime: time.Now(),
		Result:    apistructs.SuccessfulResult,
		UserAgent: r.Header.Get("User-Agent"),
		ClientIP:  GetRealIP(r),
		Request:   r,
	}
	oldUser, err := ctx.Bundle.ListUsers(apistructs.UserListRequest{UserIDs: []string{req.UserID}})
	if err != nil {
		logrus.Errorf("get old user for audit err: %v", err)
	}

	// handle
	if err := handleUpdateUserInfo(&req, operatorID.String(), token); err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	go func() {
		if err := createAudit(ctx, req, oldUser); err != nil {
			logrus.Errorf("update userinfo by admin create audit err: %v", err)
		}
	}()

	httpserver.WriteData(w, nil)
}

type ucUpdateUserInfoReq struct {
	ID       string `json:"id,omitempty"`
	UserName string `json:"username,omitempty"`
	Nick     string `json:"nickname,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
}

func handleUpdateUserInfo(req *apistructs.UserUpdateInfoRequset, operatorID string, token ucauth.OAuthToken) error {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	reqBody, err := getReqBody(req)
	if err != nil {
		return err
	}
	r, err := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/change-full-info").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Header("operatorId", operatorID).
		JSONBody(reqBody).
		Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke update userinfo, (%v)", err)
		return apierrors.ErrUpdateUserInfo.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to update userinfo, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrUpdateUserInfo.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to update userinfo: %+v", resp.Error)
		return apierrors.ErrUpdateUserInfo.InternalError(errors.New(resp.Error))
	}

	return nil
}

func getReqBody(req *apistructs.UserUpdateInfoRequset) (*ucUpdateUserInfoReq, error) {
	if req.UserID == "" {
		return nil, errors.New("user id is empty")
	}
	ucReq := &ucUpdateUserInfoReq{
		ID: req.UserID,
	}
	if req.Nick != "" {
		ucReq.Nick = req.Nick
	}
	if req.Name != "" {
		ucReq.UserName = req.Name
	}
	if req.Mobile != "" {
		ucReq.Mobile = req.Mobile
	}
	if req.Email != "" {
		ucReq.Email = req.Email
	}

	return ucReq, nil
}

// createAudit 创建审计
func createAudit(ctx *spec.AuditContext, req apistructs.UserUpdateInfoRequset, oldUser *apistructs.UserListResponseData) error {
	operator, err := ctx.Bundle.ListUsers(apistructs.UserListRequest{UserIDs: []string{ctx.UserID}})
	if err != nil {
		return err
	}
	ctx.EndTime = time.Now()

	audit := &apistructs.Audit{
		ScopeType: apistructs.SysScope,
		ScopeID:   1,
		Context: map[string]interface{}{"userName": oldUser.Users[0].Name, "nickName": oldUser.Users[0].Nick,
			"users": operator.Users},
	}

	if req.Email != oldUser.Users[0].Email {
		audit.TemplateName = apistructs.UpdateUserMailTemplate
		if err := ctx.CreateAudit(audit); err != nil {
			return err
		}
	}

	if req.Mobile != oldUser.Users[0].Phone {
		audit.TemplateName = apistructs.UpdateUserTelTemplate
		if err := ctx.CreateAudit(audit); err != nil {
			return err
		}
	}

	return nil
}

// GetRealIP 获取真实ip
func GetRealIP(request *http.Request) string {
	ra := request.RemoteAddr
	if ip := request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}
