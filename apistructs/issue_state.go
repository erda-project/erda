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

type IssueStatus struct {
	ProjectID   uint64           `json:"projectID"`
	IssueType   IssueType        `json:"issueType"`
	StateID     int64            `json:"stateID"`
	StateName   string           `json:"stateName"`
	StateBelong IssueStateBelong `json:"stateBelong"`
	Index       int64            `json:"index"`
}

// 事件主状态
type IssueStateBelong string

const (
	IssueStateBelongOpen     IssueStateBelong = "OPEN"     // 待处理
	IssueStateBelongWorking  IssueStateBelong = "WORKING"  // 进行中
	IssueStateBelongDone     IssueStateBelong = "DONE"     // 已完成
	IssueStateBelongWontfix  IssueStateBelong = "WONTFIX"  // 无需修复
	IssueStateBelongReopen   IssueStateBelong = "REOPEN"   // 重新打开
	IssueStateBelongResloved IssueStateBelong = "RESOLVED" // 已解决
	IssueStateBelongClosed   IssueStateBelong = "CLOSED"   // 已关闭
)

type IssueStateRelation struct {
	IssueStatus
	StateRelation []int64 `json:"stateRelation"`
}

type IssueTypeState struct {
	IssueType IssueType `json:"issueType"`
	State     []string  `json:"state"`
}

type IssueTypeStateID struct {
	IssueType IssueType `json:"issueType"`
	State     []string  `json:"state"`
}

type IssueStateName struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}
type IssueStateState struct {
	StateBelong IssueStateBelong `json:"stateBelong"`
	States      []IssueStateName
}
type IssueStateTypeBelong struct {
	Type   IssueType       `json:"type"`
	States IssueStateState `json:"states"`
}

type IssueStateTypeBelongResponse struct {
	Header
	Data []IssueStateState `json:"data"`
}

// 项目下工作流查询请求
type IssueStateRelationGetRequest struct {
	ProjectID uint64    `json:"projectID"`
	IssueType IssueType `json:"issueType"`
	IdentityInfo
}

// 删除状态请求
type IssueStateDeleteRequest struct {
	ProjectID int64 `json:"projectID"`
	ID        int64 `json:"id"`
	IdentityInfo
}

// 更新工作流请求
type IssueStateUpdateRequest struct {
	ProjectID int64                `json:"projectID"`
	Data      []IssueStateRelation `json:"data"`
	IdentityInfo
}

// 创建状态请求
type IssueStateCreateRequest struct {
	ProjectID   uint64           `json:"projectID"`
	IssueType   IssueType        `json:"issueType"`
	StateName   string           `json:"stateName"`
	StateBelong IssueStateBelong `json:"stateBelong"`
	IdentityInfo
}

// 获取项目下状态请求
type IssueStatesGetRequest struct {
	ProjectID uint64 `json:"projectID"`
	IdentityInfo
}

// 删除状态请求
type IssueStateDeleteResponse struct {
	Header
	Data IssueStatus `json:"data"`
}

// 按项目下任务类型分类的工作流详情
type IssueStateRelationGetResponse struct {
	Header
	Data []IssueStateRelation `json:"data"`
}

// 项目下状态列表
type IssueStatesGetResponse struct {
	Header
	Data []IssueTypeState `json:"data"`
}

// 事件主状态列表
type IssueStateTypeBelongGetResponse struct {
	Header
	Data []IssueStateState `json:"data"`
}

type IssueStateNameGetResponse struct {
	Header
	Data []IssueStatus `json:"data"`
}
