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

package metric_table

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	slow_transaction "github.com/erda-project/erda/modules/msp/apm/service/common/slow-transaction"
	"github.com/erda-project/erda/modules/msp/apm/service/datasources"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	viewtable "github.com/erda-project/erda/modules/msp/apm/service/view/table"
)

type provider struct {
	impl.DefaultTable
	slow_transaction.SlowTransactionInParams
	Log        logs.Logger
	I18n       i18n.Translator               `autowired:"i18n" translator:"msp-i18n"`
	Metric     metricpb.MetricServiceServer  `autowired:"erda.core.monitor.metric.MetricService"`
	DataSource datasources.ServiceDataSource `autowired:"component-protocol.components.datasources.msp-service"`
}

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		filter := slow_transaction.GetFilterFromGlobalState(*sdk.GlobalState)
		pageNo, pagSize := slow_transaction.GetPagingFromGlobalState(*sdk.GlobalState)
		sorts := slow_transaction.GetSortsFromGlobalState(*sdk.GlobalState)

		data, err := p.DataSource.GetTable(context.WithValue(context.Background(), common.LangKey, sdk.Lang),
			&viewtable.SlowTransactionTableBuilder{
				BaseBuildParams: &viewtable.BaseBuildParams{
					TenantId:  p.InParamsPtr.TenantId,
					ServiceId: p.InParamsPtr.ServiceId,
					StartTime: p.InParamsPtr.StartTime,
					EndTime:   p.InParamsPtr.EndTime,
					Layer:     common.TransactionLayerDb,
					LayerPath: p.InParamsPtr.LayerPath,
					FuzzyPath: false,
					OrderBy:   sorts,
					PageNo:    pageNo,
					PageSize:  pagSize,
					Metric:    p.Metric,
				},
				MinDuration: filter.MinDuration,
				MaxDuration: filter.MaxDuration,
			})
		if err != nil {
			p.Log.Error(err)
			(*sdk.GlobalState)[string(cptype.GlobalInnerKeyError)] = err.Error()
			return
		}

		p.StdDataPtr = &table.Data{
			Table: *data,
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpTableChangePage{}.OpKey(): cputil.NewOpBuilder().WithServerDataPtr(&table.OpTableChangePageServerData{}).Build(),
				table.OpTableChangeSort{}.OpKey(): cputil.NewOpBuilder().Build(),
			},
		}
	}
}

func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *provider) RegisterTablePagingOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		slow_transaction.SetPagingToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *provider) RegisterTableChangePageOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		slow_transaction.SetPagingToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *provider) RegisterTableSortOp(opData table.OpTableChangeSort) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		slow_transaction.SetSortsToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterRenderingOp()(sdk)
	}
}

func (p *provider) RegisterBatchRowsHandleOp(opData table.OpBatchRowsHandle) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *provider) RegisterRowSelectOp(opData table.OpRowSelect) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *provider) RegisterRowAddOp(opData table.OpRowAdd) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *provider) RegisterRowEditOp(opData table.OpRowEdit) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *provider) RegisterRowDeleteOp(opData table.OpRowDelete) (opFunc cptype.OperationFunc) {
	return nil
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultTable = impl.DefaultTable{}
	v := reflect.ValueOf(p)
	v.Elem().FieldByName("Impl").Set(v)
	compName := "metricTable"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "transaction-db-slow",
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
	name := "component-protocol.components.transaction-db-slow.metricTable"
	cpregister.AllExplicitProviderCreatorMap[name] = nil
	servicehub.Register(name, &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
