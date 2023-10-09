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

package sheet_user

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type Handler struct{}

func (h *Handler) SheetName() string { return vars.NameOfSheetUser }

func (h *Handler) ImportSheet(data *vars.DataForFulfill, df excel.DecodedFile) error {
	if data.IsOldExcelFormat() {
		return nil
	}
	s, ok := df.Sheets.M[h.SheetName()]
	if !ok {
		return nil
	}
	sheet := s.UnmergedSlice
	// check title
	if len(sheet) < 1 {
		return fmt.Errorf("user sheet is empty")
	}
	var members []apistructs.Member
	for _, row := range sheet[1:] {
		var member apistructs.Member
		if err := json.Unmarshal([]byte(row[2]), &member); err != nil {
			return fmt.Errorf("failed to unmarshal user info, user id: %s, err: %v", row[0], err)
		}
		members = append(members, member)
	}
	data.ImportOnly.Sheets.Optional.UserInfo = members

	// map member for import
	if err := mapMemberForImport(data, data.ImportOnly.Sheets.Optional.UserInfo); err != nil {
		return fmt.Errorf("failed to map member, err: %v", err)
	}

	return nil
}

// createIterationsIfNotExistForImport do not create user, is too hack.
// The import operator should create user first, then import.
// We can auto add user as member into project.
func mapMemberForImport(data *vars.DataForFulfill, originalProjectMembers []apistructs.Member) error {
	var usersNeedToBeAddedAsMember []apistructs.Member

	// 先把所有原有的用户都尝试添加到 project member 中
	// 如果当前项目成员中不存在，则使用邮箱/手机进行关联查找
	// 关键是在企业下找到用户，然后添加到项目成员中，再进行用户名-ID映射 (要求用户已经在企业下，导入导出不主动添加用户到企业)
	// handle original users from user sheet
	for _, originalMember := range originalProjectMembers {
		originalMember := originalMember
		// check if already in the current project member map
		if data.IsSameErdaPlatform() { // check by user id
			findMember, ok := data.ProjectMemberByUserID[originalMember.UserID]
			if ok {
				addMemberToUserIDNickMap(data, findMember)
				continue
			}
		}
		// check by other info
		findUser, err := tryToFindUserByDesensitizedPhoneEmailNickName(data, originalMember)
		if err != nil {
			return fmt.Errorf("failed to find user, originalMember: %+v, err: %v", originalMember, err)
		}
		if findUser == nil { // no matched user, just skip
			continue
		}
		newMember := apistructs.Member{
			UserID: findUser.UserID,
			Email:  findUser.Email,
			Mobile: findUser.Mobile,
			Name:   findUser.Name,
			Nick:   findUser.Nick,
			Avatar: findUser.Avatar,
			Roles:  originalMember.Roles,
			Labels: originalMember.Labels,
		}
		addMemberToUserIDNickMap(data, newMember)
		usersNeedToBeAddedAsMember = append(usersNeedToBeAddedAsMember, newMember)
	}
	// handle nicks from issue sheet
	userNickMapFromIssueSheet := make(map[string]struct{})
	for _, model := range data.ImportOnly.Sheets.Must.IssueInfo {
		if model.Common.AssigneeName != "" {
			userNickMapFromIssueSheet[model.Common.AssigneeName] = struct{}{}
		}
		if model.Common.CreatorName != "" {
			userNickMapFromIssueSheet[model.Common.CreatorName] = struct{}{}
		}
		if model.BugOnly.OwnerName != "" {
			userNickMapFromIssueSheet[model.BugOnly.OwnerName] = struct{}{}
		}
	}
	for nick := range userNickMapFromIssueSheet {
		if nick == "" {
			continue
		}
		if _, ok := data.ImportOnly.UserIDByNick[nick]; ok {
			continue
		}
		// Risk of having the same name
		findUser, err := tryToFindUserByDesensitizedPhoneEmailNickName(data, apistructs.Member{Nick: nick})
		if err != nil {
			return fmt.Errorf("failed to find user, nick: %s, err: %v", nick, err)
		}
		if findUser == nil { // no matched user, just skip
			continue
		}
		newMember := apistructs.Member{
			UserID: findUser.UserID,
			Email:  findUser.Email,
			Mobile: findUser.Mobile,
			Name:   findUser.Name,
			Nick:   findUser.Nick,
			Avatar: findUser.Avatar,
			Roles:  []string{bundle.RoleProjectDev},
			Labels: nil,
		}
		addMemberToUserIDNickMap(data, newMember)
		usersNeedToBeAddedAsMember = append(usersNeedToBeAddedAsMember, newMember)
	}

	// add member
	for _, member := range usersNeedToBeAddedAsMember {
		// add to project
		if _, ok := data.ProjectMemberByUserID[member.UserID]; !ok {
			if err := data.Bdl.AddMember(apistructs.MemberAddRequest{
				Scope: apistructs.Scope{
					Type: apistructs.ProjectScope,
					ID:   strconv.FormatUint(data.ProjectID, 10),
				},
				Roles:   polishMemberProjectRoles(data, member.Roles),
				Labels:  member.Labels,
				UserIDs: []string{member.UserID},
			}, apistructs.SystemUserID); err != nil {
				return fmt.Errorf("failed to add member into project, project id: %d, user id: %s, err: %v", data.ProjectID, member.UserID, err)
			}
		}
	}

	// refresh member map, because bdl.AddMember won't return new member info
	orgMember, projectMember, alreadyHaveProjectOwner, err := RefreshDataMembers(data.OrgID, data.ProjectID, data.Bdl)
	if err != nil {
		return fmt.Errorf("failed to refresh data members, org id: %d, project id: %d, err: %v", data.OrgID, data.ProjectID, err)
	}
	data.OrgMemberByUserID = orgMember
	data.ProjectMemberByUserID = projectMember
	data.AlreadyHaveProjectOwner = alreadyHaveProjectOwner

	return nil
}

