// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type ResourceRequest struct {
	Clusters []*ResourceCluster `json:"clusters"`
	Filters  []*ResourceFilter  `json:"filters"`
	Groups   []string           `json:"groups"`
}

type ResourceCluster struct {
	ClusterName string   `json:"clusterName"`
	HostIPs     []string `json:"hostIPs"`
}

type ResourceFilter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type GroupHostDataResponse struct {
	Header
	Data DataGroupHostDTO `json:"data"`
}

type DataGroupHostDTO struct {
	Machines []*HostData `json:"machines"`
}

type HostData struct {
	ClusterName      string  `json:"clusterName"`
	IP               string  `json:"ip"`
	Hostname         string  `json:"hostname"`
	OS               string  `json:"os"`
	KernelVersion    string  `json:"kernelVersion"`
	Labels           string  `json:"labels"`
	Tasks            float64 `json:"tasks"`
	CPUUsage         float64 `json:"cpuUsage"`
	CPURequest       float64 `json:"cpuRequest"`
	CPULimit         float64 `json:"cpuLimit"`
	CPUOrigin        float64 `json:"cpuOrigin"`
	CPUTotal         float64 `json:"cpuTotal"`
	CPUAllocatable   float64 `json:"cpuAllocatable"`
	MemUsage         float64 `json:"memUsage"`
	MemRequest       float64 `json:"memRequest"`
	MemLimit         float64 `json:"memLimit"`
	MemOrigin        float64 `json:"memOrigin"`
	MemTotal         float64 `json:"memTotal"`
	MemAllocatable   float64 `json:"memAllocatable"`
	DiskUsage        float64 `json:"diskUsage"`
	DiskLimit        float64 `json:"diskLimit"`
	DiskTotal        float64 `json:"diskTotal"`
	Load1            float64 `json:"load1"`
	Load5            float64 `json:"load5"`
	Load15           float64 `json:"load15"`
	CPUUsagePercent  float64 `json:"cpuUsagePercent"`
	MemUsagePercent  float64 `json:"memUsagePercent"`
	DiskUsagePercent float64 `json:"diskUsagePercent"`
	LoadPercent      float64 `json:"loadPercent"`
	CPUDispPercent   float64 `json:"cpuDispPercent"`
	MemDispPercent   float64 `json:"memDispPercent"`
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
