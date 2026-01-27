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

package uc

import "time"

// Response uc standard response dto
type Response[T any] struct {
	ResponseMeta
	Result T `json:"result"`
}

type ResponseMeta struct {
	Success *bool  `json:"success"`
	Error   string `json:"error"`
}

type GetUser struct {
	ID          int    `json:"user_id"`
	Name        string `json:"username"`
	Nick        string `json:"nickname"`
	AvatarURL   string `json:"avatar_url"`
	Phone       string `json:"phone_number"`
	Email       string `json:"email"`
	LastLoginAt uint64 `json:"lastLoginAt"`
}

type LoginTypes struct {
	RegistryType []string `json:"registryType"`
}

type PwdSecurityConfig struct {
	CaptchaChallengeNumber   int64 `json:"captchaChallengeNumber"`
	ContinuousPwdErrorNumber int64 `json:"continuousPwdErrorNumber"`
	MaxPwdErrorNumber        int64 `json:"maxPwdErrorNumber"`
	ResetPassWordPeriod      int64 `json:"resetPassWordPeriod"`
}

type UpdateLoginMethodRequest struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

type UpdateUserInfoRequest struct {
	ID       string `json:"id,omitempty"`
	UserName string `json:"username,omitempty"`
	Nick     string `json:"nickname,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
}

type CreateUserRequest struct {
	Users []*CreateUserItem `json:"users"`
}

type CreateUserItem struct {
	Username    string      `json:"username,omitempty"`
	Nickname    string      `json:"nickname,omitempty"`
	Mobile      string      `json:"mobile,omitempty"`
	Email       string      `json:"email,omitempty"`
	Password    string      `json:"password"`
	Avatar      string      `json:"avatar,omitempty"`
	Channel     string      `json:"channel,omitempty"`
	ChannelType string      `json:"channelType,omitempty"`
	Extra       interface{} `json:"extra,omitempty"`
	Source      string      `json:"source,omitempty"`
	SourceType  string      `json:"sourceType,omitempty"`
	Tag         string      `json:"tag,omitempty"`
	UserDetail  interface{} `json:"userDetail,omitempty"`
}

type UserPaging struct {
	Data  []*UserInPaging `json:"data"`
	Total int64           `json:"total"`
}
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
	PwdExpireAt   time.Time   `json:"pwdExpireAt"`   // 过期时间
	Extra         interface{} `json:"extra"`         // 扩展字段
	Source        string      `json:"source"`        // 用户来源
	SourceType    string      `json:"sourceType"`    // 来源类型
	Tag           string      `json:"tag"`           // 标签
	Channel       string      `json:"channel"`       // 注册渠道
	ChannelType   string      `json:"channelType"`   // 渠道类型
	TenantId      int         `json:"tenantId"`      // 租户ID
	CreatedAt     time.Time   `json:"createdAt"`     // 创建时间
	UpdatedAt     time.Time   `json:"updatedAt"`     // 更新时间
	LastLoginAt   time.Time   `json:"lastLoginAt"`   // 最后登录时间
}
