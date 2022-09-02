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
	"encoding/json"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

func (k *Kubernetes) IPToHostname(ip string) string {
	nodes, err := k.k8sClient.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("failed to list node, %v", err)
		return ""
	}

	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP && addr.Address == ip {
				return node.Name
			}
		}
	}

	return ""
}

// SetNodeLabels set the labels of k8s node
func (k *Kubernetes) SetNodeLabels(_ executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	// contents in 'hosts' maybe hostname or internalIP, it should be unified into hostname
	nodes, err := k.k8sClient.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("failed to list nodes: %v", err)
		return err
	}
	updatedHosts := make([]string, 0)
	for _, host := range hosts {
		for _, node := range nodes.Items {
			add := false
			for _, addr := range node.Status.Addresses {

				if addr.Address == host {
					add = true
					break
				}
			}
			if add {
				updatedHosts = append(updatedHosts, node.Name)
			}
		}
	}

	for _, host := range updatedHosts {
		prefixedLabels := map[string]*string{}
		node, err := k.k8sClient.ClientSet.CoreV1().Nodes().Get(context.Background(), host,
			metav1.GetOptions{})
		if err != nil {
			return err
		}

		// 1. unset all 'dice/' labels
		for k := range node.Labels {
			if !strutil.HasPrefixes(k, labelconfig.K8SLabelPrefix) {
				continue
			}
			prefixedLabels[k] = nil
		}

		// 2. set labels in param 'labels'
		for k := range labels {
			v := labels[k]
			prefixedKey := k
			if !strutil.HasPrefixes(prefixedKey, labelconfig.K8SLabelPrefix) {
				prefixedKey = strutil.Concat(labelconfig.K8SLabelPrefix, k)
			}
			prefixedLabels[prefixedKey] = &v
		}

		// 3. set them
		var patch struct {
			Metadata struct {
				Labels map[string]*string `json:"labels"` // Use '*string' to cover 'null' case
			} `json:"metadata"`
		}

		patch.Metadata.Labels = prefixedLabels
		patchData, err := json.Marshal(patch)
		if err != nil {
			return err
		}

		if _, err = k.k8sClient.ClientSet.CoreV1().Nodes().Patch(context.Background(), host, types.MergePatchType,
			patchData, metav1.PatchOptions{}); err != nil {
			return err
		}
	}

	return nil
}
