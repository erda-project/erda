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

package k8s

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
)

func (k *Kubernetes) GetWorkspaceLeftQuota(ctx context.Context, projectID, workspace string) (cpu, mem int64, err error) {
	cpuQuota, memQuota, err := k.bdl.GetWorkspaceQuota(&apistructs.GetWorkspaceQuotaRequest{
		ProjectID: projectID,
		Workspace: workspace,
	})
	if err != nil {
		return 0, 0, err
	}
	logrus.Infof("get workspace %s of project %s quota: cpu: %d. mem: %d", workspace, projectID, cpuQuota, memQuota)

	namespaces, err := k.bdl.GetWorkspaceNamespaces(&apistructs.GetWorkspaceNamespaceRequest{
		ProjectID: projectID,
		Workspace: workspace,
	})
	if err != nil {
		return 0, 0, err
	}

	cpuQty := resource.NewQuantity(0, resource.DecimalSI)
	memQty := resource.NewQuantity(0, resource.BinarySI)
	for _, namespace := range namespaces {
		pods, err := k.k8sClient.ClientSet.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{})
		if err != nil {
			return 0, 0, err
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase == "Pending" || pod.Status.Phase == "Succeeded" || pod.Status.Phase == "Failed" {
				continue
			}
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests == nil {
					continue
				}
				cpuQty.Add(*container.Resources.Requests.Cpu())
				memQty.Add(*container.Resources.Requests.Memory())
			}
		}
	}

	logrus.Infof("Requested resource for workspace %s in project %s, CPU: %d, Mem: %d\n", workspace, projectID, cpuQty.MilliValue(), memQty.Value())

	leftCPU := cpuQuota - cpuQty.MilliValue()
	leftMem := memQuota - memQty.Value()
	return leftCPU, leftMem, nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (k *Kubernetes) CheckQuota(ctx context.Context, projectID, workspace, runtimeID string, requestsCPU, requestsMem int64, kind string) (bool, error) {
	if projectID == "" || workspace == "" {
		return true, nil
	}
	if requestsCPU <= 0 && requestsMem <= 0 {
		return true, nil
	}
	leftCPU, leftMem, err := k.GetWorkspaceLeftQuota(ctx, projectID, workspace)
	if err != nil {
		return false, err
	}
	leftCPU = max(leftCPU, 0)
	leftMem = max(leftMem, 0)
	reqCPUStr := resourceToString(float64(requestsCPU), "cpu")
	leftCPUStr := resourceToString(float64(leftCPU), "cpu")
	reqMemStr := resourceToString(float64(requestsMem), "memory")
	leftMemStr := resourceToString(float64(leftMem), "memory")

	logrus.Infof("Checking workspace quota, requests cpu:%s cores, left %s cores; requests memory: %s, left %s",
		reqCPUStr, leftCPUStr, reqMemStr, leftMemStr)

	if requestsCPU > leftCPU || requestsMem > leftMem {

		humanLog, primevalLog := "", ""
		switch kind {
		case "stateless":
			humanLog = "部署失败。"
			primevalLog = " failed to deploy."
		case "stateful":
			humanLog = "addon 部署失败。"
			primevalLog = " failed to deploy addon."
		case "update":
			humanLog = "更新失败。"
			primevalLog = " failed to update."
		case "scale":
			humanLog = "扩容失败。"
			primevalLog = " failed to scale."
		}
		if runtimeID != "" {
			if err = k.bdl.CreateErrorLog(&apistructs.ErrorLogCreateRequest{
				ErrorLog: apistructs.ErrorLog{
					ResourceType:   apistructs.RuntimeError,
					ResourceID:     runtimeID,
					OccurrenceTime: strconv.FormatInt(time.Now().Unix(), 10),
					HumanLog: fmt.Sprintf("当前环境资源配额不足，%s请求CPU变化：%s核，剩余：%s核；请求内存变化：%s，剩余：%s",
						humanLog, reqCPUStr, leftCPUStr, reqMemStr, leftMemStr),
					PrimevalLog: fmt.Sprintf("Resource quota is not enough in current workspace,%s Requests CPU : %s core(s), left %s core(s). Request memroy: %s, left %s",
						primevalLog, reqCPUStr, leftCPUStr, reqMemStr, leftMemStr),
					DedupID: fmt.Sprintf("%s-scheduler-error", runtimeID),
				},
			}); err != nil {
				logrus.Errorf("failed to create quota error log when check quota, %v", err)
			} else {
				logrus.Infof("Create/Update quota error log for runtime %s succeeded", runtimeID)
			}
		}
		return false, nil
	}
	return true, nil
}

func getRequestsResources(containers []corev1.Container) (cpu, mem int64) {
	cpuQuantity := resource.NewQuantity(0, resource.DecimalSI)
	memQuantity := resource.NewQuantity(0, resource.BinarySI)
	for _, c := range containers {
		if c.Resources.Requests == nil {
			continue
		}
		cpuQuantity.Add(*c.Resources.Requests.Cpu())
		memQuantity.Add(*c.Resources.Requests.Memory())
	}
	return cpuQuantity.MilliValue(), memQuantity.Value()
}

func resourceToString(resource float64, tp string) string {
	switch tp {
	case "cpu":
		return strconv.FormatFloat(setPrec(resource/1000, 3), 'f', -1, 64)
	case "memory":
		isNegative := 1.0
		if resource < 0 {
			resource = -resource
			isNegative = -1
		}
		units := []string{"B", "K", "M", "G", "T"}
		i := 0
		for resource >= 1<<10 && i < len(units)-1 {
			resource /= 1 << 10
			i++
		}
		return fmt.Sprintf("%s%s", strconv.FormatFloat(setPrec(resource*isNegative, 3), 'f', -1, 64), units[i])
	default:
		return fmt.Sprintf("%.f", resource)
	}
}

func setPrec(f float64, prec int) float64 {
	pow := math.Pow10(prec)
	f = float64(int64(f*pow)) / pow
	return f
}
