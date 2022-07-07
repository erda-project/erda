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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/strutil"
)

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
