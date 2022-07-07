// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/kratos"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

func HandlePagingUsers(req *apistructs.UserPagingRequest, token uc.OAuthToken) (*common.UserPaging, error) {
	if token.TokenType == uc.OryCompatibleClientId {
		return kratos.HandlePagingUsers(req, token.AccessToken)
	}
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
		Success bool               `json:"success"`
		Result  *common.UserPaging `json:"result"`
		Error   string             `json:"error"`
	}
	r, err := v.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("internal status code: %v", r.StatusCode())
	}
	if !resp.Success {
		return nil, errors.New(resp.Error)
	}
	return resp.Result, nil
}

func ConvertToUserInfoExt(user *common.UserPaging) *apistructs.UserPagingData {
	var ret apistructs.UserPagingData
	ret.Total = user.Total
	ret.List = make([]apistructs.UserInfoExt, 0)
	for _, u := range user.Data {
		ret.List = append(ret.List, apistructs.UserInfoExt{
			UserInfo: apistructs.UserInfo{
				ID:          strutil.String(u.Id),
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
