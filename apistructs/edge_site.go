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
	Search    string
	PageNo    int `query:"pageNo"`
	PageSize  int `query:"pageSize"`
}

// EdgeSiteListResponse 站点列表响应体
type EdgeSiteListResponse struct {
	Total int            `json:"total"`
	List  []EdgeSiteInfo `json:"list"`
}
