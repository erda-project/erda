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
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/commodel"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline-exec-list/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline-exec-list/common/gshelper"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
)

type provider struct {
	impl.DefaultTable
	ServiceInParams
	Log             logs.Logger
	I18n            i18n.Translator         `autowired:"i18n" translator:"msp-i18n"`
	ProjectPipeline projectpipeline.Service `autowired:"erda.dop.projectpipeline.ProjectPipelineService"`
}

const (
	ColumnPipelineName    table.ColumnKey = "pipelineName"
	ColumnPipelineStatus  table.ColumnKey = "pipelineStatus"
	ColumnCostTime        table.ColumnKey = "costTime"
	ColumnApplicationName table.ColumnKey = "applicationName"
	ColumnBranch          table.ColumnKey = "branch"
	ColumnExecutor        table.ColumnKey = "executor"
	ColumnStartTime       table.ColumnKey = "startTime"

	StateKeyTransactionPaging = "paging"
	StateKeyTransactionSort   = "sort"
)

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		lang := sdk.Lang
		projectID := p.ServiceInParams.InParamsPtr.ProjectId
		pageNo, pageSize := GetPagingFromGlobalState(*sdk.GlobalState)
		sorts := GetSortsFromGlobalState(*sdk.GlobalState)

		var descCols []string
		var ascCols []string
		for _, v := range sorts {
			if v.Ascending {
				descCols = append(descCols, v.FieldKey)
			} else {
				ascCols = append(ascCols, v.FieldKey)
			}
		}

		var req = deftype.ProjectPipelineListExecHistory{
			DescCols:  descCols,
			AscCols:   ascCols,
			PageNo:    uint64(pageNo),
			PageSize:  uint64(pageSize),
			ProjectID: projectID,
		}
		helper := gshelper.NewGSHelper(sdk.GlobalState)
		if helper.GetAppsFilter() != nil {
			req.AppIDList = helper.GetAppsFilter()
		}
		if helper.GetExecutorsFilter() != nil {
			req.Executors = helper.GetExecutorsFilter()
		}
		if helper.GetPipelineNameFilter() != "" {
			req.Name = helper.GetPipelineNameFilter()
		}
		if helper.GetStatuesFilter() != nil {
			req.Statuses = helper.GetStatuesFilter()
		}
		if helper.GetBeginTimeStartFilter() != nil {
			req.StartTimeBegin = *helper.GetBeginTimeStartFilter()
		}
		if helper.GetBeginTimeEndFilter() != nil {
			req.StartTimeEnd = *helper.GetBeginTimeEndFilter()
		}

		result, err := p.ProjectPipeline.ListExecHistory(context.Background(), req)
		if err != nil {
			p.Log.Error("failed to get table data: %s", err)
			return
		}

		tableValue := InitTable(lang, p.I18n)
		tableValue.Total = uint64(result.Data.Total)
		tableValue.PageSize = uint64(pageSize)
		tableValue.PageNo = uint64(pageNo)

		for _, pipeline := range result.Data.Pipelines {
			if pipeline.DefinitionPageInfo == nil {
				continue
			}
			tableValue.Rows = append(tableValue.Rows, pipelineToRow(pipeline, lang, p.I18n))
		}

		p.StdDataPtr = &table.Data{
			Table: tableValue,
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpTableChangePage{}.OpKey(): cputil.NewOpBuilder().WithServerDataPtr(&table.OpTableChangePageServerData{}).Build(),
				table.OpTableChangeSort{}.OpKey(): cputil.NewOpBuilder().Build(),
			}}
	}
}

func GetSortsFromGlobalState(globalState cptype.GlobalStateData) []*common.Sort {
	var sorts []*common.Sort
	if sortCol, ok := globalState[StateKeyTransactionSort]; ok && sortCol != nil {
		var clientSort table.OpTableChangeSortClientData
		clientSort, ok = sortCol.(table.OpTableChangeSortClientData)
		if !ok {
			ok = mapstructure.Decode(sortCol, &clientSort) == nil
		}
		if ok {
			col := clientSort.DataRef
			if col != nil && col.AscOrder != nil {
				sorts = append(sorts, &common.Sort{
					FieldKey:  col.FieldBindToOrder,
					Ascending: *col.AscOrder,
				})
			}
		}
	}
	return sorts
}

func GetPagingFromGlobalState(globalState cptype.GlobalStateData) (pageNo int, pageSize int) {
	pageNo = 1
	pageSize = common.DefaultPageSize
	if paging, ok := globalState[StateKeyTransactionPaging]; ok && paging != nil {
		var clientPaging table.OpTableChangePageClientData
		clientPaging, ok = paging.(table.OpTableChangePageClientData)
		if !ok {
			ok = mapstructure.Decode(paging, &clientPaging) == nil
		}
		if ok {
			pageNo = int(clientPaging.PageNo)
			pageSize = int(clientPaging.PageSize)
		}
	}
	return pageNo, pageSize
}

