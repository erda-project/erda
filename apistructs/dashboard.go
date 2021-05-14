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

type DashboardCreateRequest struct {
	// 布局相关json
	Layout string `json:"layout"`

	// 绘制相关json
	DrawerInfoMap string `json:"drawerInfoMap"`
}

type DashboardCreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

type DashboardDetailRequest struct {
	// 配置id
	Id uint64 `path:"id"`
}

type DashboardDetailResponse struct {
	Header
	Data DashBoardDTO `json:"data"`
}

type DashboardListResponse struct {
	Header
	Data []DashBoardDTO `json:"data"`
}

type DashBoardDTO struct {
	// 记录主键id
	Id uint64 `json:"id"`

	// 唯一标识
	UniqueId string `json:"uniqueId"`

	// 绘制信息
	DrawerInfoMap string `json:"drawerInfoMap"`

	// 布局信息
	Layout string `json:"layout"`
}

type DashboardSpotLogLine struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Stream     string `json:"stream"`
	TimeBucket string `json:"timeBucket"`
	TimeStamp  string `json:"timestamp"`
	Content    string `json:"content"`
	Offset     string `json:"offset"`
	Level      string `json:"level"`
	RequestID  string `json:"requestId"`
}

type DashboardSpotLogData struct {
	Lines []DashboardSpotLogLine `json:"lines"`
}

type DashboardSpotLogResponse struct {
	Header
	Data DashboardSpotLogData `json:"data"`
}

type DashboardSpotLogRequest struct {
	ID     string
	Source DashboardSpotLogSource
	Stream DashboardSpotLogStream
	Count  int64
	Start  time.Duration // 纳秒
	End    time.Duration // 纳秒
}

type DashboardSpotLogStream string

var (
	DashboardSpotLogStreamStdout DashboardSpotLogStream = "stdout"
	DashboardSpotLogStreamStderr DashboardSpotLogStream = "stderr"
)

type DashboardSpotLogSource string

var (
	DashboardSpotLogSourceJob       DashboardSpotLogSource = "job"
	DashboardSpotLogSourceContainer DashboardSpotLogSource = "container"
	DashboardSpotLogSourceDeploy    DashboardSpotLogSource = "deploy"
)
