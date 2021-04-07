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

// GetMonitorResponse .
type GetMonitorAlertResponse struct {
	Header
	Data *Alert `json:"data"`
}

// Alert 告警结构体
type Alert struct {
	ID           uint64                 `json:"id"`
	Name         string                 `json:"name"`
	AlertScope   string                 `json:"alertScope"`
	AlertScopeID string                 `json:"alertScopeId"`
	Enable       bool                   `json:"enable"`
	Attributes   map[string]interface{} `json:"attributes"`
	CreateTime   int64                  `json:"createTime"`
	UpdateTime   int64                  `json:"updateTime"`
}

// GetMonitorReportTaskResponse .
type GetMonitorReportTaskResponse struct {
	Header
	Data *ReportTask `json:"data"`
}

// ReportTask
type ReportTask struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	ScopeID     string `json:"scopeId"`
	DashboardId string `json:"dashboardId"`
	Enable      bool   `json:"enable"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
