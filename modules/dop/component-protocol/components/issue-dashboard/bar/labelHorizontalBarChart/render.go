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

package labelHorizontalBarChart

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/chartbuilders"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/stackhandlers"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	issue_svc "github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "labelHorizontalBarChart",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (f *ComponentAction) setInParams(ctx context.Context) error {
	b, err := json.Marshal(cputil.SDK(ctx).InParams)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &f.InParams); err != nil {
		return err
	}

	f.InParams.ProjectID, err = strconv.ParseUint(f.InParams.FrontEndProjectID, 10, 64)
	return err
}

func (f *ComponentAction) getState(c *cptype.Component) error {
	d, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, &f.State)
}

func (f *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	f.issueSvc = ctx.Value(types.IssueService).(*issue_svc.Issue)
	if err := f.setInParams(ctx); err != nil {
		return err
	}

	if err := f.getState(c); err != nil {
		return err
	}

	helper := gshelper.NewGSHelper(gs)

	issueList := helper.GetIssueList()
	bugMap := make(map[uint64]*dao.IssueItem)
	for i := range issueList {
		issue := issueList[i]
		bugMap[issue.ID] = &issue
	}

	labels, err := f.issueSvc.GetIssueLabelsByProjectID(f.InParams.ProjectID)
	if err != nil {
		return err
	}
	var labelList []interface{}
	for i := range labels {
		l := &labels[i]
		bug, ok := bugMap[l.RefID]
		if !ok {
			continue
		}
		labelList = append(labelList, &model.LabelIssueItem{
			LabelRel: l,
			Bug:      bug,
		})
	}

	handler := stackhandlers.NewStackRetriever(
		stackhandlers.WithIssueStateList(helper.GetIssueStateList()),
		stackhandlers.WithIssueStageList(helper.GetIssueStageList()),
	).GetRetriever(f.State.Values.Type)

	builder := &chartbuilders.BarBuilder{
		Items:        labelList,
		StackHandler: handler,
		FixedXAxisOrTop: chartbuilders.FixedXAxisOrTop{
			Top:      500,
			XIndexer: getXIndexer(),
			XDisplayConverter: func(opt *chartbuilders.FixedXAxisOrTop) opts.XAxis {
				return opts.XAxis{
					Type: "value",
					Max:  opt.MaxValue,
				}
			},
		},
		YAxisOpt: chartbuilders.YAxisOpt{
			YDisplayConverter: func(opt *chartbuilders.YAxisOpt) opts.YAxis {
				return opts.YAxis{
					Type: "category",
					Data: opt.YAxis,
				}
			},
		},
		StackOpt: chartbuilders.StackOpt{
			SkipEmpty: true,
		},
		DataHandleOpt: chartbuilders.DataHandleOpt{
			SeriesConverter: chartbuilders.GetStackBarSingleSeriesConverter(),
			DataWhiteList:   f.State.Values.Value,
		},
	}

	if err := builder.Generate(); err != nil {
		return err
	}

	bar := builder.Result.Bar
	bar.Legend.Show = true
	bar.Tooltip.Show = true
	bar.Tooltip.Trigger = "axis"

	bb := bar.JSON()

	bb["animation"] = false
	n := make([]map[string]interface{}, 0)
	buf, err := json.Marshal(bb["series"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &n); err != nil {
		return err
	}
	for i := range n {
		n[i]["barWidth"] = 10
	}
	bb["series"] = n
	pageSize := 16
	if builder.Result.Size > pageSize {
		endIdx := builder.Result.Size - 1
		startIdx := endIdx - 16
		bb["grid"] = map[string]interface{}{"right": 50}
		bb["dataZoom"] = []map[string]interface{}{
			{
				"type":       "inside",
				"orient":     "vertical",
				"zoomLock":   true,
				"startValue": startIdx,
				"endValue":   endIdx,
				"throttle":   0,
			},
			{
				"type":       "slider",
				"orient":     "vertical",
				"handleSize": 20,
				"zoomLock":   true,
				"startValue": startIdx,
				"endValue":   endIdx,
				"throttle":   0,
			},
		}
	}

	props := make(map[string]interface{})
	props["title"] = "按标签分布（TOP 500）"
	props["chartType"] = "bar"
	props["option"] = bb
	props["style"] = map[string]interface{}{"height": 400}

	c.Props = props
	c.State = nil
	return nil
}

func getXIndexer() func(interface{}) string {
	return func(item interface{}) string {
		l := item.(*model.LabelIssueItem)
		if l == nil {
			return ""
		}
		return l.LabelRel.Name
	}
}
