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

package endpoints

import (
	"context"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/cmdb/utils"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/desensitize"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// 限制批量查询最大用户数
const maxUserSize = 100

// ListUser 根据user id列表批量获取用户
func (e *Endpoints) ListUser(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var (
		users []utils.User
		err   error
	)

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListUser.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		// 从 Header 获取 OrgID
		var orgID uint64
		orgIDStr := r.Header.Get(httputil.OrgHeader)
		if orgIDStr != "" {
			orgID, err = strconv.ParseUint(orgIDStr, 10, 64)
			if err != nil {
				return apierrors.ErrListUser.InvalidParameter("orgId is invalid").ToResp(), nil
			}
		}

		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  orgID,
			Resource: apistructs.MemberResource,
			Action:   apistructs.CreateAction, // 只有企业管理员可以用这个接口，用createAction代替一下
		})
		if err != nil {
			return apierrors.ErrListUser.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrListUser.AccessDenied().ToResp(), nil
		}
	}

	keyword := r.URL.Query().Get("q")
	if keyword != "" { // 按关键字查找用户
		users, err = e.uc.FindUsersByKey(keyword)
		if err != nil {
			return apierrors.ErrListUser.InternalError(err).ToResp(), nil
		}
	} else if identityInfo.IsInternalClient() { // 按userID列表批量查找用户，这个接口不能暴露出去，以防通过暴力枚举userid获取userinfo
		// 检查请求参数
		params := r.URL.Query()
		if params == nil {
			return apierrors.ErrListUser.MissingParameter("user id").ToResp(), nil
		}
		userIDs := params["userID"]
		userIDs = strutil.DedupSlice(userIDs, true)
		if len(userIDs) > maxUserSize {
			return apierrors.ErrListUser.InvalidParameter("user id too much").ToResp(), nil
		}

		users, err = e.uc.FindUsers(userIDs)
		if err != nil {
			return apierrors.ErrListUser.InternalError(err).ToResp(), nil
		}
	}

	var plaintext bool
	plaintextStr := r.URL.Query().Get("plaintext")
	if plaintextStr == "true" {
		plaintext = true
	}
	// 用户信息转换
	userInfos := make([]apistructs.UserInfo, 0, len(users))
	for i := range users {
		userInfos = append(userInfos, *convertToUserInfo(&users[i], plaintext))
	}

	return httpserver.OkResp(apistructs.UserListResponseData{Users: userInfos})
}

// GetCurrentUser 获取当前登录用户信息
func (e *Endpoints) GetCurrentUser(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetUser.NotLogin().ToResp(), nil
	}

	// 获取用户详情
	user, err := e.uc.GetUser(userID.String())
	if err != nil {
		return apierrors.ErrGetUser.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(*convertToUserInfo(user, false))
}

func convertToUserInfo(user *utils.User, plaintext bool) *apistructs.UserInfo {
	if !plaintext {
		user.Phone = desensitize.Mobile(user.Phone)
		user.Email = desensitize.Email(user.Email)
	}
	return &apistructs.UserInfo{
		ID:     strconv.FormatUint(user.ID, 10),
		Name:   user.Name,
		Nick:   user.Nick,
		Avatar: user.AvatarURL,
		Phone:  user.Phone,
		Email:  user.Email,
	}
}
