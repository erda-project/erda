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

package vars

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/desensitize"
	"github.com/erda-project/erda/pkg/strutil"
)

func (data *DataForFulfill) GetOrgUserNameByID(userID string) string {
	user, ok := data.OrgMemberByUserID[userID]
	if !ok {
		return userID
	}
	return user.GetUserName()
}

func (data *DataForFulfill) SetOrgAndProjectUserIDByUserKey() {
	orgMemberIDByUserKey := make(map[string]string)
	for userID, member := range data.OrgMemberByUserID {
		userID := userID
		userKeys := getMemberUserKeys(member)
		for _, userKey := range userKeys {
			orgMemberIDByUserKey[userKey] = userID
		}
	}
	data.ImportOnly.OrgMemberIDByUserKey = orgMemberIDByUserKey
	projectMemberIDByUserKey := make(map[string]string)
	for userID, member := range data.ProjectMemberByUserID {
		userID := userID
		userKeys := getMemberUserKeys(member)
		for _, userKey := range userKeys {
			projectMemberIDByUserKey[userKey] = userID
		}
	}
	data.ImportOnly.ProjectMemberIDByUserKey = projectMemberIDByUserKey
}

func getMemberUserKeys(member apistructs.Member) []string {
	return strutil.DedupSlice([]string{
		member.UserID,
		member.Mobile,
		desensitize.Mobile(member.Mobile),
		member.Email,
		desensitize.Email(member.Email),
		member.Nick,
		desensitize.Name(member.Nick),
		member.Name,
		desensitize.Name(member.Name),
	}, true)
}
