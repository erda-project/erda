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

// 创建字段枚举值请求
type IssuePropertyValueCreateRequest struct {
	Value      string `json:"value"`
	PropertyID int64  `json:"propertyID"`
	IdentityInfo
}

// 更新字段枚举值请求
type IssuePropertyValueUpdateRequest struct {
	PropertyID int64  `json:"propertyID"` // 字段ID
	ID         int64  `json:"id"`
	Value      string `json:"value"`
	IdentityInfo
}

// 删除字段枚举值请求
type IssuePropertyValueDeleteRequest struct {
	PropertyValueID int64 `json:"propertyID"` // 字段ID
	IdentityInfo
}

// 查询字段枚举值请求
type IssuePropertyValueGetRequest struct {
	PropertyID int64 `json:"propertyID"` // 企业
	IdentityInfo
}
