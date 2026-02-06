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

package iam

// Response iam standard iam response dto
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Help    string `json:"help"`
}

type PagingData[T any] struct {
	Total int `json:"total"`
	Data  T   `json:"data"`
}

type UserDto struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Realname    string `json:"realname"`
	Mobile      string `json:"mobile"`
	Email       string `json:"email"`
	Status      bool   `json:"status"`
	Locked      bool   `json:"locked"`
	Avatar      string `json:"avatar"`
	Source      string `json:"source"`
	LastLoginIp string `json:"lastLoginIp"`
	// TODO: time format, now response: yyyy-MM-ddTHH:mm:ss
	CreatedAt        string `json:"createdAt"`
	LastLoginAt      string `json:"lastLoginAt"`
	PasswordExpireAt string `json:"passwordExpireAt"`
	LockExpireAt     string `json:"lockExpireAt"`
	// TODO: not used now
	// Application  struct `json:"application"`
	//InviteCode   *string `json:"inviteCode"`
	//InviteUserId *int64  `json:"inviteUserId"`
}

type UserWithToken struct {
	User         UserDto       `json:"user"`
	Expire       uint          `json:"expire"`
	NewToken     string        `json:"newToken"`
	CookieConfig *CookieConfig `json:"cookieConfig,omitempty"`
}

type CookieConfig struct {
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	//MaxAge   int    `json:"maxAge,omitempty"`
	//Name     string `json:"name,omitempty"`
	//SameSite string `json:"sameSite,omitempty"`
}

type UserCreate struct {
	UserName          string `json:"username"`
	NickName          string `json:"nickname"`
	Email             string `json:"email"`
	Mobile            string `json:"mobile,omitempty"`
	Password          string `json:"password"`
	NeedResetPassword bool   `json:"needResetPassword,omitempty"`
}

type UserEventDto struct {
	// TODO: only use user id now
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	// TODO: support more fields
}
