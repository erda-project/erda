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

// NotifyHistory
type NotifyHistory struct {
	ID         int64  `json:"id"`
	NotifyName string `json:"notifyName"`
	// todo json key名需要cdp前端配合修改后再改
	NotifyItemDisplayName string         `json:"notifyItemName"`
	Channel               string         `json:"channel"`
	NotifyTargets         []NotifyTarget `json:"notifyTargets"`
	NotifySource          NotifySource   `json:"notifySource"`
	Status                string         `json:"status"`
	ErrorMsg              string         `json:"errorMsg"`
	Label                 string         `json:"label"`
	CreatedAt             time.Time      `json:"createdAt"`
}

// QueryNotifyHistoryRequest 查询通知发送记录请求
type QueryNotifyHistoryRequest struct {
	PageNo      int64  `query:"pageNo"`
	PageSize    int64  `query:"pageSize"`
	NotifyName  string `query:"notifyName"`
	StartTime   string `query:"startTime"`
	EndTime     string `query:"endTime"`
	Channel     string `query:"channel"`
	Label       string `query:"label"`
	ClusterName string `query:"clusterName"`
	OrgID       int64  `json:"-"`
}

// QueryNotifyHistoryResponse 查询通知历史纪录响应
type QueryNotifyHistoryResponse struct {
	Header
	Data QueryNotifyHistoryData `json:"data"`
}

// QueryNotifyHistoryData 通知发送记录结构
type QueryNotifyHistoryData struct {
	List  []*NotifyHistory `json:"list"`
	Total int              `json:"total"`
}

// CreateNotifyHistoryRequest 创建通知发送记录请求
type CreateNotifyHistoryRequest struct {
	NotifyName            string         `json:"notifyName"`
	NotifyItemDisplayName string         `json:"notifyItemDisplayName"`
	Channel               string         `json:"channel"`
	NotifyTargets         []NotifyTarget `json:"notifyTargets"`
	NotifySource          NotifySource   `json:"notifySource"`
	Status                string         `json:"status"`
	ErrorMsg              string         `json:"errorMsg"`
	OrgID                 int64          `json:"orgId"`
	Label                 string         `json:"label"`
	ClusterName           string         `query:"clusterName"`
}

// CreateNotifyHistoryResponse 创建通知发送记录响应
type CreateNotifyHistoryResponse struct {
	Header
	Data int64
}
