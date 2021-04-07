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
