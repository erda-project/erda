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

var (
	ResourceNormal  UsageStatusEnum = "normal"
	ResourceSuccess UsageStatusEnum = "Success"
	ResourceError   UsageStatusEnum = "error"
	ResourceWarning UsageStatusEnum = "warning"
	ResourceDanger  UsageStatusEnum = "danger"

	Asc  OrderEnum = "ascend"
	Desc OrderEnum = "descend"

	ResourcesTypes = []apistructs.K8SResType{apistructs.K8SDeployment, apistructs.K8SStatefulSet, apistructs.K8SDaemonSet}

	ColorMap = map[string]string{
		"green":         "#6CB38B",
		"purple":        "#975FA0",
		"orange":        "#F7A76B",
		"red":           "#DE5757",
		"brown":         "#A98C72",
		"steelBlue":     "#4E6097",
		"yellow":        "#F7C36B",
		"lightgreen":    "#8DB36C",
		"darkcyan":      "#498E9E",
		"darksalmon":    "#DE6F57",
		"darkslategray": "#2F4F4F",
		"maroon":        "#800000",
		"darkseagreen":  "#8FBC8F",
		"darkslateblue": "#483D8B",
		"darkgoldenrod": "#B8860B",
		"teal":          "#008080",
		"primary":       "#6a549e",
	}

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

func GetPodStatus() string {
	return ""
}
