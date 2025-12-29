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

import "github.com/erda-project/erda/internal/core/user/common"

// Response uc standard response dto
type Response[T any] struct {
	Success bool   `json:"success"`
	Result  T      `json:"result"`
	Error   string `json:"error"`
}

type ListLoginTypeResult struct {
	RegistryType []string `json:"registryType"`
}

type CurrentUser struct {
	ID          common.USERID `json:"id"`
	Email       string        `json:"email"`
	Mobile      string        `json:"mobile"`
	Username    string        `json:"username"`
	Nickname    string        `json:"nickname"`
	LastLoginAt uint64        `json:"lastLoginAt"`
}
