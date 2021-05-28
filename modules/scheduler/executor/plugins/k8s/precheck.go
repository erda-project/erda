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

package k8s

import (
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	statusUnschedulable = "unschedulable"
)

func precheck(sg *apistructs.ServiceGroup, resourceinfo apistructs.ClusterResourceInfoData) (
	apistructs.ServiceGroupPrecheckData, error) {
	r := map[string][]apistructs.ServiceGroupPrecheckNodeData{}
	serviceAffinityMap := map[string]string{}
	for _, svc := range sg.Services {
		// svcLabels structs example  [][2][]string
		// [
		//	[
		// 		[erda/org-erda erda/workspace-test erda/stateless-service]  -> must exist
		//		[erda/locked erda/location]								    -> must doesn't exist
		//  ]
		// ]
		svcLabels := make([][2][]string, 0)
		cons := constraintbuilders.K8S(&sg.ScheduleInfo2, &svc, nil, nil)
		if cons.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			svcLabels = extractLabels(cons.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.
				NodeSelectorTerms)
		}
		serviceAffinityMap[svc.Name] = pp(svcLabels)
		for ip, node := range resourceinfo.Nodes {
			if !matchNodeLabels(node.Labels, svcLabels) {
				r[svc.Name] = append(r[svc.Name],
					apistructs.ServiceGroupPrecheckNodeData{
						IP:     ip,
						Status: statusUnschedulable,
						Info: fmt.Sprintf("标签限制: [%s], 节点标签: %v",
							pp(svcLabels), node.Labels),
					})
			} else {
				r[svc.Name] = append(r[svc.Name],
					apistructs.ServiceGroupPrecheckNodeData{
						IP:     ip,
						Status: "ok",
					})
			}
		}
	}
	infoList := []string{}
	statusList := []string{}
	for service, results := range r {
		status := statusUnschedulable
		for _, result := range results {
			if result.Status == "ok" {
				status = "ok"
			}
		}
		if status != "ok" {
			infoList = append(infoList, fmt.Sprintf("%s: %s", service, serviceAffinityMap[service]))
		}
		statusList = append(statusList, status)
	}
	resultStatus := "ok"
	for _, status := range statusList {
		if status != "ok" {
			resultStatus = statusUnschedulable
		}
	}
	info := ""
	if resultStatus == statusUnschedulable {
		info = strutil.Join(infoList, "\n", true)
	}
	return apistructs.ServiceGroupPrecheckData{
		Nodes:  r,
		Status: resultStatus,
		Info:   info,
	}, nil
}

// extractLabels
// return [(exist_label_list, not_exist_label_list), (exist_label_list, not_exist_label_list),...]
// The relationship between different tuples is or
func extractLabels(terms []v1.NodeSelectorTerm) [][2][]string {
	r := [][2][]string{}
	for _, t := range terms {
		exist := []string{}
		notexist := []string{}
		for _, expr := range t.MatchExpressions {
			if expr.Operator == "Exists" {
				exist = append(exist, expr.Key)
			} else if expr.Operator == "DoesNotExist" {
				notexist = append(notexist, expr.Key)
			}
		}
		r = append(r, [2][]string{exist, notexist})
	}
	return r
}

func matchNodeLabels(nodeLabels []string, svcLabels [][2][]string) bool {
	// If labels schedule already disabled, all nodes can be matched.
	if len(svcLabels) == 0 {
		return true
	}

	sort.Strings(nodeLabels)

	for _, labelTuple := range svcLabels {
		isSuccess := true
		for _, exist := range labelTuple[0] {
			i := sort.SearchStrings(nodeLabels, exist)
			if i == len(nodeLabels) || nodeLabels[i] != exist {
				isSuccess = false
			}
		}
		for _, notExist := range labelTuple[1] {
			i := sort.SearchStrings(nodeLabels, notExist)
			if i != len(nodeLabels) && nodeLabels[i] == notExist {
				isSuccess = false
			}
		}
		if isSuccess {
			return true
		}
	}

	return false
}

func pp(svcLabels [][2][]string) string {
	descs := make([]string, 0)
	for _, labelTuple := range svcLabels {
		desc := fmt.Sprintf("需要存在标签: %v, 需要不存在标签: %v", labelTuple[0], labelTuple[1])
		descs = append(descs, desc)
	}

	return strutil.Join(descs, " OR ")
}
