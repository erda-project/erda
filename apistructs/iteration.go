package apistructs

import "time"

// IterationCreateRequest 创建迭代请求
type IterationCreateRequest struct {
	// +optional
	StartedAt *time.Time `json:"startedAt"`
	// +optional
	FinishedAt *time.Time `json:"finishedAt"`
	// +required
	ProjectID uint64 `json:"projectID"`
	// +required
	Title string `json:"title"`
	// +optional
	Content string `json:"content"`

	// internal use, get from *http.Request
	IdentityInfo
}

// IterationCreateResponse 创建迭代响应
type IterationCreateResponse struct {
	Header
	Data *Iteration `json:"data"`
}

// IterationState 迭代归档状态
type IterationState string

const (
	IterationStateFiled   IterationState = "FILED"
	IterationStateUnfiled IterationState = "UNFILED"
)

// IterationUpdateRequest 更新迭代请求
type IterationUpdateRequest struct {
	// +required
	Title string `json:"title"`
	// +required
	Content string `json:"content"`
	// +required
	StartedAt *time.Time `json:"startedAt"`
	// +required
	FinishedAt *time.Time `json:"finishedAt"`
	// +required
	State IterationState `json:"state"`
	// internal use, get from *http.Request
	IdentityInfo
}

// IterationUpdateResponse 更新迭代响应
type IterationUpdateResponse struct {
	Header
	Data uint64 `json:"data"`
}

type IterationPagingRequest struct {
	// +optional default 1
	PageNo uint64
	// +optional default 10
	PageSize uint64
	// +optional 根据迭代结束时间过滤
	Deadline string `schema:"deadline"`
	// +required
	ProjectID uint64 `schema:"projectID"`
	// +optional 根据归档状态过滤
	State IterationState `schema:"state"`
	// +optional 是否查询事项概览，默认查询
	WithoutIssueSummary bool `schema:"withoutIssueSummary"`
}

type IterationPagingResponse struct {
	Header
	UserInfoHeader
	Data *IterationPagingResponseData `json:"data"`
}

type IterationPagingResponseData struct {
	Total uint64      `json:"total"`
	List  []Iteration `json:"list"`
}

// IterationGetResponse 迭代详情响应
type IterationGetResponse struct {
	Header
	UserInfoHeader
	Data Iteration `json:"data"`
}

type Iteration struct {
	ID           int64          `json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	StartedAt    *time.Time     `json:"startedAt"`
	FinishedAt   *time.Time     `json:"finishedAt"`
	ProjectID    uint64         `json:"projectID"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Creator      string         `json:"creator"`
	State        IterationState `json:"state"`
	IssueSummary ISummary       `json:"issueSummary"`
}

// ISummary 与迭代相关的事件完成状态的统计信息
type ISummary struct {
	Requirement ISummaryState `json:"requirement"`
	Task        ISummaryState `json:"task"`
	Bug         ISummaryState `json:"bug"`
}

type ISummaryState struct {
	Done   int `json:"done"`
	UnDone int `json:"undone"`
}
