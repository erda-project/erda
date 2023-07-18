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

package project_report

import (
	"sort"
	"strconv"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
)

const (
	metricGroupName = "project_delivery_report"
)

const (
	labelOrgName = "org_name"

	labelMeta          = "_meta"
	labelMetricScope   = "_metric_scope"
	labelMetricScopeID = "_metric_scope_id"
)

var (
	// TotalRequirementNum requirement type issue total num for all statuses under the project
	TotalRequirementNum = "total_requirement_num"
	// TotalTaskNum task type issue total num for all statuses under the project
	TotalTaskNum = "total_task_num"
)

func generateProjectMetricLabels(projectDto *apistructs.ProjectDTO, orgDto *orgpb.Org) ([]string, []string, map[string]string) {

	metricLabelMap := map[string]string{
		labelMeta:          "true",
		labelMetricScope:   "org",
		labelMetricScopeID: orgDto.Name,
		labelOrgName:       orgDto.Name,
		// tenant
		apistructs.LabelOrgID:       strconv.FormatUint(projectDto.OrgID, 10),
		apistructs.LabelOrgName:     orgDto.Name,
		apistructs.LabelProjectID:   strconv.FormatUint(projectDto.ID, 10),
		apistructs.LabelProjectName: projectDto.Name,
	}

	var metricsKeys []string
	for k := range metricLabelMap {
		metricsKeys = append(metricsKeys, k)
	}
	// the order of metricsKeys is guaranteed to be consistent,
	// so that prometheus will not report an error
	sort.Strings(metricsKeys)

	var metricsValues []string
	for _, k := range metricsKeys {
		metricsValues = append(metricsValues, metricLabelMap[k])
	}

	return metricsKeys, metricsValues, metricLabelMap
}