func pipelineToRow(pipeline apistructs.PagePipeline, lang i18n.LanguageCodes, i18n i18n.Translator) table.Row {
	return table.Row{
		ID:         table.RowID(fmt.Sprintf("pipeline-id-%v", pipeline.ID)),
		Selectable: true,
		Selected:   false,
		CellsMap: map[table.ColumnKey]table.Cell{
			ColumnPipelineName:    table.NewTextCell(pipeline.DefinitionPageInfo.Name).Build(),
			ColumnPipelineStatus:  table.NewTextCell(i18n.Text(lang, string(ColumnPipelineStatus)+pipeline.Status.String())).Build(),
			ColumnCostTime:        table.NewTextCell(fmt.Sprintf("%v s", pipeline.CostTimeSec)).Build(),
			ColumnApplicationName: table.NewTextCell(getApplicationNameFromDefinitionRemote(pipeline.DefinitionPageInfo.SourceRemote)).Build(),
			ColumnBranch:          table.NewTextCell(pipeline.DefinitionPageInfo.SourceRef).Build(),
			ColumnExecutor:        table.NewUserCell(commodel.User{ID: pipeline.DefinitionPageInfo.Creator}).Build(),
			ColumnStartTime:       table.NewTextCell(pipeline.TimeBegin.Format("2006-01-02 15:04:05")).Build(),
		},
		Operations: map[cptype.OperationKey]cptype.Operation{
			table.OpRowSelect{}.OpKey(): cputil.NewOpBuilder().Build(),
		},
	}
}

func getApplicationNameFromDefinitionRemote(remote string) string {
	values := strings.Split(remote, string(filepath.Separator))
	if len(values) != 3 {
		return remote
	}
	return values[2]
}

func InitTable(lang i18n.LanguageCodes, i18n i18n.Translator) table.Table {
	return table.Table{
		Columns: table.ColumnsInfo{
			Orders: []table.ColumnKey{ColumnCostTime, ColumnStartTime},
			ColumnsMap: map[table.ColumnKey]table.Column{
				ColumnPipelineName:    {Title: i18n.Text(lang, string(ColumnPipelineName)), EnableSort: false},
				ColumnPipelineStatus:  {Title: i18n.Text(lang, string(ColumnPipelineStatus)), EnableSort: false},
				ColumnCostTime:        {Title: i18n.Text(lang, string(ColumnCostTime)), EnableSort: true, FieldBindToOrder: string(ColumnCostTime)},
				ColumnApplicationName: {Title: i18n.Text(lang, string(ColumnApplicationName)), EnableSort: false},
				ColumnBranch:          {Title: i18n.Text(lang, string(ColumnBranch)), EnableSort: false},
				ColumnExecutor:        {Title: i18n.Text(lang, string(ColumnExecutor)), EnableSort: false},
				ColumnStartTime:       {Title: i18n.Text(lang, string(ColumnStartTime)), EnableSort: true, FieldBindToOrder: string(ColumnCostTime)},
			},
		},
	}
}

func (p *provider) RegisterTableChangePageOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		(*sdk.GlobalState)[StateKeyTransactionPaging] = opData.ClientData
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *provider) RegisterTableSortOp(opData table.OpTableChangeSort) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		(*sdk.GlobalState)[StateKeyTransactionSort] = opData.ClientData
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *provider) RegisterTablePagingOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return nil
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
	compName := "tabsTable"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "project-pipeline-exec-list",
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
	servicehub.Register("component-protocol.components.project-pipeline-exec-list.tabsTable", &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}

type Model struct {
	ProjectId uint64 `json:"projectId"`
}

type ServiceInParams struct {
	InParamsPtr *Model
}

func (b *ServiceInParams) CustomInParamsPtr() interface{} {
	if b.InParamsPtr == nil {
		b.InParamsPtr = &Model{}
	}
	return b.InParamsPtr
}

func (b *ServiceInParams) EncodeFromCustomInParams(customInParamsPtr interface{}, stdInParamsPtr *cptype.ExtraMap) {
	cputil.MustObjJSONTransfer(customInParamsPtr, stdInParamsPtr)
}

func (b *ServiceInParams) DecodeToCustomInParams(stdInParamsPtr *cptype.ExtraMap, customInParamsPtr interface{}) {
	cputil.MustObjJSONTransfer(stdInParamsPtr, customInParamsPtr)
}
