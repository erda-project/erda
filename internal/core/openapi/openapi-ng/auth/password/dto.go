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

package password

import (
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/apistructs"
)

type LoginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *commonpb.UserInfo      `json:"user,omitempty"`
	Token *apistructs.OAuth2Token `json:"token,omitempty"`
}
