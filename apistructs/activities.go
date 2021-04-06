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
