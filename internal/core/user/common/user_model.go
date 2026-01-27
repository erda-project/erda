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

	"github.com/erda-project/erda-proto-go/core/user/pb"
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
	Token     string       `json:"token"`
	SessionID string       `json:"sessionID"`
	Cookie    *http.Cookie `json:"cookie,omitempty"`
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

const SystemOperator = "system"

var SystemUser = User{
	ID:   SystemOperator,
	Name: SystemOperator,
	Nick: SystemOperator,
}

type UserScopeInfo struct {
	OrgID uint64 `json:"orgId"`
	// dont care other fields
}

func ToPbUser(user *User) *pb.User {
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
