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

package common

import (
	"errors"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

const (
	Default = "Default"

	NodeStatusReady int = iota
	NodeStatusError
	NodeStatusFreeze
)

type SteveStatus struct {
	Value      string `json:"value,omitempty"`
	RenderType string `json:"renderType"`
	Status     string `json:"status"`
}

var (
	/*
		node phase
	*/
	//NodeSuccess SteveStatusEnum = "success"
	//NodeDefault SteveStatusEnum = "default"
	//NodeFreeze  SteveStatusEnum = "freeze"
	//NodeError   SteveStatusEnum = "error"
	//
	//NodeSuccessCN SteveStatusEnum = "正常"
	//NodeDefaultCN SteveStatusEnum = "默认"
	//NodeFreezeCN  SteveStatusEnum = "冻结"
	//NodeErrorCN   SteveStatusEnum = "节点错误"
	//
	//// NodeReady means kubelet is healthy and ready to accept pods.
	//NodeReady SteveStatusEnum = "Ready"
	//// NodeMemoryPressure means the kubelet is under pressure due to insufficient available memory.
	//NodeMemoryPressure SteveStatusEnum = "MemoryPressure"
	//// NodeDiskPressure means the kubelet is under pressure due to insufficient available disk.
	//NodeDiskPressure SteveStatusEnum = "DiskPressure"
	//// NodePIDPressure means the kubelet is under pressure due to insufficient available PID.
	//NodePIDPressure SteveStatusEnum = "PIDPressure"
	//// NodeNetworkUnavailable means that network for the node is not correctly configured.
	//NodeNetworkUnavailable SteveStatusEnum = "NetworkUnavailable"
	//
	///*
	//	pod status
	//*/
	//PodRunning   = SteveStatusEnum(v1.PodRunning)
	//PodPending   = SteveStatusEnum(v1.PodPending)
	//PodSuccessed = SteveStatusEnum(v1.PodSucceeded)
	//PodFailed    = SteveStatusEnum(v1.PodFailed)
	//PodUnknown   = SteveStatusEnum(v1.PodUnknown)
	//
	//PodRunningCN   SteveStatusEnum = "运行"
	//PodPendingCN   SteveStatusEnum = "预备"
	//PodSuccessedCN SteveStatusEnum = "退出成功"
	//PodFailedCN    SteveStatusEnum = "退出错误"
	//PodUnknownCN   SteveStatusEnum = "未知"
	//
	///*
	//	resource usage status
	//*/
	//ResourceDefault UsageStatusEnum = "default"
	//ResourceSafe    UsageStatusEnum = "safe"
	//ResourceWarning UsageStatusEnum = "warning"
	//ResourceDanger  UsageStatusEnum = "danger"
	//
	//ResourceDefaultCN UsageStatusEnum = "默认"
	//ResourceSafeCN    UsageStatusEnum = "安全"
	//ResourceWarningCN UsageStatusEnum = "警告"
	//ResourceDangerCN  UsageStatusEnum = "危险"

	/*
		resource usage status
	*/
	//WorkflowDefault UsageStatusEnum = "default"
	//ResourceSafe    UsageStatusEnum = "safe"
	//ResourceWarning UsageStatusEnum = "warning"
	//ResourceDanger  UsageStatusEnum = "danger"

	// cmp bashboard table
	CMPDashboardChangePageNoOperationKey   cptype.OperationKey = "changePageNo"
	CMPDashboardChangePageSizeOperationKey cptype.OperationKey = "changePageSize"
	CMPDashboardSortByColumnOperationKey   cptype.OperationKey = "changeSort"

	// cmp bashboard clusterFilter
	CMPDashboardFilterOperationKey cptype.OperationKey = "filter"

	// Freeze Button
	CMPDashboardDeleteNode   cptype.OperationKey = "delete"
	CMPDashboardUnfreezeNode cptype.OperationKey = "unfreeze"
	CMPDashboardFreezeNode   cptype.OperationKey = "freeze"
)
var (
	NodeNotFoundErr           = errors.New("node not found")
	ClusterNotFoundErr        = errors.New("cluster not found")
	PodNotFoundErr            = errors.New("pod not found")
	OperationsEmptyErr        = errors.New("operation is empty")
	ResourceEmptyErr          = errors.New("node resource is empty")
	ProtocolComponentEmptyErr = errors.New("component is nil or property empty")
	BundleEmptyErr            = errors.New("bundle is empty")
	NothingToBeDoneErr        = errors.New("nothing to be done")

	TypeNotAvailableErr = errors.New("type not available")
	ResourceNotFoundErr = errors.New("resource type not available")

	//util error
	PtrRequiredErr = errors.New("ptr is required")
)
