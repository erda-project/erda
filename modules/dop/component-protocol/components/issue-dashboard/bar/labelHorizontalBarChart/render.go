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

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	issue_svc "github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
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

	stateMap := make(map[uint64]dao.IssueState)
	for _, i := range f.State.IssueStateList {
		stateMap[i.ID] = i
	}

	bugList := common.IssueListFilter(f.State.IssueList, func(i int) bool {
		v := f.State.IssueList[i].FilterPropertyRetriever(f.State.Values.Type)
		return f.State.Values.Value == nil || strutil.Exist(f.State.Values.Value, v)
	})

	bugMap := make(map[uint64]*dao.IssueItem)
	for i := range bugList {
		issue := &f.State.IssueList[i]
		if issue.Type != apistructs.IssueTypeBug {
			continue
		}
		bugMap[issue.ID] = issue
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

	handler := common.StackRetriever(f.State.Values.Type)

	series, colors, realY := common.GroupToVerticalBarData(labelList, handler, nil, func(label interface{}) string {
		l := label.(*model.LabelIssueItem)
		if l == nil {
			return ""
		}
		return l.LabelRel.Name
	}, common.GetStackBarSingleSeriesConverter(), 500)

	bar := charts.NewBar()
	bar.Colors = colors
	bar.MultiSeries = series
	bar.XAxisList[0] = opts.XAxis{
		Type: "value",
	}

	bar.YAxisList[0] = opts.YAxis{
		Type: "category",
		Data: realY,
	}

	props := make(map[string]interface{})
	props["title"] = "按标签分布（TOP 500）"
	props["chartType"] = "bar"
	props["option"] = bar.JSON()

	c.Props = props
	return nil
}
