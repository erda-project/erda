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

type (
	SteveStatusEnum string
	UsageStatusEnum string
	OrderEnum       string
	WorkflowEnum    string
)

type SteveStatus struct {
	Value      SteveStatusEnum `json:"value,omitempty"`
	RenderType string          `json:"renderType"`
	Status     SteveStatusEnum `json:"status"`
	Tip        string          `json:"tip"`
}

var (
	ResourceNormal  UsageStatusEnum = "normal"
	ResourceSuccess UsageStatusEnum = "Success"
	ResourceError   UsageStatusEnum = "error"
	ResourceWarning UsageStatusEnum = "warning"
	ResourceDanger  UsageStatusEnum = "danger"

	Asc  OrderEnum = "ascend"
	Desc OrderEnum = "descend"

	Deployments  WorkflowEnum = "Deployments"
	StatefulSets WorkflowEnum = "StatefulSets"
	DaemonSets   WorkflowEnum = "DaemonSets"
	Jobs         WorkflowEnum = "Jobs"
	CronJobs     WorkflowEnum = "CronJobs"

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

type ChartDataItem struct {
	Value float64 `json:"value"`
	Time  int64   `json:"time"`
}
