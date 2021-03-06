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

// HostFetchRequest 主机详情请求
type HostFetchRequest struct {
	ClusterName string `query:"clusterName"`
}

// HostFetchResponse 主机详情响应
type HostFetchResponse struct {
	Header
	Data Host `json:"data"`
}

// HostListRequest 主机列表请求
type HostListRequest struct {
	ClusterName string `query:"clusterName"`
}

// HostListResponse 主机列表响应
type HostListResponse struct {
	Header
	Data []Host `json:"data"`
}

// Host 主机元数据
type Host struct {
	Name          string  `json:"hostname"`                               // 主机名
	OrgName       string  `json:"orgName"`                                // 企业名称
	Cluster       string  `json:"cluster_full_name" gorm:"index:cluster"` // 集群名字
	Cpus          float64 `json:"cpus"`                                   // 总CPU个数
	CpuUsage      float64 `json:"cpuUsage"`                               // CPU使用核数
	Memory        int64   `json:"memory"`                                 // 总内存数（字节）
	MemoryUsage   int64   `json:"memoryUsage"`                            // 内存使用
	Disk          int64   `json:"disk"`                                   // 磁盘大小（字节）
	DiskUsage     int64   `json:"diskUsage"`                              // 磁盘使用大小（字节）
	Load5         float64 `json:"load5"`                                  // 负载值
	PrivateAddr   string  `json:"private_addr"`                           // 内网地址
	Labels        string  `json:"labels"`                                 // 环境标签
	OS            string  `json:"os"`                                     // 操作系统类型
	KernelVersion string  `json:"kernel_version"`                         // 内核版本
	SystemTime    string  `json:"system_time"`                            // 系统时间
	Birthday      int64   `json:"created_at"`                             // 创建时间（operator定义）
	Deleted       bool    `json:"deleted"`                                // 资源是否被删除
	TimeStamp     int64   `json:"timestamp"`                              // 消息本身的时间戳
}
