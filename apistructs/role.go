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

// /api/roles/<role>
// method: put
// 更改用户角色
type RoleChangeRequest struct {
	// 用户角色
	Role string         `json:"-" path:"role"`
	Body RoleChangeBody `json:"body"`
}

type RoleChangeBody struct {
	// 目标id, 对应的applicationId, projectId, orgId
	TargetId string `json:"targetId"`

	// 目标类型 APPLICATION,PROJECT,ORG
	TargetType string `json:"targetType"`
	UserId     string `json:"userId"`
}

// /api/roles/<role>
// method: put
type RoleChangeResponse struct {
	Header
	Data bool `json:"data"`
}
