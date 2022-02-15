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

package table

import (
	"context"
	"time"

	"github.com/ahmetb/go-linq/v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-infra/providers/i18n"
	monitorpb "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/modules/msp/apm/alert/components/msp-alert-event-list/common"
)

type provider struct {
	impl.DefaultTable
	Log                 logs.Logger
	I18n                i18n.Translator              `autowired:"i18n" translator:"msp-alert-event-list"`
	MonitorAlertService monitorpb.AlertServiceServer `autowired:"erda.core.monitor.alert.AlertService"`
}

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		inParams, err := common.ParseFromCpSdk(sdk)
		if err != nil {
			(*sdk.GlobalState)[string(cptype.GlobalInnerKeyError)] = err.Error()
			return nil
		}
		filters := common.NewConfigurableFilterOptions().GetFromGlobalState(*sdk.GlobalState)
		pageNo, pageSize := common.GetPagingFromGlobalState(*sdk.GlobalState)
		sorts := common.GetSortsFromGlobalState(*sdk.GlobalState)

		data, err := p.queryAlertEvents(sdk, sdk.Ctx, inParams, filters, sorts, pageNo, pageSize)
		if err != nil {
			p.Log.Error(err)
			(*sdk.GlobalState)[string(cptype.GlobalInnerKeyError)] = err.Error()
			return nil
		}

		p.StdDataPtr = &table.Data{
			Table: *data,
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpTableChangePage{}.OpKey(): cputil.NewOpBuilder().WithServerDataPtr(&table.OpTableChangePageServerData{}).Build(),
				table.OpTableChangeSort{}.OpKey(): cputil.NewOpBuilder().Build(),
			},
		}
		return nil
	}
}

func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *provider) RegisterTablePagingOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		common.SetPagingToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterInitializeOp()(sdk)
		return nil
	}
}

func (p *provider) RegisterTableChangePageOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		common.SetPagingToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterInitializeOp()(sdk)
		return nil
	}
}

func (p *provider) RegisterTableSortOp(opData table.OpTableChangeSort) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		common.SetSortsToGlobalState(*sdk.GlobalState, opData.ClientData)
		p.RegisterRenderingOp()(sdk)
		return nil
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

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

func (p *provider) queryAlertEvents(sdk *cptype.SDK, ctx context.Context, params *common.InParams, filters *common.ConfigurableFilterOptions, sorts []*common.Sort, pageNo int64, pageSize int64) (*table.Table, error) {
	var sortDefines []*monitorpb.AlertEventSort
	linq.From(sorts).Select(func(i interface{}) interface{} {
		return &monitorpb.AlertEventSort{
			SortField:  i.(*common.Sort).FieldKey,
			Descending: !i.(*common.Sort).Ascending,
		}
	}).ToSlice(&sortDefines)
	events, err := p.MonitorAlertService.GetAlertEvents(ctx, &monitorpb.GetAlertEventRequest{
		Scope:   params.Scope,
		ScopeId: params.ScopeId,
		Condition: &monitorpb.GetAlertEventRequestCondition{
			Name:                 filters.Name,
			AlertLevels:          filters.AlertLevels,
			AlertIds:             filters.AlertIds,
			AlertStates:          filters.AlertStates,
			AlertSources:         filters.AlertSources,
			LastTriggerTimeMsMin: filters.LastTriggerTimeMin,
			LastTriggerTimeMsMax: filters.LastTriggerTimeMax,
		},
		Sorts:    sortDefines,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	// todo@ggp calc related event history count

	t := &table.Table{
		Columns: table.ColumnsInfo{
			Orders: []table.ColumnKey{"Name", "AlertName", "RuleName", "TriggerCount", "AlertLevel", "AlertState", "AlertSource", "LastTriggerTime"},
			ColumnsMap: map[table.ColumnKey]table.Column{
				"Name":            {Title: sdk.I18n("Name")},
				"AlertName":       {Title: sdk.I18n("RuleName")},
				"TriggerCount":    {Title: sdk.I18n("TriggerCount")},
				"AlertLevel":      {Title: sdk.I18n("AlertLevel")},
				"AlertState":      {Title: sdk.I18n("AlertState")},
				"AlertSource":     {Title: sdk.I18n("AlertSource")},
				"LastTriggerTime": {Title: sdk.I18n("LastTriggerTime"), EnableSort: true, FieldBindToOrder: "LastTriggerTime"},
			},
		},
		Total:    uint64(events.Total),
		PageNo:   uint64(pageNo),
		PageSize: uint64(pageSize),
	}

	for _, item := range events.Items {
		t.Rows = append(t.Rows, table.Row{
			ID: table.RowID(item.Id),
			CellsMap: map[table.ColumnKey]table.Cell{
				"Name":            table.NewTextCell(item.Name).Build(),
				"AlertName":       table.NewTextCell(item.AlertName).Build(),
				"TriggerCount":    table.NewTextCell("todo").Build(),
				"AlertLevel":      table.NewTextCell(item.AlertLevel).Build(),
				"AlertState":      table.NewTextCell(item.AlertState).Build(),
				"AlertSource":     table.NewTextCell(item.AlertSource).Build(),
				"LastTriggerTime": table.NewTextCell(time.Unix(item.LastTriggerTime/1e3, 0).Format("2006/01/02 15:04:05")).Build(),
			},
		})
	}

	return t, nil
}

func init() {
	cpregister.RegisterProviderComponent(common.ScenarioKey, common.ComponentNameTable, &provider{})
}
