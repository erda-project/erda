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

package cputil

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/rancher/wrangler/pkg/data"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
)

// ParseWorkloadStatus get status for workloads from .metadata.fields
func ParseWorkloadStatus(obj data.Object) (string, string, error) {
	kind := obj.String("kind")
	fields := obj.StringSlice("metadata", "fields")

	switch kind {
	case "Deployment":
		if len(fields) != 8 {
			return "", "", fmt.Errorf("deployment %s has invalid fields length", obj.String("metadata", "name"))
		}
		// up-to-date and available
		if fields[2] == fields[3] {
			return "Active", "green", nil
		} else {
			return "Error", "red", nil
		}
	case "DaemonSet":
		if len(fields) != 11 {
			return "", "", fmt.Errorf("daemonset %s has invalid fields length", obj.String("metadata", "name"))
		}
		// desired and ready
		if fields[1] == fields[3] {
			return "Active", "green", nil
		} else {
			return "Error", "red", nil
		}
	case "StatefulSet":
		if len(fields) != 5 {
			return "", "", fmt.Errorf("statefulSet %s has invalid fields length", obj.String("metadata", "name"))
		}
		//
		readyPods := strings.Split(fields[1], "/")
		if readyPods[0] == readyPods[1] {
			return "Active", "green", nil
		} else {
			return "Error", "red", nil
		}
	case "Job":
		if len(fields) != 7 {
			return "", "", fmt.Errorf("job %s has invalid fields length", obj.String("metadata", "name"))
		}
		active := obj.String("status", "active")
		failed := obj.String("status", "failed")
		if failed != "" && failed != "0" {
			return "Failed", "red", nil
		} else if active != "" && active != "0" {
			return "Active", "green", nil
		} else {
			return "Succeeded", "steelblue", nil
		}
	case "CronJob":
		return "Active", "green", nil
	default:
		return "", "", fmt.Errorf("valid workload kind: %v", kind)
	}
}

// ParseWorkloadID get workloadKind, namespace and name from id
func ParseWorkloadID(id string) (apistructs.K8SResType, string, string, error) {
	splits := strings.Split(id, "_")
	if len(splits) != 3 {
		return "", "", "", fmt.Errorf("invalid workload id: %s", id)
	}
	return apistructs.K8SResType(splits[0]), splits[1], splits[2], nil
}

// GetWorkloadAgeAndImage get age and image for workloads from .metadata.fields
func GetWorkloadAgeAndImage(obj data.Object) (string, string, error) {
	kind := obj.String("kind")
	fields := obj.StringSlice("metadata", "fields")

	switch kind {
	case "Deployment":
		if len(fields) != 8 {
			return "", "", fmt.Errorf("deployment %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[4], fields[6], nil
	case "DaemonSet":
		if len(fields) != 11 {
			return "", "", fmt.Errorf("daemonset %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[7], fields[9], nil
	case "StatefulSet":
		if len(fields) != 5 {
			return "", "", fmt.Errorf("statefulSet %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[2], fields[4], nil
	case "Job":
		if len(fields) != 7 {
			return "", "", fmt.Errorf("job %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[3], fields[5], nil
	case "CronJob":
		if len(fields) != 9 {
			return "", "", fmt.Errorf("cronJob %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[5], fields[7], nil
	default:
		return "", "", fmt.Errorf("invalid workload kind: %s", kind)
	}
}

// ResourceToString return resource with unit
// Only support resource.DecimalSI and resource.BinarySI format
// Original unit is m (for DecimalSI) or B (for resource.BinarySI)
// Accurate to 3 decimal places. Zero in suffix will be removed
func ResourceToString(sdk *cptype.SDK, res float64, format resource.Format) string {
	switch format {
	case resource.DecimalSI:
		return fmt.Sprintf("%s%s", strconv.FormatFloat(setPrec(res/1000, 3), 'f', -1, 64), sdk.I18n("core"))
	case resource.BinarySI:
		units := []string{"B", "K", "M", "G", "T"}
		i := 0
		for res >= 1<<10 && i < len(units)-1 {
			res /= 1 << 10
			i++
		}
		return fmt.Sprintf("%s%s", strconv.FormatFloat(setPrec(res, 3), 'f', -1, 64), units[i])
	default:
		return fmt.Sprintf("%d", int64(res))
	}
}

func setPrec(f float64, prec int) float64 {
	pow := math.Pow10(prec)
	f = float64(int64(f*pow)) / pow
	return f
}

// CalculateNodeRes calculate unallocated cpu, memory and left cpu, mem, pods for given node and its allocated cpu, memory
func CalculateNodeRes(node data.Object, allocatedCPU, allocatedMem, allocatedPods int64) (unallocatedCPU, unallocatedMem, leftCPU, leftMem, leftPods int64) {
	allocatableCPUQty, _ := resource.ParseQuantity(node.String("status", "allocatable", "cpu"))
	allocatableMemQty, _ := resource.ParseQuantity(node.String("status", "allocatable", "memory"))
	allocatablePodQty, _ := resource.ParseQuantity(node.String("status", "allocatable", "pods"))
	capacityCPUQty, _ := resource.ParseQuantity(node.String("status", "capacity", "cpu"))
	capacityMemQty, _ := resource.ParseQuantity(node.String("status", "capacity", "memory"))

	unallocatedCPU = capacityCPUQty.MilliValue() - allocatableCPUQty.MilliValue()
	unallocatedMem = capacityMemQty.Value() - allocatableMemQty.Value()
	leftCPU = allocatableCPUQty.MilliValue() - allocatedCPU
	leftMem = allocatableMemQty.Value() - allocatedMem
	leftPods = allocatablePodQty.Value() - allocatedPods
	return
}
