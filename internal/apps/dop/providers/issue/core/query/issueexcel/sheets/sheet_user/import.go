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

	"github.com/mohae/deepcopy"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type Handler struct{ sheets.DefaultImporter }

func (h *Handler) SheetName() string { return vars.NameOfSheetUser }

func (h *Handler) DecodeSheet(data *vars.DataForFulfill, s *excel.Sheet) error {
	if data.IsOldExcelFormat() {
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

	return nil
}

func (h *Handler) BeforeCreateIssues(data *vars.DataForFulfill) error {
	if err := addMemberIntoProject(data, data.ImportOnly.Sheets.Optional.UserInfo); err != nil {
		return fmt.Errorf("failed to add member into project, err: %v", err)
	}
	return nil
}

// addMemberIntoProject do not create user, is too hack.
func addMemberIntoProject(data *vars.DataForFulfill, projectMembersFromUserSheet []apistructs.Member) error {
	var usersNeedToBeAddedAsProjectMember []apistructs.Member

	// handle original users from user-sheet
	// 先把所有原有的用户都尝试添加到 project member 中
	// 如果当前项目成员中不存在，则使用邮箱/手机进行关联查找
	// 关键是在企业下找到用户，然后添加到项目成员中，再进行用户名-ID映射 (要求用户已经在企业下，导入导出不主动添加用户到企业)
	for _, originalProjectMember := range projectMembersFromUserSheet {
		originalProjectMember := originalProjectMember
		// check if already in the current project member map
		if data.IsSameErdaPlatform() { // check by user id
			_, ok := data.ProjectMemberByUserID[originalProjectMember.UserID]
			if ok {
				continue
			}
		}
		// check by other info
		findUser, err := tryToFindUserByUserKeys(data, originalProjectMember)
		if err != nil {
			return fmt.Errorf("failed to find user, originalProjectMember: %+v, err: %v", originalProjectMember, err)
		}
		if findUser == nil { // no matched user, just skip
			continue
		}
		newMember := deepcopy.Copy(*findUser).(apistructs.Member)
		newMember.Roles = originalProjectMember.Roles
		newMember.Labels = originalProjectMember.Labels
		usersNeedToBeAddedAsProjectMember = append(usersNeedToBeAddedAsProjectMember, newMember)
	}

	// handle users from issue-sheet
	userKeyMapFromIssueSheet := make(map[string]struct{})
	for _, model := range data.ImportOnly.Sheets.Must.IssueInfo {
		if model.Common.AssigneeName != "" {
			userKeyMapFromIssueSheet[model.Common.AssigneeName] = struct{}{}
		}
		if model.Common.CreatorName != "" {
			userKeyMapFromIssueSheet[model.Common.CreatorName] = struct{}{}
		}
		if model.BugOnly.OwnerName != "" {
			userKeyMapFromIssueSheet[model.BugOnly.OwnerName] = struct{}{}
		}
		// custom fields
		for _, userKey := range collectUserKeysFromCustomFields(data, model) {
			userKeyMapFromIssueSheet[userKey] = struct{}{}
		}
	}
	for userKey := range userKeyMapFromIssueSheet {
		if userKey == "" {
			continue
		}
		if _, ok := data.ImportOnly.OrgMemberIDByUserKey[userKey]; !ok { // ignore user not in org
			continue
		}
		findUser, err := tryToFindUserByUserKeys(data, apistructs.Member{Nick: userKey})
		if err != nil {
			return fmt.Errorf("failed to find user, userKey: %s, err: %v", userKey, err)
		}
		if findUser == nil { // no matched user, just skip
			continue
		}
		newMember := deepcopy.Copy(*findUser).(apistructs.Member)
		newMember.Roles = []string{bundle.RoleProjectDev}
		usersNeedToBeAddedAsProjectMember = append(usersNeedToBeAddedAsProjectMember, newMember)
	}

	// add project member
	for _, member := range usersNeedToBeAddedAsProjectMember {
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
	if err := RefreshDataMembers(data); err != nil {
		return fmt.Errorf("failed to refresh data members, org id: %d, project id: %d, err: %v", data.OrgID, data.ProjectID, err)
	}

	return nil
}

func collectUserKeysFromCustomFields(data *vars.DataForFulfill, model vars.IssueSheetModel) []string {
	var customFields []vars.ExcelCustomField
	var cfType pb.PropertyIssueTypeEnum_PropertyIssueType
	switch model.Common.IssueType {
	case pb.IssueTypeEnum_REQUIREMENT:
		customFields = model.RequirementOnly.CustomFields
		cfType = pb.PropertyIssueTypeEnum_REQUIREMENT
	case pb.IssueTypeEnum_TASK:
		customFields = model.TaskOnly.CustomFields
		cfType = pb.PropertyIssueTypeEnum_TASK
	case pb.IssueTypeEnum_BUG:
		customFields = model.BugOnly.CustomFields
		cfType = pb.PropertyIssueTypeEnum_BUG
	default:
		return nil
	}
	var userKeys []string
	for _, excelCf := range customFields {
		if excelCf.Value == "" {
			continue
		}
		cf := data.CustomFieldMapByTypeName[cfType][excelCf.Title]
		if cf == nil {
			continue
		}
		if cf.PropertyType == pb.PropertyTypeEnum_Person {
			userKeys = append(userKeys, excelCf.Value)
		}
	}
	return strutil.DedupSlice(userKeys, true)
}

type Voucher struct {
	Type   string
	Value  string
	Result map[string]*commonpb.UserInfo // key is user id
}

func tryToFindUserByUserKeys(data *vars.DataForFulfill, member apistructs.Member) (*apistructs.Member, error) {
	// 使用 phone/email/nick/name 字段 `*` 后的数据分别匹配，全部匹配上即可
	// 为空的字段不参加匹配
	phoneMatchedUserID := data.ImportOnly.OrgMemberIDByUserKey[member.Mobile]
	emailMatchedUserID := data.ImportOnly.OrgMemberIDByUserKey[member.Email]
	nickMatchedUserID := data.ImportOnly.OrgMemberIDByUserKey[member.Nick]
	nameMatchedUserID := data.ImportOnly.OrgMemberIDByUserKey[member.Name]
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
func RefreshDataMembers(data *vars.DataForFulfill) error {
	// org
	orgMemberQuery := apistructs.MemberListRequest{
		ScopeType:         apistructs.OrgScope,
		ScopeID:           data.OrgID,
		PageNo:            1,
		PageSize:          99999,
		DesensitizeEmail:  false,
		DesensitizeMobile: false,
	}
	orgMember, err := data.Bdl.ListMembers(orgMemberQuery)
	if err != nil {
		return fmt.Errorf("failed to list orgMember, err: %v", err)
	}
	orgMemberMap := map[string]apistructs.Member{}
	for _, member := range orgMember {
		orgMemberMap[member.UserID] = member
	}

	// project
	projectMemberQuery := apistructs.MemberListRequest{
		ScopeType:         apistructs.ProjectScope,
		ScopeID:           int64(data.ProjectID),
		PageNo:            1,
		PageSize:          99999,
		DesensitizeEmail:  false,
		DesensitizeMobile: false,
	}
	projectMember, err := data.Bdl.ListMembers(projectMemberQuery)
	if err != nil {
		return fmt.Errorf("failed to list projectMember, err: %v", err)
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

	data.OrgMemberByUserID = orgMemberMap
	data.ProjectMemberByUserID = projectMemberMap
	data.AlreadyHaveProjectOwner = alreadyHaveProjectOwner
	data.SetOrgAndProjectUserIDByUserKey()

	return nil
}
