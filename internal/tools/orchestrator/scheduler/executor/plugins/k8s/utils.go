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
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
)

func runtimeIDMatch(podRuntimeID string, pod apiv1.Pod) bool {
	if podRuntimeID == "" {
		return true
	}
	runtimeIDFromPod := ""
	runtimeIDFromPod, ok := pod.Labels["DICE_RUNTIME"]
	if !ok {
		runtimeIDFromPod, ok = pod.Labels["DICE_RUNTIME_ID"]
	}

	if runtimeIDFromPod == "" {
		for _, v := range pod.Spec.Containers[0].Env {
			if v.Name == "DICE_RUNTIME" || v.Name == "DICE_RUNTIME_ID" {
				runtimeIDFromPod = v.Value
				break
			}
		}
	}

	if runtimeIDFromPod != "" && runtimeIDFromPod == podRuntimeID {
		return true
	}
	return false
}

func (k *Kubernetes) composeNewKey(keys []string) string {
	var newKey = strings.Builder{}
	for _, key := range keys {
		newKey.WriteString(key)
	}
	return newKey.String()
}

func (k *Kubernetes) DeployInEdgeCluster() bool {
	clusterInfo, err := k.ClusterInfo.Get()
	if err != nil {
		logrus.Warningf("failed to get cluster info, error: %v", err)
		return false
	}

	if clusterInfo[string(apistructs.DICE_IS_EDGE)] != "true" {
		return false
	}

	return true
}

func (k *Kubernetes) getClusterIP(namespace, name string) (string, error) {
	svc, err := k.GetService(namespace, name)
	if err != nil {
		return "", err
	}
	return svc.Spec.ClusterIP, nil
}

func (k *Kubernetes) whichOperator(operator string) (addon.AddonOperator, error) {
	switch operator {
	case "elasticsearch":
		return k.elasticsearchoperator, nil
	case "redis":
		return k.redisoperator, nil
	case "mysql":
		return k.mysqloperator, nil
	case "canal":
		return k.canaloperator, nil
	case "daemonset":
		return k.daemonsetoperator, nil
	case apistructs.AddonSourcecov:
		return k.sourcecovoperator, nil
	case apistructs.AddonRocketMQ:
		return k.rocketmqoperator, nil
	}
	return nil, fmt.Errorf("not found")
}
func (k *Kubernetes) setProjectServiceName(sg *apistructs.ServiceGroup) {
	for index, service := range sg.Services {
		service.ProjectServiceName = k.composeNewKey([]string{service.Name, "-", sg.ID})
		sg.Services[index] = service
	}
}

func (k *Kubernetes) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	r, err := k.resourceInfo.Get(brief)
	if err != nil {
		return r, err
	}

	osr := k.overSubscribeRatio.GetOverSubscribeRatios()
	r.ProdCPUOverCommit = osr.SubscribeRatio.CPURatio
	r.DevCPUOverCommit = osr.DevSubscribeRatio.CPURatio
	r.TestCPUOverCommit = osr.TestSubscribeRatio.CPURatio
	r.StagingCPUOverCommit = osr.StagingSubscribeRatio.CPURatio
	r.ProdMEMOverCommit = osr.SubscribeRatio.MemRatio
	r.DevMEMOverCommit = osr.DevSubscribeRatio.MemRatio
	r.TestMEMOverCommit = osr.TestSubscribeRatio.MemRatio
	r.StagingMEMOverCommit = osr.StagingSubscribeRatio.MemRatio

	return r, nil
}
