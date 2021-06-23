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

// ProjectLabelType 标签类型
type ProjectLabelType string

const (
	LabelTypeIssue ProjectLabelType = "issue" // issue 标签类型
)

// ProjectLabel 标签
type ProjectLabel struct {
	ID        int64            `json:"id"`
	Name      string           `json:"name"`
	Type      ProjectLabelType `json:"type"`
	Color     string           `json:"color"`
	ProjectID uint64           `json:"projectID"`
	Creator   string           `json:"creator"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// ProjectLabelCreateRequest POST /api/labels 创建标签
type ProjectLabelCreateRequest struct {
	Name      string           `json:"name"`      // +required 标签名称
	Type      ProjectLabelType `json:"type"`      // +required 标签作用类型
	Color     string           `json:"color"`     // +required 标签颜色
	ProjectID uint64           `json:"projectID"` // +required 标签所属项目

	// internal use
	IdentityInfo
}

type ListByNamesAndProjectIDRequest struct {
	ProjectID uint64   `json:"projectID"`
	Name      []string `json:"name"`
}

type ListLabelByIDsRequest struct {
	IDs []uint64 `json:"ids"`
}

// ProjectLabelCreateResponse POST /api/labels 创建标签响应结构
type ProjectLabelCreateResponse struct {
	Header
	Data int64 `json:"data"`
}

// LabelUpdateRequest PUT /api/labels 更新标签信息
type ProjectLabelUpdateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`

	ID int64 `json:"-"`
	// internal use
	IdentityInfo
}

// ProjectLabelListRequest 标签列表请求
type ProjectLabelListRequest struct {
	ProjectID uint64           `schema:"projectID"`
	Key       string           `schema:"key"`  // 按标签名称模糊查询
	Type      ProjectLabelType `schema:"type"` // 标签作用类型
	PageNo    uint64           `schema:"pageNo"`
	PageSize  uint64           `schema:"pageSize"`
}

// ProjectLabelListResponse GET /api/labels 标签列表响应
type ProjectLabelListResponse struct {
	Header
	UserInfoHeader
	Data *ProjectLabelListResponseData `json:"data"`
}

// ProjectLabelListResponseData 标签列表响应数据结构
type ProjectLabelListResponseData struct {
	Total int64          `json:"total"`
	List  []ProjectLabel `json:"list"`
}

// ProjectLabelGetByIDResponseData 通过id获取标签响应
// 由于与删除label时产生审计事件所需要的返回一样，所以删除label时也用这个接收返回
type ProjectLabelGetByIDResponseData struct {
	Header
	Data ProjectLabel
}

type ProjectLabelsResponse struct {
	Header
	Data []ProjectLabel `json:"data"`
}
