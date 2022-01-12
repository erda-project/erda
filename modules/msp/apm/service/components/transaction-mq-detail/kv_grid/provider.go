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

package kv_grid

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/kv"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/kv/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/msp/apm/service/datasources"
	"github.com/erda-project/erda/modules/msp/apm/service/view/card"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
)

const (
	totalCount  string = "totalCount"
	avgRps      string = "avgRps"
	avgDuration string = "avgDuration"
	slowCount   string = "slowCount"
	errorCount  string = "errorCount"
	errorRate   string = "errorRate"
)

type provider struct {
	impl.DefaultKV
	Log        logs.Logger
	I18n       i18n.Translator               `autowired:"i18n" translator:"msp-i18n"`
	Metric     metricpb.MetricServiceServer  `autowired:"erda.core.monitor.metric.MetricService"`
	DataSource datasources.ServiceDataSource `autowired:"component-protocol.components.datasources.msp-service"`
}

// RegisterInitializeOp .
func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		lang := sdk.Lang
		startTime := int64(p.StdInParamsPtr.Get("startTime").(float64))
		endTime := int64(p.StdInParamsPtr.Get("endTime").(float64))
		tenantId := p.StdInParamsPtr.Get("tenantId").(string)
		serviceId := p.StdInParamsPtr.Get("serviceId").(string)
		layerPath := p.StdInParamsPtr.Get("layerPath").(string)
		ctx := context.WithValue(context.Background(), common.LangKey, lang)

		data := kv.Data{}
		var list []*kv.KV
		switch sdk.Comp.Name {
		case totalCount:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeReqCount, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		case avgRps:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeRps, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		case avgDuration:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeAvgDuration, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		case slowCount:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeSlowCount, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		case errorCount:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeErrorCount, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		case errorRate:
			cell, err := p.DataSource.GetCard(ctx, card.CardTypeErrorRate, startTime, endTime, tenantId, serviceId, common.TransactionLayerMq, layerPath)
			if err != nil {
				p.Log.Error("failed to get card: %s", err)
				break
			}
			list = append(list, cell)
		}
		data.List = list
		p.StdDataPtr = &data
	}
}

// RegisterRenderingOp .
func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultKV = impl.DefaultKV{}
	v := reflect.ValueOf(p)
	v.Elem().FieldByName("Impl").Set(v)
	compName := "kvGrid"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "transaction-mq-detail",
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
	name := "component-protocol.components.transaction-mq-detail.kvGrid"
	cpregister.AllExplicitProviderCreatorMap[name] = nil
	servicehub.Register(name, &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
