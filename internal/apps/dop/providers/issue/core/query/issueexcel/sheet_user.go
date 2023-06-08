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

package issueexcel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/legacy/services/permission"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (data DataForFulfill) genUserSheet() (excel.Rows, error) {
	var lines excel.Rows
	// title: user id, user name, user info (JSON)
	title := excel.Row{
		excel.NewTitleCell("user id"),
		excel.NewTitleCell("user name"),
		excel.NewTitleCell("user detail (json)"),
	}
	lines = append(lines, title)
	// data
	for _, user := range data.ProjectMemberByUserID {
		userInfo, err := json.Marshal(user)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user info, user id: %s, err: %v", user.UserID, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(user.UserID),
			excel.NewCell(user.Nick),
			excel.NewCell(string(userInfo)),
		})
	}

	return lines, nil
}

func (data DataForFulfill) decodeUserSheet(excelSheets [][][]string) ([]apistructs.Member, error) {
	if data.IsOldExcelFormat() {
		return nil, nil
	}
	sheet := excelSheets[indexOfSheetUser]
	// check title
	if len(sheet) < 1 {
		return nil, fmt.Errorf("user sheet is empty")
	}
	var members []apistructs.Member
	for _, row := range sheet[1:] {
		var member apistructs.Member
		if err := json.Unmarshal([]byte(row[2]), &member); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info, user id: %s, err: %v", row[0], err)
		}
		members = append(members, member)
	}
	return members, nil
}

// createIterationsIfNotExistForImport do not create user, is too hack.
// The import operator should create user first, then import.
// We can auto add user as member into project.
func (data *DataForFulfill) mapMemberForImport(originalMembers []apistructs.Member, issueSheetModels []IssueSheetModel) error {
	data.ImportOnly.UserIDsByNick = make(map[string]string)

	// handle nicks from issue sheet
	var userNicksFromIssueSheet []string
	for _, model := range issueSheetModels {
		userNicksFromIssueSheet = append(userNicksFromIssueSheet, model.Common.CreatorName, model.Common.AssigneeName, model.BugOnly.OwnerName)
	}
	userNicksFromIssueSheet = strutil.DedupSlice(userNicksFromIssueSheet, true)
	// try to find user in current project members, only can be found by nick (have the risk of same name)
	projectMemberByNick := make(map[string]apistructs.Member)
	for _, member := range data.ProjectMemberByUserID {
		projectMemberByNick[member.Nick] = member
	}
	orgMemberByNick := make(map[string]apistructs.Member)
	for _, member := range data.OrgMemberByUserID {
		orgMemberByNick[member.Nick] = member
	}
	for _, nick := range userNicksFromIssueSheet {
		m, ok := projectMemberByNick[nick]
		if !ok {
			m, ok = orgMemberByNick[nick] // only try to find in org member, due to the risk of same name
			if !ok {
				return fmt.Errorf("failed to find user in project member, nick: %s, please add user to project first", nick)
			}
			// auto add user to project
			if err := data.ImportOnly.Bdl.AddMember(apistructs.MemberAddRequest{
				Scope: apistructs.Scope{
					Type: apistructs.ProjectScope,
					ID:   strconv.FormatUint(data.ProjectID, 10),
				},
				Roles:   []string{"Dev"},
				UserIDs: []string{m.UserID},
			}, apistructs.SystemUserID); err != nil {
				return fmt.Errorf("failed to add member into project by nick, project id: %d, user nick: %s, user id: %s, err: %v", data.ProjectID, nick, m.UserID, err)
			}
		}
		data.ImportOnly.UserIDsByNick[nick] = m.UserID
	}

	// handle original users from user sheet
	for _, originalMember := range originalMembers {
		originalMember := originalMember
		// check if already in the current project member map
		projectMember, ok := data.tryToFindMemberInCurrentProject(originalMember)
		if ok {
			data.ImportOnly.UserIDsByNick[originalMember.Nick] = projectMember.UserID
			continue
		}
		// not found in project member
		userID := originalMember.UserID
		// 如果在同一个平台，用户肯定存在，直接进行 member 添加
		// 只有不在一个平台时，才进行用户查找
		// 如果是内部保留用户，也不需要查找
		if !data.IsSameErdaPlatform() && !permission.IsReservedInternalUserID(userID) {
			// check user exist or not, by phone/email
			userFind, err := data.tryToFindUser(originalMember)
			if err != nil {
				return fmt.Errorf("failed to find user, originalMember: %+v, err: %v", originalMember, err)
			}
			// if user not exist, just throw error, and the import operator should let the user register first.
			if userFind == nil {
				return fmt.Errorf("user not exist, should register by email/phone first. "+
					"originalMember name: %s, nick: %s, email: %s, phone: %s",
					originalMember.Name, originalMember.Nick, originalMember.Email, originalMember.Mobile)
			}
			userID = userFind.ID
		}
		data.ImportOnly.UserIDsByNick[originalMember.Nick] = userID
		// if user exist, just add originalMember into project
		// add to org first
		if _, ok := data.OrgMemberByUserID[userID]; !ok {
			if err := data.ImportOnly.Bdl.AddMember(apistructs.MemberAddRequest{
				Scope: apistructs.Scope{
					Type: apistructs.OrgScope,
					ID:   strconv.FormatInt(data.OrgID, 10),
				},
				Roles:   []string{"Dev"},
				Labels:  nil,
				UserIDs: []string{userID},
			}, apistructs.SystemUserID); err != nil {
				return fmt.Errorf("failed to add member into org, org id: %d, user id: %s, err: %v", data.OrgID, userID, err)
			}
		}
		// add to project
		if _, ok := data.ProjectMemberByUserID[userID]; !ok {
			if err := data.ImportOnly.Bdl.AddMember(apistructs.MemberAddRequest{
				Scope: apistructs.Scope{
					Type: apistructs.ProjectScope,
					ID:   strconv.FormatUint(data.ProjectID, 10),
				},
				Roles:   originalMember.Roles,
				Labels:  originalMember.Labels,
				UserIDs: []string{userID},
			}, apistructs.SystemUserID); err != nil {
				return fmt.Errorf("failed to add member into project, project id: %d, user id: %s, err: %v", data.ProjectID, userID, err)
			}
		}
	}
	return nil
}

func (data DataForFulfill) tryToFindMemberInCurrentProject(originalMember apistructs.Member) (*apistructs.Member, bool) {
	if data.IsSameErdaPlatform() {
		// just find by id
		member, ok := data.ProjectMemberByUserID[originalMember.UserID]
		if !ok {
			return nil, false
		}
		return &member, ok
	}
	for _, member := range data.ProjectMemberByUserID {
		if member.Mobile == originalMember.Mobile || member.Email == originalMember.Email {
			return &member, true
		}
	}
	return nil, false
}

func (data DataForFulfill) tryToFindUser(member apistructs.Member) (*userpb.User, error) {
	// find user by phone/email
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")
	if member.Mobile != "" {
		resp, err := data.ImportOnly.Identity.FindUsersByKey(ctx, &userpb.FindUsersByKeyRequest{
			Key: member.Mobile,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find user by mobile, mobile: %s, err: %v", member.Mobile, err)
		}
		if len(resp.Data) > 0 {
			return resp.Data[0], nil
		}
	}
	if member.Email != "" {
		resp, err := data.ImportOnly.Identity.FindUsersByKey(ctx, &userpb.FindUsersByKeyRequest{
			Key: member.Email,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find user by email, email: %s, err: %v", member.Email, err)
		}
		if len(resp.Data) > 0 {
			return resp.Data[0], nil
		}
	}
	return nil, nil
}
