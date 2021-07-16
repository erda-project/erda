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

package model

// Host 主机元数据
type Host struct {
	BaseModel
	Name          string  `json:"hostname"`                                                 // 主机名
	OrgName       string  `gorm:"type:varchar(100);index:org_name"`                         // 企业名称
	Cluster       string  `json:"cluster_full_name" gorm:"type:varchar(100);index:cluster"` // 集群名字
	PrivateAddr   string  `json:"private_addr"`                                             // 内网地址
	Cpus          float64 `json:"cpus"`                                                     // 总CPU个数
	CpuUsage      float64 `json:"cpuUsage"`                                                 // CPU使用核数
	Memory        int64   `json:"memory"`                                                   // 总内存数（字节）
	MemoryUsage   int64   `json:"memoryUsage"`                                              // 内存使用（字节）
	Disk          int64   `json:"disk"`                                                     // 磁盘大小（字节）
	DiskUsage     int64   `json:"diskUsage"`                                                // 磁盘使用（字节）
	Load5         float64 `json:"load5"`                                                    // 负载值
	Labels        string  `json:"labels"`                                                   // 环境标签
	OS            string  `json:"os"`                                                       // 操作系统类型
	KernelVersion string  `json:"kernel_version"`                                           // 内核版本
	SystemTime    string  `json:"system_time"`                                              // 系统时间
	Birthday      int64   `json:"created_at"`                                               // 创建时间（operator定义）
	TimeStamp     int64   `json:"timestamp"`                                                // 消息本身的时间戳
	Deleted       bool    `json:"deleted"`                                                  // 资源是否被删除
}

// TableName 设置模型对应数据库表名称
func (Host) TableName() string {
	return "cm_hosts"
}
