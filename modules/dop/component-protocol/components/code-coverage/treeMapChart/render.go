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

package treeMapChart

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider

	ctxBdl *bundle.Bundle
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
	projectName := ""
	if recordID != 0 {
		record, err := c.svc.GetCodeCoverageRecord(recordID)
		if err != nil {
			return err
		}
		reportContent := record.ReportContent
		if len(reportContent) > 0 {
			data = reportContent[0].Nodes
		}
		title = fmt.Sprintf("%s %s", record.TimeCreated.Format("2006-01-02 15:03:04"), title)
		project, err := c.ctxBdl.GetProject(record.ProjectID)
		if err != nil {
			return err
		}
		projectName = project.DisplayName
	}
	c.Props = map[string]interface{}{
		//"requestIgnore": []string{"props"},
		"title": title,
		"style": map[string]interface{}{
			"height": 600,
		},
		"chartType": "treemap",
		"option": map[string]interface{}{
			//"tooltip": map[string]interface{}{
			//	"show":      true,
			//	"formatter": "{@parent}: {@[1]} <br /> {@abc}: {@[2]}",
			//},
			"tooltip": map[string]interface{}{
				"show": true,
			},
			"series": []interface{}{
				map[string]interface{}{
					"name":            projectName,
					"type":            "treemap",
					"roam":            false,
					"leafDepth":       2,
					"width":           "100%",
					"height":          "100%",
					"colorMappingBy":  "value",
					"visualDimension": 8,
					"visualMin":       0,
					"visualMax":       100,
					"data":            data,
					"color":           []string{"maroon", "orange", "yellow", "green", "darkseagreen"},
					"levels": []interface{}{
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.6},
							"itemStyle": map[string]interface{}{
								"gapWidth": 4,
							},
						},
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.6},
							"itemStyle": map[string]interface{}{
								"gapWidth": 2,
							},
						},
						map[string]interface{}{
							"colorSaturation": []interface{}{0.3, 0.5},
							"itemStyle": map[string]interface{}{
								"gapWidth": 1,
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
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	ca.ctxBdl = bdl
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
