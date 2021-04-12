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
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/openapi/api/apis"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apierrors"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
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

	data, err := handlePagingUsers(req, token)
	if err != nil {
		errorresp.ErrWrite(err, w)
		return
	}

	httpserver.WriteData(w, convertToUserInfoExt(data))
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

func handlePagingUsers(req *apistructs.UserPagingRequest, token ucauth.OAuthToken) (*userPaging, error) {
	v := httpclient.New().Get(discover.UC()).Path("/api/user/admin/paging").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if req.Name != "" {
		v.Param("username", req.Name)
	}
	if req.Nick != "" {
		v.Param("nickname", req.Nick)
	}
	if req.Phone != "" {
		v.Param("mobile", req.Phone)
	}
	if req.Email != "" {
		v.Param("email", req.Email)
	}
	if req.Locked != nil {
		v.Param("locked", strconv.Itoa(*req.Locked))
	}
	if req.Source != "" {
		v.Param("source", req.Source)
	}
	if req.PageNo > 0 {
		v.Param("pageNo", strconv.Itoa(req.PageNo))
	}
	if req.PageSize > 0 {
		v.Param("pageSize", strconv.Itoa(req.PageSize))
	}
	// 批量查询用户
	var resp struct {
		Success bool        `json:"success"`
		Result  *userPaging `json:"result"`
		Error   string      `json:"error"`
	}
	r, err := v.Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrListUser.InternalError(err)
	}
	if !r.IsOK() {
		return nil, apierrors.ErrListUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return nil, apierrors.ErrListUser.InternalError(errors.New(resp.Error))
	}
	return resp.Result, nil
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

func convertToUserInfoExt(user *userPaging) *apistructs.UserPagingData {
	var ret apistructs.UserPagingData
	ret.Total = user.Total
	ret.List = make([]apistructs.UserInfoExt, 0)
	for _, u := range user.Data {
		ret.List = append(ret.List, apistructs.UserInfoExt{
			UserInfo: apistructs.UserInfo{
				ID:          strconv.FormatUint(u.Id, 10),
				Name:        u.Username,
				Nick:        u.Nickname,
				Avatar:      u.Avatar,
				Phone:       u.Mobile,
				Email:       u.Email,
				LastLoginAt: time.Time(u.LastLoginAt).Format("2006-01-02 15:04:05"),
				PwdExpireAt: time.Time(u.PwdExpireAt).Format("2006-01-02 15:04:05"),
				Source:      u.Source,
			},
			Locked: u.Locked,
		})
	}
	return &ret
}

// userInPaging 用户中心分页用户数据结构
type userInPaging struct {
	Id            uint64      `json:"id"`            // 主键
	Avatar        string      `json:"avatar"`        // 头像
	Username      string      `json:"username"`      // 用户名
	Nickname      string      `json:"nickname"`      // 昵称
	Mobile        string      `json:"mobile"`        // 手机号
	Email         string      `json:"email"`         // 邮箱
	Enabled       bool        `json:"enabled"`       // 是否启用
	UserDetail    interface{} `json:"userDetail"`    // 用户详细信息
	Locked        bool        `json:"locked"`        // 冻结FLAG(0:NOT,1:YES)
	PasswordExist bool        `json:"passwordExist"` // 密码是否存在
	PwdExpireAt   timestamp   `json:"pwdExpireAt"`   // 过期时间
	Extra         interface{} `json:"extra"`         // 扩展字段
	Source        string      `json:"source"`        // 用户来源
	SourceType    string      `json:"sourceType"`    // 来源类型
	Tag           string      `json:"tag"`           // 标签
	Channel       string      `json:"channel"`       // 注册渠道
	ChannelType   string      `json:"channelType"`   // 渠道类型
	TenantId      int         `json:"tenantId"`      // 租户ID
	CreatedAt     timestamp   `json:"createdAt"`     // 创建时间
	UpdatedAt     timestamp   `json:"updatedAt"`     // 更新时间
	LastLoginAt   timestamp   `json:"lastLoginAt"`   // 最后登录时间
}

// millisecond epoch
type timestamp time.Time

func (t timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *timestamp) UnmarshalJSON(s []byte) (err error) {
	r := strings.Replace(string(s), `"`, ``, -1)
	if r == "null" {
		return
	}

	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q/1000, 0)
	return
}

type userPaging struct {
	Data  []userInPaging `json:"data"`
	Total int            `json:"total"`
}
