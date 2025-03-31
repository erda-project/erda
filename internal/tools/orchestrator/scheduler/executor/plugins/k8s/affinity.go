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

	apiv1 "k8s.io/api/core/v1"
)

func (k *Kubernetes) composeNodeAntiAffinityPreferredWithWorkspace(workspace string) []apiv1.PreferredSchedulingTerm {
	var (
		workspaceKeys           = []string{"dev", "test", "staging", "prod"}
		weightMap               = map[string]int32{"dev": 60, "test": 60, "staging": 80, "prod": 100}
		preferredSchedulerTerms = make([]apiv1.PreferredSchedulingTerm, 0, len(workspaceKeys))
	)

	for index, key := range workspaceKeys {
		if strings.ToLower(workspace) == key {
			workspaceKeys = append(workspaceKeys[:index], workspaceKeys[index+1:]...)
			break
		}
	}

	for _, key := range workspaceKeys {
		preferredSchedulerTerms = append(preferredSchedulerTerms, apiv1.PreferredSchedulingTerm{
			Weight: weightMap[key],
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      fmt.Sprintf("dice/workspace-%s", key),
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		})
	}
	return preferredSchedulerTerms
}

func (k *Kubernetes) composeDeploymentNodeAntiAffinityPreferred(workspace string) []apiv1.PreferredSchedulingTerm {
	preferredSchedulerTerms := k.composeNodeAntiAffinityPreferredWithWorkspace(workspace)
	return append(preferredSchedulerTerms, apiv1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: apiv1.NodeSelectorTerm{
			MatchExpressions: []apiv1.NodeSelectorRequirement{
				{
					Key:      "dice/stateful-service",
					Operator: apiv1.NodeSelectorOpDoesNotExist,
				},
			},
		},
	})
}

func (k *Kubernetes) composeStatefulSetNodeAntiAffinityPreferred(workspace string) []apiv1.PreferredSchedulingTerm {
	preferredSchedulerTerms := k.composeNodeAntiAffinityPreferredWithWorkspace(workspace)
	return append(preferredSchedulerTerms, apiv1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: apiv1.NodeSelectorTerm{
			MatchExpressions: []apiv1.NodeSelectorRequirement{
				{
					Key:      "dice/stateless-service",
					Operator: apiv1.NodeSelectorOpDoesNotExist,
				},
			},
		},
	})
}
