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

package common

import (
	"errors"
	"github.com/erda-project/erda/apistructs"
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
	RenderType string          `json:"render_type"`
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

	ResourcesTypes = []apistructs.K8SResType{apistructs.K8SDeployment,apistructs.K8SCRONJOB,apistructs.K8SStatefulSet,apistructs.K8SJOB,apistructs.K8SDaemonSet}
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
