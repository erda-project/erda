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

package slow_req_distribution

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/bubblegraph/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	slow_transaction "github.com/erda-project/erda/modules/msp/apm/service/common/slow-transaction"
	"github.com/erda-project/erda/modules/msp/apm/service/datasources"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
)

type provider struct {
	impl.DefaultBubbleGraph
	slow_transaction.SlowTransactionInParams
	Log        logs.Logger
	I18n       i18n.Translator               `autowired:"i18n" translator:"msp-i18n"`
	Metric     metricpb.MetricServiceServer  `autowired:"erda.core.monitor.metric.MetricService"`
	DataSource datasources.ServiceDataSource `autowired:"component-protocol.components.datasources.msp-service"`
}

// RegisterInitializeOp .
func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		bubble, err := p.DataSource.GetBubbleChart(context.WithValue(context.Background(), common.LangKey, sdk.Lang),
			datasources.BubbleChartSlowReqDistribution,
			p.InParamsPtr.StartTime,
			p.InParamsPtr.EndTime,
			p.InParamsPtr.TenantId,
			p.InParamsPtr.ServiceId,
			common.TransactionLayerCache,
			p.InParamsPtr.LayerPath)

		if err != nil {
			p.Log.Error(err)
			(*sdk.GlobalState)[string(cptype.GlobalInnerKeyError)] = err.Error()
			return
		}

		p.StdDataPtr = bubble
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
	compName := "slowReqDistribution"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "transaction-cache-slow",
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
	name := "component-protocol.components.transaction-cache-slow.slowReqDistribution"
	cpregister.AllExplicitProviderCreatorMap[name] = nil
	servicehub.Register(name, &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
