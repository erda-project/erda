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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/code-coverage/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	defaultListSize = 7
	timeFormat      = "01-02 15:04"
	goTimeFormat    = "2006-01-02 15:03:04"
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
	Value    []int64 `json:"value"`
	RecordID uint64  `json:"recordID"`
}

type Operation struct {
	Key      string                 `json:"key"`
	Reload   bool                   `json:"reload"`
	FillMeta string                 `json:"fillMeta"`
	Meta     map[string]interface{} `json:"meta"`
}

func (ca *ComponentAction) setProps(data apistructs.CodeCoverageExecRecordData) {
	ca.Type = "Chart"
	ca.Props = make(map[string]interface{})
	ca.Props["title"] = "项目代码覆盖率趋势"
	ca.Props["chartType"] = "line"
	var timeList []string
	var valueLst []interface{}
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

func (ca *ComponentAction) setDefaultTimeRange() {
	now := time.Now()
	oneMonthAgo := now.AddDate(0, 0, -30)
	oneMonthRange := []int64{oneMonthAgo.Unix() * 1000, now.Unix() * 1000}
	ca.State.Value = oneMonthRange
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := ca.GenComponentState(c); err != nil {
		return err
	}
	svc := ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)
	ca.svc = svc
	sdk := cputil.SDK(ctx)
	projectIDStr := sdk.InParams["projectId"].(string)
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
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
	case cptype.InitializeOperation, cptype.DefaultRenderingKey:
		if len(ca.State.Value) < 2 {
			ca.setDefaultTimeRange()
		}
		start, end := convertTimeRange(ca.State.Value)
		recordRsp, err := svc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
			ProjectID: projectID,
			TimeBegin: start,
			TimeEnd:   end,
			Statuses:  []apistructs.CodeCoverageExecStatus{apistructs.SuccessStatus},
			Asc:       true,
		})
		if err != nil {
			return err
		}
		ca.setProps(recordRsp)
		if len(recordRsp.List) > 0 {
			ca.State.RecordID = recordRsp.List[len(recordRsp.List)-1].ID
		}
		ca.Operations = getOperations()
	case cptype.RenderingOperation:
		if len(ca.State.Value) < 2 {
			ca.setDefaultTimeRange()
		}
		start, end := convertTimeRange(ca.State.Value)
		recordRsp, err := ca.svc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
			ProjectID: uint64(projectID),
			Statuses:  []apistructs.CodeCoverageExecStatus{apistructs.SuccessStatus},
			TimeBegin: start,
			TimeEnd:   end,
			Asc:       true,
		})
		if err != nil {
			return err
		}
		ca.setProps(recordRsp)
		if len(recordRsp.List) > 0 {
			ca.State.RecordID = recordRsp.List[len(recordRsp.List)-1].ID
		}
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "codeCoverChart", func() servicehub.Provider {
		return &ComponentAction{}
	})
}

func convertTimeRange(timeRange []int64) (startTime, endTime string) {
	start, end := timeRange[0], timeRange[1]
	startT := time.Unix(start/1000, 0)
	endT := time.Unix(end/1000, 0)
	return startT.Format(goTimeFormat), endT.Format(goTimeFormat)
}
