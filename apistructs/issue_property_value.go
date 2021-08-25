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
