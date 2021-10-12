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

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
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

	series, colors, realY, total := common.GroupToBarData(labelList, f.State.Values.Value, handler, nil, func(label interface{}) string {
		l := label.(*model.LabelIssueItem)
		if l == nil {
			return ""
		}
		return l.LabelRel.Name
	}, common.GetStackBarSingleSeriesConverter(), 500, false, true)

	per := 100.0
	cnt := len(total)
	if cnt > 0 {
		per = 16 * 100 / float64(cnt)
	}
	if per > 100 {
		per = 100.0
	}
	maxValue := 0
	for _, t := range total {
		if t > maxValue {
			maxValue = t
		}
	}

	bar := charts.NewBar()
	bar.Legend.Show = true
	bar.Tooltip.Show = true
	bar.Tooltip.Trigger = "axis"
	bar.Colors = colors
	bar.MultiSeries = series
	bar.XAxisList[0] = opts.XAxis{
		Type: "value",
		Max:  maxValue,
	}

	bar.YAxisList[0] = opts.YAxis{
		Type: "category",
		Data: realY,
	}

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
	bb["grid"] = map[string]interface{}{"right": 50}
	bb["dataZoom"] = []map[string]interface{}{
		{
			"type":     "inside",
			"orient":   "vertical",
			"zoomLock": true,
			"start":    0,
			"end":      per,
			"throttle": 0,
		},
		{
			"type":       "slider",
			"orient":     "vertical",
			"handleSize": 20,
			"zoomLock":   true,
			"start":      0,
			"end":        per,
			"throttle":   0,
		},
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
