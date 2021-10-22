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

package ContainerTable

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func TestComponentContainerTable_Transfer(t *testing.T) {
	component := &ContainerTable{
		Data: map[string][]Data{
			"test": {
				{
					Status: Status{
						RenderType: "text",
						Size:       "small",
						Value: StatusValue{
							Label: "test",
							Color: "test",
						},
					},
					Ready: "true",
					Name:  "test",
					Images: Images{
						RenderType: "text",
						Value: Value{
							Text: "test",
						},
					},
					RestartCount: "10",
					Operate: Operate{
						Operations: map[string]Operation{
							"testOp": {
								Key:    "test",
								Text:   "test",
								Reload: true,
							},
						},
						RenderType: "text",
					},
				},
			},
		},
		State: State{
			ClusterName: "testCluster",
			PodID:       "testID",
		},
		Props: Props{
			RequestIgnore: []string{"data"},
			RowKey:        "id",
			Pagination:    true,
			Scroll: Scroll{
				X: 1,
			},
			Columns: []Column{
				{
					Width:     120,
					DataIndex: "test",
					Title:     "test",
					Fixed:     "test",
				},
			},
		},
	}

	expectedData, err := json.Marshal(component)
	if err != nil {
		t.Error(err)
	}

	result := &cptype.Component{}
	component.Transfer(result)
	resultData, err := json.Marshal(result)
	if err != nil {
		t.Error(err)
	}

	if string(expectedData) != string(resultData) {
		t.Errorf("test failed, expected:\n%s\ngot:\n%s", expectedData, resultData)
	}
}
