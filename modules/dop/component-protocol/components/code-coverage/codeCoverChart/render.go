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

package codeCoverChart

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/dop/component-protocol/components/code-coverage/common"

	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	defaultListSize = 7
	timeFormat      = "01-02 15:04"
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

type Operation struct {
	Key      string                 `json:"key"`
	Reload   bool                   `json:"reload"`
	FillMeta string                 `json:"fillMeta"`
	Meta     map[string]interface{} `json:"meta"`
}

func (ca *ComponentAction) setProps(data apistructs.CodeCoverageExecRecordData) {
	ca.Type = "Chart"
	ca.Props["title"] = "项目代码覆盖率趋势"
	ca.Props["chartType"] = "line"
	timeList := []string{}
	valueLst := []interface{}{}
	for _, r := range data.List {
		t := r.TimeCreated.Format(timeFormat)
		timeList = append(timeList, t)
		valueLst = append(valueLst, map[string]interface{}{
			"value":      r.Coverage,
			"symbolSize": 24,
			"symbol":     "pin",
		})
	}
	ca.Props["option"] = map[string]interface{}{
		"xAxis": map[string]interface{}{
			"data": timeList,
		},
		"yAxis": map[string]interface{}{
			"axisLabel": map[string]interface{}{
				"formatter": "{value}%",
			},
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "覆盖率趋势图",
				"data": valueLst,
				"areaStyle": map[string]interface{}{
					"opacity": 0.1,
				},
			},
		},
	}
}

// GenComponentState 获取state
func (i *ComponentAction) GenComponentState(c *apistructs.Component) error {
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

func getOperations() map[string]interface{} {
	return map[string]interface{}{
		"click": Operation{
			Key:      "selectChartItem",
			Reload:   true,
			FillMeta: "data",
			Meta:     map[string]interface{}{"data": ""},
		},
	}
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	svc := ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)
	ca.svc = svc
	ca.ctxBdl = bdl
	projectIDStr := bdl.InParams["project_id"]
	projectID, err := strconv.ParseInt(fmt.Sprintf("%v", projectIDStr), 10, 64)
	if err != nil {
		return err
	}

	switch event.Operation {
	case common.CoverChartSelectItemOperationKey:
		recordIDStr := event.OperationData["data"]
		recordID, err := strconv.ParseInt(fmt.Sprintf("%v", recordIDStr), 10, 64)
		if err != nil {
			return err
		}
		ca.State.RecordID = uint64(recordID)
	case cptype.InitializeOperation:
		recordRsp, err := svc.ListCodeCoverageRecode(apistructs.CodeCoverageListRequest{
			ProjectID: uint64(projectID),
			PageSize:  defaultListSize,
		})
		if err != nil {
			return err
		}
		ca.setProps(recordRsp)
		if len(recordRsp.List) > 0 {
			ca.State.RecordID = recordRsp.List[len(recordRsp.List)-1].ID
		}
		ca.Operations = getOperations()
	}
	return nil
}
