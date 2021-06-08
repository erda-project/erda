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

import (
	"time"
)

// TestPlan 测试计划
type TestPlan struct {
	ID         uint64            `json:"id"`
	Name       string            `json:"name"`
	OwnerID    string            `json:"ownerID"`
	PartnerIDs []string          `json:"partnerIDs"`
	Status     TPStatus          `json:"status"`
	ProjectID  uint64            `json:"projectID"`
	CreatorID  string            `json:"creatorID"`
	UpdaterID  string            `json:"updaterID"`
	CreatedAt  *time.Time        `json:"createdAt"`
	UpdatedAt  *time.Time        `json:"updatedAt"`
	Summary    string            `json:"summary"`
	StartedAt  *time.Time        `json:"startedAt"`
	EndedAt    *time.Time        `json:"endedAt"`
	RelsCount  TestPlanRelsCount `json:"relsCount"`
	Type       TestPlanType      `json:"type"`
	Inode      string            `json:"inode,omitempty"`
}

// TestPlanRelsCount 测试计划关联的测试用例状态个数
type TestPlanRelsCount struct {
	Total uint64 `json:"total"`
	Init  uint64 `json:"init"`
	Succ  uint64 `json:"succ"`
	Fail  uint64 `json:"fail"`
	Block uint64 `json:"block"`
}

// TPStatus 测试计划状态
type TPStatus string

const (
	TPStatusDoing TPStatus = "DOING"
	TPStatusPause TPStatus = "PAUSE"
	TPStatusDone  TPStatus = "DONE"
)

func (s TPStatus) Valid() bool {
	switch s {
	case TPStatusDoing, TPStatusPause, TPStatusDone:
		return true
	default:
		return false
	}
}

type TestPlanType string

var (
	TestPlanTypeManual   TestPlanType = "m"
	TestPlanTypeAutoTest TestPlanType = "a"
)

// TestPlanCreateRequest 测试计划创建请求
type TestPlanCreateRequest struct {
	Name       string   `json:"name"`
	OwnerID    string   `json:"ownerID"`
	PartnerIDs []string `json:"partnerIDs"`
	ProjectID  uint64   `json:"projectID"`

	// 是否是自动化测试计划
	IsAutoTest bool `json:"isAutoTest"`

	IdentityInfo
}

// TestPlanCreateResponse 测试计划创建响应
type TestPlanCreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// TestPlanUpdateRequest 测试计划更新请求
type TestPlanUpdateRequest struct {
	Name                  string         `json:"name"`
	OwnerID               string         `json:"ownerID"`
	PartnerIDs            []string       `json:"partnerIDs"`
	Status                TPStatus       `json:"status"`
	Summary               string         `json:"summary"`
	TimestampSecStartedAt *time.Duration `json:"timestampSecStartedAt"`
	TimestampSecEndedAt   *time.Duration `json:"timestampSecEndedAt"`

	TestPlanID uint64 `json:"-"`

	IdentityInfo
}

// TestPlanGetResponse 测试计划详情响应
type TestPlanGetResponse struct {
	Header
	Data TestPlan `json:"data"`
}

// TestPlanPagingRequest 测试计划列表请求
type TestPlanPagingRequest struct {
	Name      string       `schema:"name"`
	Statuses  []TPStatus   `schema:"status"`
	ProjectID uint64       `schema:"projectID"`
	Type      TestPlanType `schema:"type"`

	// member about
	OwnerIDs   []string `schema:"ownerID"`
	PartnerIDs []string `schema:"partnerID"`
	UserIDs    []string `schema:"userID"` // 只要是成员就可以，即我负责的或我参与的

	// +optional default 1
	PageNo uint64 `schema:"pageNo"`
	// +optional default 10
	PageSize uint64 `schema:"pageSize"`

	IdentityInfo
}

// TestPlanPagingResponse 测试计划响应
type TestPlanPagingResponse struct {
	Header
	UserInfoHeader
	Data TestPlanPagingResponseData `json:"data"`
}

// TestPlanPagingResponseData 测试计划响应数据
type TestPlanPagingResponseData struct {
	Total   uint64     `json:"total"`
	List    []TestPlan `json:"list"`
	UserIDs []string   `json:"userIDs,omitempty"`
}

