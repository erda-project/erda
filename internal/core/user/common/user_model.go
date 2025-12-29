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

package common

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

type User struct {
	ID        string `json:"user_id"`
	Name      string `json:"username"`
	Nick      string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Phone     string `json:"phone_number"`
	Email     string `json:"email"`
	State     string `json:"state"`
}

type UserInfo struct {
	ID               USERID          `json:"id"`
	Token            string          `json:"token"`
	Email            string          `json:"email"`
	EmailExist       bool            `json:"emailExist"`
	PasswordExist    bool            `json:"passwordExist"`
	PhoneExist       bool            `json:"phoneExist"`
	Birthday         string          `json:"birthday"`
	PasswordStrength int             `json:"passwordStrength"`
	Phone            string          `json:"phone"`
	AvatarUrl        string          `json:"avatarUrl"`
	UserName         string          `json:"username"`
	NickName         string          `json:"nickName"`
	Enabled          bool            `json:"enabled"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	LastLoginAt      string          `json:"lastLoginAt"`
	SessionRefresh   *SessionRefresh `json:"sessionRefresh,omitempty"`
}

type SessionRefresh struct {
	Token  string       `json:"token"`
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

type USERID string

func (u USERID) String() string { return string(u) }

// maybe int or string, unmarshal them to string(USERID)
func (u *USERID) UnmarshalJSON(b []byte) error {
	var intid int
	if err := json.Unmarshal(b, &intid); err != nil {
		var stringid string
		if err := json.Unmarshal(b, &stringid); err != nil {
			return err
		}
		*u = USERID(stringid)
		return nil
	}
	*u = USERID(strconv.Itoa(intid))
	return nil
}

type UserPaging struct {
	Data  []UserInPaging `json:"data"`
	Total int            `json:"total"`
}

// userInPaging 用户中心分页用户数据结构
type UserInPaging struct {
	Id            interface{} `json:"id"`            // 主键
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

const SystemOperator = "system"

var SystemUser = User{
	ID:   SystemOperator,
	Name: SystemOperator,
	Nick: SystemOperator,
}

//type CookieConfig struct {
//	Name     string `json:"name,omitempty"`
//	Domain   string `json:"domain,omitempty"`
//	Path     string `json:"path,omitempty"`
//	MaxAge   int    `json:"maxAge,omitempty"`
//	Secure   bool   `json:"secure,omitempty"`
//	HttpOnly bool   `json:"httpOnly,omitempty"`
//	SameSite string `json:"sameSite,omitempty"`
//}

func ToPbUser(user User) *pb.User {
	return &pb.User{
		ID:        user.ID,
		Name:      user.Name,
		Nick:      user.Nick,
		AvatarURL: user.AvatarURL,
		Phone:     user.Phone,
		Email:     user.Email,
		State:     user.State,
	}
}

func NewUserInfoFromDTO(dto *apistructs.UserInfoDto) *UserInfo {
	if dto == nil {
		return nil
	}
	return &UserInfo{
		ID:         USERID(strutil.String(dto.UserID)),
		Email:      dto.Email,
		EmailExist: dto.Email != "",
		Phone:      dto.Phone,
		PhoneExist: dto.Phone != "",
		AvatarUrl:  dto.AvatarURL,
		UserName:   dto.Username,
		NickName:   dto.NickName,
	}
}