type Voucher struct {
	Type   string
	Value  string
	Result map[string]*userpb.User // key is user id
}

func tryToFindUserByDesensitizedPhoneEmailNickName(data *vars.DataForFulfill, member apistructs.Member) (*apistructs.Member, error) {
	// 使用 phone/email/nick/name 字段 `*` 后的数据分别匹配，全部匹配上即可
	// 为空的字段不参加匹配
	phoneMatchedUserID := data.ImportOnly.OrgMemberByDesensitizedKey[member.Mobile]
	emailMatchedUserID := data.ImportOnly.OrgMemberByDesensitizedKey[member.Email]
	nickMatchedUserID := data.ImportOnly.OrgMemberByDesensitizedKey[member.Nick]
	nameMatchedUserID := data.ImportOnly.OrgMemberByDesensitizedKey[member.Name]
	// check ids all equals
	compares := make(map[string]struct{})
	var findUserID string
	if phoneMatchedUserID != "" {
		compares[phoneMatchedUserID] = struct{}{}
		findUserID = phoneMatchedUserID
	}
	if emailMatchedUserID != "" {
		compares[emailMatchedUserID] = struct{}{}
		findUserID = emailMatchedUserID
	}
	if nickMatchedUserID != "" {
		compares[nickMatchedUserID] = struct{}{}
		findUserID = nickMatchedUserID
	}
	if nameMatchedUserID != "" {
		compares[nameMatchedUserID] = struct{}{}
		findUserID = nameMatchedUserID
	}
	if len(compares) == 1 {
		findUser := data.OrgMemberByUserID[findUserID]
		return &findUser, nil
	}
	return nil, nil
}

// getPartVoucherFromDesensitizedData
// abc -> abc
// abc*d -> d
// "" -> ""
func getPartVoucherFromDesensitizedData(voucher string) string {
	if voucher == "" {
		return ""
	}
	parts := strings.Split(voucher, "*")
	return parts[len(parts)-1]
}

func addMemberToUserIDNickMap(data *vars.DataForFulfill, member apistructs.Member) {
	if member.Email != "" {
		data.ImportOnly.UserIDByNick[member.Email] = member.UserID
	}
	if member.Mobile != "" {
		data.ImportOnly.UserIDByNick[member.Mobile] = member.UserID
	}
	if member.Nick != "" {
		data.ImportOnly.UserIDByNick[member.Nick] = member.UserID
	}
	if member.Name != "" {
		data.ImportOnly.UserIDByNick[member.Name] = member.UserID
	}
}

// polishMemberProjectRoles remove duplicated owner role
func polishMemberProjectRoles(data *vars.DataForFulfill, roles []string) []string {
	var newRoles []string
	for _, role := range roles {
		if role == bundle.RoleProjectOwner {
			if data.AlreadyHaveProjectOwner {
				newRoles = append(newRoles, bundle.RoleProjectLead) // downgrade to lead
				continue
			}
			data.AlreadyHaveProjectOwner = true
			newRoles = append(newRoles, role)
			continue
		}
		newRoles = append(newRoles, role)
	}
	return strutil.DedupSlice(newRoles, true)
}

// RefreshDataMembers return org member, project member, alreadyHaveProjectOwner and error.
func RefreshDataMembers(orgID int64, projectID uint64, bdl *bundle.Bundle) (map[string]apistructs.Member, map[string]apistructs.Member, bool, error) {
	// org
	orgMemberQuery := apistructs.MemberListRequest{
		ScopeType:         apistructs.OrgScope,
		ScopeID:           orgID,
		PageNo:            1,
		PageSize:          99999,
		DesensitizeEmail:  false,
		DesensitizeMobile: false,
	}
	orgMember, err := bdl.ListMembers(orgMemberQuery)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to list orgMember, err: %v", err)
	}
	orgMemberMap := map[string]apistructs.Member{}
	for _, member := range orgMember {
		orgMemberMap[member.UserID] = member
	}

	// project
	projectMemberQuery := apistructs.MemberListRequest{
		ScopeType:         apistructs.ProjectScope,
		ScopeID:           int64(projectID),
		PageNo:            1,
		PageSize:          99999,
		DesensitizeEmail:  false,
		DesensitizeMobile: false,
	}
	projectMember, err := bdl.ListMembers(projectMemberQuery)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to list projectMember, err: %v", err)
	}
	var alreadyHaveProjectOwner bool
	projectMemberMap := map[string]apistructs.Member{}
	for _, member := range projectMember {
		projectMemberMap[member.UserID] = member
		// check project owner
		if alreadyHaveProjectOwner {
			continue
		}
		for _, role := range member.Roles {
			if role == bundle.RoleProjectOwner {
				alreadyHaveProjectOwner = true
				break
			}
		}
	}

	return orgMemberMap, projectMemberMap, alreadyHaveProjectOwner, nil
}
