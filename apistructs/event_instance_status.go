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

// InstanceStatusData 是调度器为实例状态变化事件而定义的结构体
type InstanceStatusData struct {
	ClusterName string `json:"clusterName,omitempty"`
	RuntimeName string `json:"runtimeName,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`

	// 事件id
	// k8s 中是 containerID
	// marathon 中是 taskID
	ID string `json:"id,omitempty"`

	// 容器ip
	IP string `json:"ip,omitempty"`

	// 包含Running,Killed,Failed,Healthy,UnHealthy等状态
	InstanceStatus string `json:"instanceStatus,omitempty"`

	// 宿主机ip
	Host string `json:"host,omitempty"`

	// 事件额外描述，可能为空
	Message string `json:"message,omitempty"`

	// 时间戳到纳秒级
	Timestamp int64 `json:"timestamp"`
}
