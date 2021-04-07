// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

import "time"

type TestPlanMember struct {
	ID         uint64             `json:"id"`
	TestPlanID uint64             `json:"testPlanID"`
	Role       TestPlanMemberRole `json:"role"`
	UserID     string             `json:"userID"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

type TestPlanMemberRole string

var (
	TestPlanMemberRoleOwner   TestPlanMemberRole = "Owner"
	TestPlanMemberRolePartner TestPlanMemberRole = "Partner"
)

func (role TestPlanMemberRole) Valid() bool {
	switch role {
	case TestPlanMemberRoleOwner, TestPlanMemberRolePartner:
		return true
	default:
		return false
	}
}
func (role TestPlanMemberRole) Invalid() bool {
	return !role.Valid()
}
func (role TestPlanMemberRole) IsOwner() bool {
	return role == TestPlanMemberRoleOwner
}
func (role TestPlanMemberRole) IsPartner() bool {
	return role == TestPlanMemberRolePartner
}
