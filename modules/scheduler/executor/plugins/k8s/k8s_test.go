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
	"testing"

	"gotest.tools/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestComposeDeploymentNodeAffinityPreferredWithServiceWorkspace(t *testing.T) {
	k := Kubernetes{}
	workspace := "DEV"

	deploymentPreferred := []apiv1.PreferredSchedulingTerm{
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-test",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 80,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-staging",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-prod",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/stateful-service",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
	}

	resPreferred := k.composeDeploymentNodeAntiAffinityPreferred(workspace)
	assert.DeepEqual(t, deploymentPreferred, resPreferred)
}

func TestComposeStatefulSetNodeAffinityPreferredWithServiceWorkspace(t *testing.T) {
	k := Kubernetes{}
	workspace := "PROD"

	statefulSetPreferred := []apiv1.PreferredSchedulingTerm{
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-dev",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-test",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 80,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-staging",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/stateless-service",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
	}
	resPreferred := k.composeStatefulSetNodeAntiAffinityPreferred(workspace)
	assert.DeepEqual(t, statefulSetPreferred, resPreferred)
}
