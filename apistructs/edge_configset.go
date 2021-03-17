package apistructs

import "time"

// EdgeConfigSetInfo 边缘站点配置信息
type EdgeConfigSetInfo struct {
	ID          int64     `json:"id"`
	OrgID       int64     `json:"orgID"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	ClusterID   int64     `json:"clusterID"`
	ClusterName string    `json:"clusterName"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EdgeConfigSetCreateRequest 创建边缘站点请求
type EdgeConfigSetCreateRequest struct {
	ClusterID   int64  `json:"clusterID"`
	OrgID       int64  `json:"orgID"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// EdgeConfigSetUpdateRequest 更新边缘站点请求
type EdgeConfigSetUpdateRequest struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// EdgeConfigSetListPageRequest 分页查询请求
type EdgeConfigSetListPageRequest struct {
	OrgID     int64
	ClusterID int64
	NotPaging bool
	PageNo    int `query:"pageNo"`
	PageSize  int `query:"pageSize"`
}

// EdgeConfigSetListResponse 站点列表响应体
type EdgeConfigSetListResponse struct {
	Total int                 `json:"total"`
	List  []EdgeConfigSetInfo `json:"list"`
}
