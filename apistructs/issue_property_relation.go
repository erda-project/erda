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

import "strconv"

// 自定义字段详情
type IssuePropertyInstance struct {
	IssuePropertyIndex
	ArbitraryValue   interface{}         `json:"arbitraryValue"` // 自定义的值
	EnumeratedValues []PropertyEnumerate `json:"enumeratedValues"`
	Values           []int64             `json:"values"`
}

func (i IssuePropertyInstance) GetArb() string {
	switch i.ArbitraryValue.(type) {
	case float64:
		return strconv.FormatInt(int64(i.ArbitraryValue.(float64)), 10)
	case string:
		return i.ArbitraryValue.(string)
	default:
		return ""
	}
}

type PropertyEnumerate struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type IssueAndPropertyAndValue struct {
	IssueID  int64                        `json:"issueID"`
	Property []IssuePropertyExtraProperty `json:"property"`
}

// 用于自定义字段实例操作
type IssuePropertyExtraProperty struct {
	PropertyID       int64        `json:"propertyID"`
	PropertyType     PropertyType `json:"propertyType"`
	PropertyName     string       `json:"propertyName"`
	Required         bool         `json:"required"`
	DisplayName      string       `json:"displayName"`
	ArbitraryValue   interface{}  `json:"arbitraryValue"`
	EnumeratedValues []Enumerate  `json:"enumeratedValues"`
	Values           []int64      `json:"values"`
}

// 创建事件的自定义字段请求
type IssuePropertyRelationCreateRequest struct {
	OrgID     int64                   `json:"orgID"`
	ProjectID int64                   `json:"projectID"`
	IssueID   int64                   `json:"issueID"`  // 事件ID
	Property  []IssuePropertyInstance `json:"property"` // 自定义字段
	IdentityInfo
}

// 更新事件的自定义字段请求
type IssuePropertyRelationUpdateRequest struct {
	OrgID     int64 `json:"orgID"`
	ProjectID int64 `json:"projectID"`
	IssueID   int64 `json:"issueID"`
	IssuePropertyExtraProperty
	IdentityInfo
}

func (i IssuePropertyRelationUpdateRequest) GetArb() string {
	switch i.ArbitraryValue.(type) {
	case float64:
		return strconv.FormatInt(int64(i.ArbitraryValue.(float64)), 10)
	case string:
		return i.ArbitraryValue.(string)
	default:
		return ""
	}
}

// 查询事件的自定义字段请求
type IssuePropertyRelationGetRequest struct {
	OrgID             int64             `json:"orgID"`
	IssueID           int64             `json:"issueID"`           // 事件ID
	PropertyIssueType PropertyIssueType `json:"propertyIssueType"` // 任务类型
	IdentityInfo
}

// 查询事件的自定义字段响应
type IssuePropertyRelationGetResponse struct {
	Header
	IssueAndPropertyAndValue
}
