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

import (
	"encoding/json"
	"time"
)

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

type TimestampMs struct {
	time.Time
}

func (t *TimestampMs) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var ts int64
	if err := json.Unmarshal(data, &ts); err == nil {
		t.Time = time.Unix(ts/1000, (ts%1000)*1e6)
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			t.Time = time.Time{}
			return nil
		}
		parsed, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	}

	return json.Unmarshal(data, &t.Time)
}

func (t TimestampMs) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time.UnixMilli())
}

type UserPaging struct {
	Data  []*UserDto `json:"data"`
	Total int64      `json:"total"`
}
type UserDto struct {
	Id            interface{}  `json:"id"`
	Avatar        string       `json:"avatar"`
	Username      string       `json:"username"`
	Nickname      string       `json:"nickname"`
	Mobile        string       `json:"mobile"`
	Email         string       `json:"email"`
	Enabled       bool         `json:"enabled"`
	UserDetail    interface{}  `json:"userDetail"`
	Locked        bool         `json:"locked"`
	PasswordExist bool         `json:"passwordExist"`
	PwdExpireAt   *TimestampMs `json:"pwdExpireAt"`
	Extra         interface{}  `json:"extra"`
	Source        string       `json:"source"`
	SourceType    string       `json:"sourceType"`
	Tag           string       `json:"tag"`
	Channel       string       `json:"channel"`
	ChannelType   string       `json:"channelType"`
	TenantId      int          `json:"tenantId"`
	CreatedAt     TimestampMs  `json:"createdAt"`
	UpdatedAt     *TimestampMs `json:"updatedAt"`
	LastLoginAt   *TimestampMs `json:"lastLoginAt"`
}
