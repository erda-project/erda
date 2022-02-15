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

package unrecover_alert_chart

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/msp/apm/alert/components/msp-alert-overview/common"
	"github.com/erda-project/erda/pkg/common/errors"
)

const parseLayout = "2006-01-02T15:04:05Z"
const formatLayout = "2006-01-02 15:04:05"

var sharedMetricq metricpb.MetricServiceServer

type provider struct {
	Metric metricpb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	sharedMetricq = p.Metric
	return nil
}

func init() {
	servicehub.Register(fmt.Sprintf("%s.%s.provider", common.ScenarioKey, common.ComponentNameUnRecoverAlertChart), &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
	base.InitProviderWithCreator(common.ScenarioKey, common.ComponentNameUnRecoverAlertChart, func() servicehub.Provider {
		return &SimpleChart{}
	})
}

func (s *SimpleChart) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := s.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	chart, err := s.getUnRecoverAlertEventsChart()
	if err != nil {
		return err
	}

	s.Type = "SimpleChart"
	s.Data = Data{
		Main:  strconv.Itoa(s.getLatestChartValue(chart)),
		Sub:   cputil.I18n(ctx, "UnRecoverAlerts"),
		Chart: *chart,
	}

	return s.SetToProtocolComponent(c)
}

func (s *SimpleChart) getUnRecoverAlertEventsChart() (*Chart, error) {
	inParams, err := common.ParseFromCpSdk(s.sdk)
	if err != nil {
		return nil, errors.NewInvalidParameterError("InParams", err.Error())
	}
	statement := fmt.Sprintf("SELECT max_value(sum(if(eq(trigger::tag,'alert'),1,0))-sum(if(eq(trigger::tag,'alert'),0,1)),0) "+
		"FROM analyzer_alert "+
		"WHERE alert_scope::tag=$scope AND alert_scope_id::tag=$scope_id "+
		"GROUP BY time(%s)", common.GetInterval(inParams.StartTime, inParams.EndTime, time.Second, 10))

	params := map[string]*structpb.Value{
		"scope":    structpb.NewStringValue(inParams.Scope),
		"scope_id": structpb.NewStringValue(inParams.ScopeId),
	}

	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(inParams.StartTime, 10),
		End:       strconv.FormatInt(inParams.EndTime, 10),
		Statement: statement,
		Params:    params,
	}

	response, err := s.Metric.QueryWithInfluxFormat(s.sdk.Ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	rows := response.Results[0].Series[0].Rows
	if len(rows) == 0 {
		return nil, errors.NewInternalServerErrorMessage("empty query result")
	}

	var xAxis []string
	var yAxis []int
	for _, row := range rows {
		row.Values[0].GetNumberValue()
		date := row.Values[0].GetStringValue()
		parse, _ := time.ParseInLocation(parseLayout, date, time.Local)
		xAxis = append(xAxis, parse.Format(formatLayout))
		yAxis = append(yAxis, int(row.Values[1].GetNumberValue()))
	}

	chart := &Chart{
		XAxis: xAxis,
		Series: []SeriesData{
			{Name: s.sdk.I18n(common.ComponentNameUnRecoverAlertChart), Data: yAxis},
		},
	}
	return chart, nil
}

func (s *SimpleChart) getLatestChartValue(chart *Chart) int {
	if len(chart.Series) == 0 {
		return 0
	}
	if len(chart.Series[0].Data) == 0 {
		return 0
	}
	return chart.Series[0].Data[len(chart.Series[0].Data)-1]
}