// TestPlanTestSetsListRequest 测试计划下的测试集列表请求
type TestPlanTestSetsListRequest struct {
	TestPlanID      uint64 `schema:"-"`
	ParentTestSetID uint64 `schema:"parentTestSetID"`

	IdentityInfo
}
type TestPlanTestSetListResponse TestSetListResponse

// TestPlanCaseRelPagingRequest 测试计划内测试用例列表请求
type TestPlanCaseRelPagingRequest struct {
	PageNo   int64 `schema:"pageNo"`
	PageSize int64 `schema:"pageSize"`

	RelIDs       []uint64             `schema:"relationID"`
	TestPlanID   uint64               `schema:"-"`
	TestSetID    uint64               `schema:"testSetID"`
	Query        string               `schema:"query"`
	Priorities   []TestCasePriority   `schema:"priority"`
	UpdaterIDs   []string             `schema:"updaterID"`
	ExecutorIDs  []string             `schema:"executorID"`
	ExecStatuses []TestCaseExecStatus `schema:"execStatus"`

	// 更新时间，外部传参使用时间戳
	TimestampSecUpdatedAtBegin *time.Duration `schema:"timestampSecUpdatedAtBegin"` // 更新时间左值, 包含区间值
	TimestampSecUpdatedAtEnd   *time.Duration `schema:"timestampSecUpdatedAtEnd"`   // 更新时间右值, 包含区间值
	// 更新时间，内部使用直接赋值
	UpdatedAtBeginInclude *time.Time `schema:"-"`
	UpdatedAtEndInclude   *time.Time `schema:"-"`

	// order by field
	OrderByPriorityAsc   *bool `schema:"orderByPriorityAsc"`
	OrderByPriorityDesc  *bool `schema:"orderByPriorityDesc"`
	OrderByUpdaterIDAsc  *bool `schema:"orderByUpdaterIDAsc"`
	OrderByUpdaterIDDesc *bool `schema:"orderByUpdaterIDDesc"`
	OrderByUpdatedAtAsc  *bool `schema:"orderByUpdatedAtAsc"`
	OrderByUpdatedAtDesc *bool `schema:"orderByUpdatedAtDesc"`
	OrderByIDAsc         *bool `schema:"orderByIDAsc"`
	OrderByIDDesc        *bool `schema:"orderByIDDesc"`

	IdentityInfo
}
type TestPlanCaseRelPagingResponse struct {
	Header
	Data *TestPlanCasePagingResponseData `json:"data"`
}

// TestPlanCasePagingResponseData 测试计划内测试用例列表响应数据
type TestPlanCasePagingResponseData struct {
	Total    uint64                    `json:"total"`
	TestSets []TestSetWithPlanCaseRels `json:"testSets"`
	UserIDs  []string                  `json:"userIDs"`
}

// ApiApiTestExecuteRequest 执行接口测试计划请求
type TestPlanAPITestExecuteRequest struct {
	TestPlanID  uint64   `json:"testPlanID"`
	TestCaseIDs []uint64 `json:"testCaseIDs"`
	EnvID       uint64   `json:"envID"`

	IdentityInfo
}
type TestPlanAPITestExecuteResponse struct {
	Header
	Data uint64 `json:"data"` // triggered pipeline id
}

type AutotestExecuteTestPlansRequest struct {
	TestPlan               TestPlanV2        `json:"testPlan"`
	ClusterName            string            `json:"clusterName"`
	Labels                 map[string]string `json:"labels"`
	UserID                 string            `json:"userId"`
	ConfigManageNamespaces string            `json:"configManageNamespaces"`
	IdentityInfo           IdentityInfo      `json:"userId"`
}

type AutotestExecuteTestPlansResponse struct {
	Header
	Data *PipelineDTO `json:"data"`
}

type AutotestCancelTestPlansRequest struct {
	TestPlan TestPlanV2 `json:"testPlan"`
	UserID   string     `json:"userId"`
}

type AutotestCancelTestPlansResponse struct {
	Header
	Data string `json:"data"`
}
