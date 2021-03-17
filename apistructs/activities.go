package apistructs

import (
	"time"
)

// ActivitiyListRequest GET /api/activities 活动查询请求结构
type ActivitiyListRequest struct {
	OrgID         int64  `query:"orgId"`
	ProjectID     int64  `query:"projectId"`
	ApplicationID int64  `query:"applicationId"`
	RuntimeID     int64  `query:"runtimeId"`
	UserID        string `query:"userId"`

	// default 1
	PageNo int `query:"pageNo"`

	// default 20
	PageSize int `query:"pageSize"`
}

// ActivityListResponse GET api/activities 活动查询响应结构
type ActivityListResponse struct {
	Header
	Data ActivityListResponseData `json:"data"`
}

// ActivityListResponse 活动列表返回结构
type ActivityListResponseData struct {
	Total int           `json:"total"`
	List  []ActivityDTO `json:"list"`
}

// ActivityDTO 活动结构
type ActivityDTO struct {
	ID            int64       `json:"id"`
	OrgID         int64       `json:"orgId"`
	ProjectID     int64       `json:"projectId"`
	ApplicationID int64       `json:"applicationId"`
	RuntimeID     int64       `json:"runtimeId"`
	UserID        string      `json:"userId"`
	Type          string      `json:"type"`
	Action        string      `json:"action"`
	Desc          string      `json:"desc"`
	Context       interface{} `json:"context"`
	CreatedAt     time.Time   `json:"createdAt"`
}
