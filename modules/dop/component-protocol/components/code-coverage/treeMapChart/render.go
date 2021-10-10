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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider

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
	title := "报告详情"
	if recordID != 0 {
		record, err := c.svc.GetCodeCoverageRecord(recordID)
		if err != nil {
			return err
		}
		data = record.ReportContent
		title = fmt.Sprintf("%s %s", record.TimeCreated.Format("2006-01-02 15:03:04"), title)
	}
	c.Props = map[string]interface{}{
		//"requestIgnore": []string{"props"},
		"title": title,
		"style": map[string]interface{}{
			"height": 600,
		},
		"chartType": "treemap",
		"option": map[string]interface{}{
			"tooltip": map[string]interface{}{
				"show":      true,
				"formatter": "{@parent}: {@[1]} <br /> {@abc}: {@[2]}",
			},
			"series": []interface{}{
				map[string]interface{}{
					"name":           "All",
					"type":           "treemap",
					"roam":           "move",
					"leafDepth":      2,
					"colorMappingBy": "value",
					"data":           data,
					"color":          []string{"#800000", "#F7A76B", "#F7C36B", "#6CB38B", "#8FBC8F"},
					"levels": []interface{}{
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.6},
							"itemStyle": map[string]interface{}{
								"borderColor": "#555",
								"borderWidth": 4,
								"gapWidth":    4,
							},
						},
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.6},
							"itemStyle": map[string]interface{}{
								"borderColorSaturation": "0.7",
								"borderWidth":           2,
								"gapWidth":              2,
							},
						},
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.5},
							"itemStyle": map[string]interface{}{
								"borderColorSaturation": "0.6",
								"gapWidth":              1,
							},
						},
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.5},
						},
					},
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
	svc := ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)
	ca.svc = svc
	ca.Type = "Chart"
	recordID := ca.State.RecordID
	if err := ca.setProps(recordID); err != nil {
		return err
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "treeMapChart", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
