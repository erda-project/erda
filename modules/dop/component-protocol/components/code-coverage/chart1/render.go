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

package head

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
	ctxBdl protocol.ContextBundle
	svc    *code_coverage.CodeCoverage

	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
	Data       map[string]interface{}
}

type State struct {
	RecordID uint64 `json:"recordID"`
}

func (i *ComponentAction) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	fmt.Println(state)
	i.State = state
	return nil
}

func (c *ComponentAction) setProps(recordID uint64) error {
	var data []*apistructs.CodeCoverageNode
	if recordID != 0 {
		record, err := c.svc.GetCodeCoverageRecord(recordID)
		if err != nil {
			return err
		}
		data = record.ReportContent
	}
	c.Props = map[string]interface{}{
		"title": "矩阵图",
		"style": map[string]interface{}{
			"height": 600,
		},
		"chartType": "treemap",
		"option": map[string]interface{}{
			"series": []interface{}{
				map[string]interface{}{
					"name": "All",
					"top":  80,
					"type": "treemap",
					"label": map[string]interface{}{
						"show":      true,
						"formatter": "{b}",
					},
					"itemStyle": map[string]interface{}{
						"normal": map[string]interface{}{
							"borderColor": "black",
						},
					},
					"visualMin":       -100,
					"visualMax":       100,
					"visualDimension": 3,
					"levels": []interface{}{
						map[string]interface{}{
							"itemStyle": map[string]interface{}{
								"borderWidth": 3,
								"borderColor": "#333",
								"gapWidth":    3,
							},
						},
						map[string]interface{}{
							"color": []interface{}{
								"942e38",
								"#aaa",
								"#269f3c",
							},
							"colorMappingBy": "value",
							"itemStyle": map[string]interface{}{
								"gapWidth": 1,
							},
						},
					},
					"data": data,
				},
			},
		},
	}
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := ca.GenComponentState(c); err != nil {
		return err
	}
	ca.Type = "Chart"
	recordID := ca.State.RecordID
	if err := ca.setProps(recordID); err != nil {
		return err
	}
	return nil
}
