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

// 实例状态合集
const (
	InstanceStatusStarting  string = "Starting" // 已启动，但未收到健康检查事件，瞬态
	InstanceStatusRunning   string = "Running"
	InstanceStatusHealthy   string = "Healthy"
	InstanceStatusUnHealthy string = "UnHealthy" // 已启动但收到未通过健康检查事件

	InstanceStatusDead     string = "Dead"     // TODO Finished 等下面状态后续可以去除，暂未兼容保留
	InstanceStatusFinished string = "Finished" // 已完成，退出码为0
	InstanceStatusFailed   string = "Failed"   // 已退出，退出码非0
	InstanceStatusKilled   string = "Killed"   // 已被杀
	InstanceStatusStopped  string = "Stopped"  // 已停止，Scheduler与DCOS断连期间事件丢失，后续补偿时，需将Healthy置为Stopped
	InstanceStatusUnknown  string = "Unknown"
	InstanceStatusOOM      string = "OOM"
)
