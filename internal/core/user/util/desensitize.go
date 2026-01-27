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
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/desensitize"
)

func Densensitize(IDs []string, b []*commonpb.UserInfo, needDesensitize bool) map[string]apistructs.UserInfo {
	users := make(map[string]apistructs.UserInfo, len(b))
	if needDesensitize {
		for i := range b {
			users[b[i].Id] = apistructs.UserInfo{
				ID:     "",
				Email:  desensitize.Email(b[i].Email),
				Phone:  desensitize.Mobile(b[i].Phone),
				Avatar: b[i].Avatar,
				Name:   desensitize.Name(b[i].Name),
				Nick:   desensitize.Name(b[i].Nick),
			}
		}
	} else {
		// Desensitize email and phone info
		for i := range b {
			users[b[i].Id] = apistructs.UserInfo{
				ID:     b[i].Id,
				Email:  desensitize.Email(b[i].Email),
				Phone:  desensitize.Mobile(b[i].Phone),
				Avatar: b[i].Avatar,
				Name:   b[i].Name,
				Nick:   b[i].Nick,
			}
		}
	}
	for _, userID := range IDs {
		_, exist := users[userID]
		if !exist {
			users[userID] = apistructs.UserInfo{
				ID:     userID,
				Email:  "",
				Phone:  "",
				Avatar: "",
				Name:   "用户已注销",
				Nick:   "用户已注销",
			}
		}
	}
	return users
}
