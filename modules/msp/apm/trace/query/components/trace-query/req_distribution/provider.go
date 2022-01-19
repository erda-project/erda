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

package req_distribution

import (
	"context"
	"reflect"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/bubblegraph"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/bubblegraph/impl"
	structure "github.com/erda-project/erda-infra/providers/component-protocol/components/commodel/data-structure"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace/query"
	"github.com/erda-project/erda/modules/msp/apm/trace/query/commom/custom"
	"github.com/erda-project/erda/pkg/math"
)

type provider struct {
	impl.DefaultBubbleGraph
	custom.TraceInParams
	Log          logs.Logger
	I18n         i18n.Translator              `autowired:"i18n"`
	TraceService *query.TraceService          `autowired:"erda.msp.apm.trace.TraceService"`
	Metric       metricpb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
}

// RegisterInitializeOp .
func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		params := p.TraceInParams.InParamsPtr

		if params.TenantId == "" {
			return nil
		}

		//lang := sdk.Lang
		response, err := p.TraceService.GetTraceReqDistribution(context.Background(), *p.TraceInParams.InParamsPtr)
		if err != nil {
			p.Log.Error(err)
		}
		dataBuilder := bubblegraph.NewDataBuilder().WithTitle(p.I18n.Text(sdk.Lang, "traceDistribution")).
			WithYOptions(bubblegraph.NewOptionsBuilder().WithType(structure.Time).WithPrecision(structure.Nanosecond).Build())
		if response == nil {
			p.StdDataPtr = dataBuilder.Build()
			return nil
		}
		rows := response.Results[0].Series[0].Rows
		if rows == nil || len(rows) == 0 {
			p.StdDataPtr = dataBuilder.Build()
			return nil
		}
		for _, row := range response.Results[0].Series[0].Rows {
			timeFormat := row.Values[0].GetStringValue()
			timeFormat = strings.ReplaceAll(timeFormat, "T", " ")
			timeFormat = strings.ReplaceAll(timeFormat, "Z", "")
			x := timeFormat
			y := math.DecimalPlacesWithDigitsNumber(row.Values[1].GetNumberValue(), 2)
			size := row.Values[2].GetNumberValue()

			dataBuilder.WithBubble(bubblegraph.NewBubbleBuilder().
				WithValueX(x).
				WithValueY(y).
				WithValueSize(size).
				WithDimension("Req Distribution").
				Build())
		}
		p.StdDataPtr = dataBuilder.Build()
		return nil
	}
}

// RegisterRenderingOp .
func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultBubbleGraph = impl.DefaultBubbleGraph{}
	v := reflect.ValueOf(p)
	v.Elem().FieldByName("Impl").Set(v)
	compName := "reqDistribution"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "trace-query",
		CompName: compName,
		Creator:  func() cptype.IComponent { return p },
	})
	return nil
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

func init() {
	name := "component-protocol.components.trace-query.reqDistribution"
	cpregister.AllExplicitProviderCreatorMap[name] = nil
	servicehub.Register(name, &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
