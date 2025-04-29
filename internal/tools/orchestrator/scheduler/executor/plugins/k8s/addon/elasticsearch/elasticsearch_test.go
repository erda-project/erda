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

package elasticsearch

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/oversubscriberatio"
)

func TestValidate(t *testing.T) {
	esOperator := ElasticsearchOperator{
		k8s:         nil,
		statefulset: nil,
		ns:          nil,
		service:     nil,
		overcommit:  nil,
		secret:      nil,
		configmap:   nil,
		imageSecret: nil,
		client:      nil,
	}

	testcases := []struct {
		name  string
		input *apistructs.ServiceGroup
		want  error
	}{
		{
			name: "valid USE_OPERATOR",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{},
				},
			},
			want: fmt.Errorf("[BUG] sg need USE_OPERATOR label"),
		},
		{
			name: "USE_OPERATOR is not elasticsearch",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{
						"USE_OPERATOR": "test",
					},
				},
			},
			want: fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'elasticsearch'"),
		},
		{
			name: "not VERSION",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{
						"USE_OPERATOR": "elasticsearch",
					},
				},
			},
			want: fmt.Errorf("[BUG] sg need VERSION label"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := esOperator.Validate(tc.input)
			if err != nil && tc.want == nil {
				t.Errorf("expected no error, got %v", err)
			}
			assert.EqualError(t, err, tc.want.Error())
		})
	}
}

func TestNodeSetsConvert(t *testing.T) {
	esOperator := ElasticsearchOperator{
		overcommit: oversubscriberatio.New(map[string]string{}),
	}

	testcases := []struct {
		name   string
		sg     *apistructs.ServiceGroup
		scname string
	}{
		{
			name: "no custom ik configure",
			sg: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Services: []apistructs.Service{
						{
							Volumes: []apistructs.Volume{
								{
									SCVolume: apistructs.SCVolume{
										Capacity:         int32(50),
										StorageClassName: "sc",
										Snapshot: &apistructs.VolumeSnapshot{
											MaxHistory: -1,
										},
									},
								},
							},
							DeploymentLabels: map[string]string{
								"USE_OPERATOR": "elasticsearch",
							},
							Env: map[string]string{
								"DICE_WORKSPACE": "test",
							},
							Resources: apistructs.Resources{
								Cpu: 0.2,
								Mem: 2000,
							},
						},
					},
					Labels: map[string]string{
						"dice.test": "test",
					},
				},
			},
			scname: "dice-local-volume",
		},
		{
			name: "custom ik configure",
			sg: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Services: []apistructs.Service{
						{
							Name: "test",
							Volumes: []apistructs.Volume{
								{
									SCVolume: apistructs.SCVolume{
										Capacity:         int32(50),
										StorageClassName: "sc",
										Snapshot: &apistructs.VolumeSnapshot{
											MaxHistory: 2,
										},
									},
								},
							},
							DeploymentLabels: map[string]string{
								"USE_OPERATOR": "elasticsearch",
							},
							Env: map[string]string{
								"DICE_WORKSPACE": "test",
								"EXT_DICT":       "https://test.com",
								"EXT_STOP_DICT":  "https://test.com",
							},
							Resources: apistructs.Resources{
								Cpu: 0.2,
								Mem: 2000,
							},
						},
					},

					Labels: map[string]string{
						"dice.test": "test",
					},
				},
			},
			scname: "dice-local-volume",
		},
		{
			name: "custom ik configure",
			sg: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Services: []apistructs.Service{
						{
							Name: "test",
							Volumes: []apistructs.Volume{
								{
									SCVolume: apistructs.SCVolume{
										Capacity:         int32(50),
										StorageClassName: "sc",
										Snapshot: &apistructs.VolumeSnapshot{
											MaxHistory: 2,
										},
									},
								},
							},
							DeploymentLabels: map[string]string{
								"USE_OPERATOR": "elasticsearch",
							},
							Env: map[string]string{
								"DICE_WORKSPACE": "test",
								"EXT_DICT":       "https://test.com",
								"EXT_STOP_DICT":  "https://test.com",
							},
							Resources: apistructs.Resources{
								Cpu: 0.2,
								Mem: 2000,
							},
						},
					},

					Labels: map[string]string{
						"dice.test": "test",
					},
				},
			},
			scname: "alicloud-disk-ssd-on-erda",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := esOperator.NodeSetsConvert(tc.sg, "dice-local-volume", &corev1.NodeAffinity{})
			if err != nil {

			}
		})
	}
}
