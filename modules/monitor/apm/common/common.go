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

package apm

import (
	"strconv"
)

const (
	Tags                      = "tags"
	Fields                    = "fields"
	Timestamp                 = "timestamp"
	Columns                   = "columns"
	TagsTargetApplicationId   = Tags + Sep4 + "target_application_id"
	TagsTargetRuntimeName     = Tags + Sep4 + "target_runtime_name"
	TagsTargetServiceId       = Tags + Sep4 + "target_service_id"
	TagsTargetServiceName     = Tags + Sep4 + "target_service_name"
	TagsTargetApplicationName = Tags + Sep4 + "target_application_name"
	TagsTargetRuntimeId       = Tags + Sep4 + "target_runtime_id"
	TagsTargetTerminusKey     = Tags + Sep4 + "target_terminus_key"
	TagsServiceMesh           = Tags + Sep4 + "service_mesh"
	TagsSourceApplicationId   = Tags + Sep4 + "source_application_id"
	TagsSourceRuntimeName     = Tags + Sep4 + "source_runtime_name"
	TagsSourceServiceName     = Tags + Sep4 + "source_service_name"
	TagsSourceServiceId       = Tags + Sep4 + "source_service_id"
	TagsSourceApplicationName = Tags + Sep4 + "source_application_name"
	TagsTargetAddonType       = Tags + Sep4 + "target_addon_type"
	TagsSourceRuntimeId       = Tags + Sep4 + "source_runtime_id"
	TagsSourceAddonType       = Tags + Sep4 + "source_addon_type"
	TagsTargetAddonId         = Tags + Sep4 + "target_addon_id"
	TagsTargetAddonGroup      = Tags + Sep4 + "target_addon_group"
	TagsSourceAddonId         = Tags + Sep4 + "source_addon_id"
	TagsSourceAddonGroup      = Tags + Sep4 + "source_addon_group"
	TagsSourceTerminusKey     = Tags + Sep4 + "source_terminus_key"
	TagsComponent             = Tags + Sep4 + "component"
	TagsHost                  = Tags + Sep4 + "host"
	TagsApplicationId         = Tags + Sep4 + "application_id"
	TagsApplicationName       = Tags + Sep4 + "application_name"
	TagsRuntimeId             = Tags + Sep4 + "runtime_id"
	TagsRuntimeName           = Tags + Sep4 + "runtime_name"
	TagsTerminusKey           = Tags + Sep4 + "terminus_key"
	TagsServiceName           = Tags + Sep4 + "service_name"
	TagsServiceId             = Tags + Sep4 + "service_id"
	FieldsCountSum            = Fields + Sep4 + "count_sum"
	FieldElapsedSum           = Fields + Sep4 + "elapsed_sum"
	FieldsErrorsSum           = Fields + Sep4 + "errors_sum"
)

const (
	Sep1              = "-"
	Sep2              = "*"
	Sep3              = "_"
	Sep4              = "."
	EmptyIndex        = Spot + Sep1 + "empty"
	Spot              = "spot"
	TimeForSplitIndex = 24 * 60 * 60 * 1000
)

const (
	// permission resources
	Monitor             string = "Monitor"
	MonitorTopology     string = "monitor_topology"
	Report              string = "report"
	MonitorProjectAlert string = "monitor_project_alert"
	MicroService        string = "microservice_metric"
)

func CreateEsIndices(indexKey string, startTimeMs int64, endTimeMs int64) []string {
	var indices []string
	if startTimeMs > endTimeMs {
		indices = append(indices, EmptyIndex)
	}
	timestampMs := startTimeMs - startTimeMs%TimeForSplitIndex
	endTimeMs = endTimeMs - endTimeMs%TimeForSplitIndex

	for startTimestampMs := timestampMs; startTimestampMs <= endTimeMs; startTimestampMs += TimeForSplitIndex {
		index := Spot + Sep1 + indexKey + Sep1 + Sep2 + Sep1 + strconv.FormatInt(startTimestampMs, 10)
		indices = append(indices, index)
	}
	if len(indices) <= 0 {
		indices = append(indices, EmptyIndex)
	}
	return indices
}
