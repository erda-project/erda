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
	"encoding/json"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
)

func Test_radar(t *testing.T) {
	ctx := context.Background()
	context.WithValue(ctx, cptype.GlobalInnerKeyCtxSDK, &cptype.SDK{})

	radar := charts.NewRadar()
	radar.Indicator = []*opts.Indicator{
		{Name: "test-case", Max: 100, Min: 0, Color: ""},
		{Name: "config-sheet", Max: 100, Min: 0, Color: ""},
		{Name: "test-plan", Max: 100, Min: 0, Color: ""},
		{Name: "code-quality", Max: 100, Min: 0, Color: ""},
		{Name: "issue-bug", Max: 100, Min: 0, Color: ""},
	}
	radar.AddSeries(
		"quality",
		[]opts.RadarData{
			{
				Name:  "",
				Value: []uint64{100, 90, 80, 70, 60},
			},
		},
		charts.WithAreaStyleOpts(opts.AreaStyle{Color: "", Opacity: 0.2}),
		charts.WithLabelOpts(opts.Label{Show: true, Color: "", Position: "", Formatter: ""}),
	)
	radar.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{Show: false, Trigger: "item"}),
		charts.WithTitleOpts(opts.Title{Title: "quality score"}),
	)
	_, err := json.MarshalIndent(radar.JSON(), "", "  ")
	assert.NoError(t, err)
}

func Test_polishToFloat64Score(t *testing.T) {
	type args struct {
		score decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "less than 0",
			args: args{
				score: decimal.NewFromInt(-1),
			},
			want: float64(0),
		},
		{
			name: "bigger than 100",
			args: args{
				score: decimal.NewFromInt(101),
			},
			want: float64(100),
		},
		{
			name: "3 digit",
			args: args{
				score: decimal.NewFromFloat(11.234),
			},
			want: float64(11.23),
		},
		{
			name: "1 digit",
			args: args{
				score: decimal.NewFromFloat(11.2),
			},
			want: float64(11.2),
		},
		{
			name: "no digit",
			args: args{
				score: decimal.NewFromInt(11),
			},
			want: float64(11),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := polishToFloat64Score(tt.args.score); got != tt.want {
				t.Errorf("polishToFloat64Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQ_calcAtPlanScore(t *testing.T) {
	q := Q{}

	// none at plan at all
	hForNone := gshelper.NewGSHelper(nil)
	scoreForNone := q.calcAtPlanScore(context.Background(), hForNone)
	assert.Equal(t, decimal.NewFromInt(0), scoreForNone)

	// normal at plans
	gsForNormal := &cptype.GlobalStateData{}
	hForNormal := gshelper.NewGSHelper(gsForNormal)
	hForNormal.SetGlobalAutoTestPlanList([]*apistructs.TestPlanV2{
		{ // some executed
			ExecuteApiNum: 22,
			SuccessApiNum: 17,
			TotalApiNum:   30,
		},
		{ // empty plan
			ExecuteApiNum: 0,
			SuccessApiNum: 0,
			TotalApiNum:   0,
		},
		{ // all passed
			ExecuteApiNum: 30,
			SuccessApiNum: 30,
			TotalApiNum:   30,
		},
		{ // all failed
			ExecuteApiNum: 30,
			SuccessApiNum: 0,
			TotalApiNum:   30,
		},
	})
	scoreForNormal := q.calcAtPlanScore(context.Background(), hForNormal)
	expectedScoreForNormal := decimal.NewFromFloat((float64(22+0+30+30) / float64(30+0+30+30)) * (float64(17+0+30+0) / float64(30+0+30+30)) * 100)
	assert.Equal(t, expectedScoreForNormal.Round(2), scoreForNormal.Round(2))
}

func TestFloatPrecision(t *testing.T) {
	a := decimal.NewFromFloat(0.3)
	b := decimal.NewFromFloat(0.6)
	c, _ := a.Add(b).Float64()
	assert.Equal(t, float64(0.9), c)
}
