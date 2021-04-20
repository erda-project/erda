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

// Package event manipulates the k8s api of event object
package event

var (
	errPullImage               = "拉取镜像失败"
	errImageNeverPull          = "服务配置了永不拉取镜像，而调度的宿主机又没有对应镜像"
	errInvalidImageName        = "无效的镜像名"
	errNetworkNotReady         = "节点网络组件异常"
	errHostPortConflict        = "宿主机端口冲突"
	errNodeSelectorMismatching = "节点标签不匹配"
	errInsufficientFreeCPU     = "CPU 资源不足"
	errInsufficientFreeMemory  = "内存资源不足"
	errMountVolume             = "存储卷挂载失败"
	errAlreadyMountedVolume    = "存储卷已经被挂载过"
	errNodeRebooted            = "节点重启"
)
