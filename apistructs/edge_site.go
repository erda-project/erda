package apistructs

import "time"

type EdgeSiteInfo struct {
	ID          int64     `json:"id"`
	OrgID       int64     `json:"orgID"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	ClusterID   int64     `json:"clusterID"`
	ClusterName string    `json:"clusterName"`
	Logo        string    `json:"logo"`
	Description string    `json:"description"`
	NodeCount   string    `json:"nodeCount"`
	Status      int64     `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EdgeSiteCreateRequest 创建边缘站点请求
type EdgeSiteCreateRequest struct {
	OrgID       int64  `json:"orgID"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ClusterID   int64  `json:"clusterID"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Status      int64  `json:"status"`
}

// EdgeSiteUpdateRequest 更新边缘站点请求
type EdgeSiteUpdateRequest struct {
	DisplayName string `json:"displayName"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Status      int64  `json:"status"`
}

// EdgeSiteListPageRequest 分页查询请求, NotPaging 参数默认为 false，开启分页
type EdgeSiteListPageRequest struct {
	OrgID     int64
	ClusterID int64
	NotPaging bool
	PageNo    int `query:"pageNo"`
	PageSize  int `query:"pageSize"`
}

// EdgeSiteListResponse 站点列表响应体
type EdgeSiteListResponse struct {
	Total int            `json:"total"`
	List  []EdgeSiteInfo `json:"list"`
}
