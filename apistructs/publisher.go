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

// PublisherCreateRequest POST /api/publishers 创建Publisher请求结构
type PublisherCreateRequest struct {
	Name          string `json:"name"`
	PublisherType string `json:"publisherType"`
	Logo          string `json:"logo"`
	Desc          string `json:"desc"`
	OrgID         uint64 `json:"orgId"`
}

// PublisherCreateResponse POST /api/publishers 创建Publisher响应结构
type PublisherCreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// PublisherUpdateRequest PUT /api/publishers 更新Publisher请求结构
type PublisherUpdateRequest struct {
	ID   uint64 `json:"id"`
	Logo string `json:"logo"`
	Desc string `json:"desc"`
}

// PublisherUpdateResponse PUT /api/publishers/{publisherId} 更新Publisher响应结构
type PublisherUpdateResponse struct {
	Header
	Data interface{} `json:"data"`
}

// PublisherDeleteResponse DELETE /api/publishers/{publisherId} 删除Publisher响应结构
type PublisherDeleteResponse struct {
	Header
	Data uint64 `json:"data"`
}

//PublisherDetailResponse GET /api/publishers/{publisherId} Publisher详情响应结构
type PublisherDetailResponse struct {
	Header
	Data PublisherDTO `json:"data"`
}

// PublisherListRequest GET /api/publishers 获取Publisher列表请求
type PublisherListRequest struct {
	OrgID uint64 `query:"orgId"`

	// 是否只展示已加入的 Publisher
	Joined bool `query:"joined"`

	// 对Publisher名进行like查询
	Query    string `query:"q"`
	Name     string `query:"name"`
	PageNo   int    `query:"pageNo"`
	PageSize int    `query:"pageSize"`
}

// PublisherListResponse GET /api/publishers 查询Publisher响应
type PublisherListResponse struct {
	Header
	Data PagingPublisherDTO `json:"data"`
}

// PagingPublisherDTO 查询Publisher响应Body
type PagingPublisherDTO struct {
	Total int            `json:"total"`
	List  []PublisherDTO `json:"list"`
}

//PublisherDTO Publisher结构
type PublisherDTO struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	PublisherType string    `json:"publishType"`
	PublisherKey  string    `json:"publishKey"`
	OrgID         uint64    `json:"orgId"`
	Creator       string    `json:"creator"`
	Logo          string    `json:"logo"`
	Desc          string    `json:"desc"`
	Joined        bool      `json:"joined"`    // 用户是否已加入Publisher
	CreatedAt     time.Time `json:"createdAt"` // Publisher创建时间
	UpdatedAt     time.Time `json:"updatedAt"` // Publisher更新时间

	NexusRepositories    []*NexusRepository `json:"nexusRepositories"`
	PipelineCmNamespaces []string           `json:"pipelineCmNamespaces"` // 同步 nexus 配置至 pipeline cm
}

// CreateOrgPublisherRequest POST
type CreateOrgPublisherRequest struct {
	Name string `json:"name"`
}
