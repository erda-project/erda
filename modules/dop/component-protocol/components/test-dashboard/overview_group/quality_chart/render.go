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

package quality_chart

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/numeral"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "quality_chart",
		func() servicehub.Provider { return &Q{} })
}

type Q struct {
	base.DefaultProvider

	// in params
	projectID uint64

	// depends
	dbClient *dao.DBClient
	coco     *code_coverage.CodeCoverage
}

type (
	Props struct {
		RadarOption map[string]interface{} `json:"option"`
		Style       Style                  `json:"style"`
		Title       string                 `json:"title"`
	}
	Style struct {
		Height uint64 `json:"height"`
	}
)

func (q *Q) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("%v", r)
		}
	}()

	h := gshelper.NewGSHelper(gs)
	q.projectID, _ = strconv.ParseUint(cputil.GetInParamByKey(ctx, "projectID").(string), 10, 64)
	q.dbClient = ctx.Value(types.DBClient).(*dao.DBClient)
	q.coco = ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)

	// calc score
	mtScore := q.calcMtPlanScore(ctx, h)
	atScore := q.calcAtPlanScore(ctx, h)
	bugScore := q.calcBugScore(ctx, h)
	cocoScore := q.calcCodeCoverage(ctx, h)
	bugReopenScore := q.calcBugReopenRate(ctx, h)

	// radar options
	radar := charts.NewRadar()
	radar.Indicator = []*opts.Indicator{
		{Name: cputil.I18n(ctx, "radar-manual-test-plan"), Max: 100, Min: 0, Color: ""},
		{Name: cputil.I18n(ctx, "radar-auto-test-plan"), Max: 100, Min: 0, Color: ""},
		{Name: cputil.I18n(ctx, "radar-issue-bug"), Max: 100, Min: 0, Color: ""},
		{Name: cputil.I18n(ctx, "radar-code-coverage"), Max: 100, Min: 0, Color: ""},
		{Name: cputil.I18n(ctx, "radar-bug-reopen-rate"), Max: 100, Min: 0, Color: ""},
	}
	radar.AddSeries(
		cputil.I18n(ctx, "radar-quality"),
		[]opts.RadarData{
			{
				Name: "",
				Value: []float64{
					polishScore(mtScore),
					polishScore(atScore),
					polishScore(bugScore),
					polishScore(cocoScore),
					polishScore(bugReopenScore),
				},
			},
		},
		charts.WithAreaStyleOpts(opts.AreaStyle{Color: "", Opacity: 0.2}),
		charts.WithLabelOpts(opts.Label{Show: true, Color: "", Position: "", Formatter: ""}),
	)
	radar.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "item"}),
	)
	radar.JSON()

	// set props
	c.Props = Props{
		RadarOption: radar.JSON(),
		Style:       Style{Height: 265},
		Title:       cputil.I18n(ctx, "radar-total-quality-score"),
	}

	return nil
}

// score = SUM(passed)/SUM(executed) * 100
// value range: 0-100
func (q *Q) calcMtPlanScore(ctx context.Context, h *gshelper.GSHelper) float64 {
	mtPlans := h.GetGlobalManualTestPlanList()

	var numCasePassed, numCaseExecuted, numCaseTotal uint64
	for _, plan := range mtPlans {
		numCasePassed += plan.RelsCount.Succ
		numCaseExecuted += plan.RelsCount.Succ + plan.RelsCount.Block + plan.RelsCount.Block
		numCaseTotal += plan.RelsCount.Total
	}

	if numCaseTotal == 0 {
		return 0
	}
	score := float64(numCasePassed) / float64(numCaseTotal) * float64(numCaseExecuted) / float64(numCaseTotal) * 100
	return score
}

// score = SUM(all_at_plan_latest_passed_rate)/NUM(at_plan) * SUM(all_at_plan_latest_passed_rate)/NUM(at_plan) * 100
// value range: 0-100
func (q *Q) calcAtPlanScore(ctx context.Context, h *gshelper.GSHelper) float64 {
	// TODO use at_block value directly
	return 70
}

// bug score: score = 100 - DI (result must >= 0)
// DI = NUM(FATAL)*10 + NUM(SERIOUS)*3 + NUM(NORMAL)*1 + NUM(SLIGHT)*0.1
// value range: 0-100
func (q *Q) calcBugScore(ctx context.Context, h *gshelper.GSHelper) float64 {
	m, err := q.dbClient.CountBugBySeverity(q.projectID, h.GetGlobalSelectedIterationIDs())
	if err != nil {
		panic(err)
	}

	var numFatal, numSerious, numNormal, numSlight uint64
	for severity, count := range m {
		switch severity {
		case apistructs.IssueSeverityFatal:
			numFatal += count
		case apistructs.IssueSeveritySerious:
			numSerious += count
		case apistructs.IssueSeverityNormal:
			numNormal += count
		case apistructs.IssueSeveritySlight:
			numSlight += count
		}
	}

	DI := float64(numFatal*10) + float64(numSlight*3) + float64(numNormal*1) + float64(numSlight)*0.1

	score := 100 - DI

	return score
}

// score = code_coverage_rate * 100
// value range: 0-100
func (q *Q) calcCodeCoverage(ctx context.Context, h *gshelper.GSHelper) float64 {
	// get the latest success record
	data, err := q.coco.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
		ProjectID:      q.projectID,
		PageNo:         1,
		PageSize:       1,
		Asc:            false,
		ReportStatuses: []apistructs.CodeCoverageExecStatus{apistructs.SuccessStatus},
	})
	if err != nil {
		panic(err)
	}
	if len(data.List) == 0 {
		return 0
	}

	score := data.List[0].Coverage
	return score
}

// score = 100 - reopen_rate*100
// reopen_rate =
// value range: 0-100
func (q *Q) calcBugReopenRate(ctx context.Context, h *gshelper.GSHelper) float64 {
	reopenCount, totalCount, err := q.dbClient.BugReopenCount(q.projectID, h.GetGlobalSelectedIterationIDs())
	if err != nil {
		panic(err)
	}
	if totalCount == 0 {
		return 0
	}
	score := float64(reopenCount) / float64(totalCount) * 100
	return score
}

// polishScore set precision to 2, range from 0-100
func polishScore(score float64) float64 {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return numeral.Round(score, 2)
}
