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
		cons := constraintbuilders.K8S(&sg.ScheduleInfo2, &svc, nil, nil)
		svclabels := extractLabels(cons.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
		serviceAffinityMap[svc.Name] = pp(svclabels)
		for ip, node := range resourceinfo.Nodes {
			if !matchNodeLabels(node.Labels, svclabels) {
				r[svc.Name] = append(r[svc.Name],
					apistructs.ServiceGroupPrecheckNodeData{
						IP:     ip,
						Status: statusUnschedulable,
						Info: fmt.Sprintf("标签限制: [%s], 节点标签: %v",
							pp(svclabels), node.Labels),
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
// 不同元组之间为 或 关系
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

func matchNodeLabels(nodelabels []string, svclabels [][2][]string) bool {
	sort.Strings(nodelabels)
	for _, labelTuple := range svclabels {
		succ := true
		for _, exist := range labelTuple[0] {
			i := sort.SearchStrings(nodelabels, exist)
			if i == len(nodelabels) || nodelabels[i] != exist {
				succ = false
			}
		}
		for _, notexist := range labelTuple[1] {
			i := sort.SearchStrings(nodelabels, notexist)
			if i != len(nodelabels) && nodelabels[i] == notexist {
				succ = false
			}
		}
		if succ {
			return true
		}
	}
	return false
}

func pp(svclabels [][2][]string) string {
	descs := []string{}
	for _, labelTuple := range svclabels {
		desc := fmt.Sprintf("需要存在标签: %v, 需要不存在标签: %v", labelTuple[0], labelTuple[1])
		descs = append(descs, desc)
	}

	return strutil.Join(descs, " OR ")
}
