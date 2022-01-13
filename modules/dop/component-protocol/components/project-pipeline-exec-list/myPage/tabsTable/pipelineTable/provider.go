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

package pipelineTable

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/commodel"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline-exec-list/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline-exec-list/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
)

type provider struct {
	impl.DefaultTable
	sdk      *cptype.SDK
	bdl      *bundle.Bundle
	gsHelper *gshelper.GSHelper
	InParams InParams `json:"-"`

	ProjectPipeline projectpipeline.Service `autowired:"erda.dop.projectpipeline.ProjectPipelineServiceMethod" required:"true"`
}

const (
	ColumnPipelineName    table.ColumnKey = "pipelineName"
	ColumnPipelineStatus  table.ColumnKey = common.ColumnPipelineStatus
	ColumnCostTime        table.ColumnKey = "costTime"
	ColumnApplicationName table.ColumnKey = "applicationName"
	ColumnBranch          table.ColumnKey = "branch"
	ColumnExecutor        table.ColumnKey = "executor"
	ColumnStartTime       table.ColumnKey = "startTime"

	ColumnCostTimeOrder  = "cost_time_sec"
	ColumnStartTimeOrder = "time_begin"

	StateKeyTransactionPaging = "paging"
	StateKeyTransactionSort   = "sort"
)

func (p *provider) BeforeHandleOp(sdk *cptype.SDK) {
	p.sdk = sdk
	if err := p.setInParams(); err != nil {
		panic(err)
	}
	p.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.gsHelper = gshelper.NewGSHelper(sdk.GlobalState)
	p.ProjectPipeline = sdk.Ctx.Value(types.ProjectPipelineService).(*projectpipeline.ProjectPipelineService)
	//cputil.MustObjJSONTransfer(&p.StdStatePtr, &p.State)
}

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.sdk = sdk
		projectID := p.InParams.ProjectID
		pageNo, pageSize := GetPagingFromGlobalState(*sdk.GlobalState)
		sorts := GetSortsFromGlobalState(*sdk.GlobalState)

		var descCols []string
		var ascCols []string
		for _, v := range sorts {
			if v.FieldKey != ColumnCostTimeOrder && v.FieldKey != ColumnStartTimeOrder {
				continue
			}
			if v.Ascending {
				ascCols = append(ascCols, v.FieldKey)
			} else {
				descCols = append(descCols, v.FieldKey)
			}
		}

		var req = deftype.ProjectPipelineListExecHistory{
			DescCols:  descCols,
			AscCols:   ascCols,
			PageNo:    uint64(pageNo),
			PageSize:  uint64(pageSize),
			ProjectID: projectID,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
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

		tableValue := p.InitTable()
		tableValue.PageSize = uint64(pageSize)
		tableValue.PageNo = uint64(pageNo)

		result, err := p.ProjectPipeline.ListExecHistory(context.Background(), req)
		if err != nil {
			logrus.Error("failed to get table data: %s", err)
			p.StdDataPtr = &table.Data{
				Table: tableValue,
				Operations: map[cptype.OperationKey]cptype.Operation{
					table.OpTableChangePage{}.OpKey(): cputil.NewOpBuilder().WithServerDataPtr(&table.OpTableChangePageServerData{}).Build(),
					table.OpTableChangeSort{}.OpKey(): cputil.NewOpBuilder().Build(),
				}}
			return
		}

		tableValue.Total = uint64(result.Data.Total)
		for _, pipeline := range result.Data.Pipelines {
			if pipeline.DefinitionInfo == nil {
				continue
			}
			tableValue.Rows = append(tableValue.Rows, p.pipelineToRow(pipeline))
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

func (p *provider) pipelineToRow(pipeline apistructs.PagePipeline) table.Row {
	return table.Row{
		ID:         table.RowID(fmt.Sprintf("%v", pipeline.ID)),
		Selectable: true,
		Selected:   false,
		CellsMap: map[table.ColumnKey]table.Cell{
			ColumnPipelineName:   table.NewTextCell(pipeline.DefinitionInfo.Name).Build(),
			ColumnPipelineStatus: table.NewTextCell(cputil.I18n(p.sdk.Ctx, string(ColumnPipelineStatus)+pipeline.Status.String())).Build(),
			ColumnCostTimeOrder: table.NewTextCell(func() string {
				if pipeline.CostTimeSec <= 0 {
					return "-"
				} else {
					return fmt.Sprintf("%v s", pipeline.CostTimeSec)
				}
			}()).Build(),
			ColumnApplicationName: table.NewTextCell(getApplicationNameFromDefinitionRemote(pipeline.DefinitionInfo.SourceRemote)).Build(),
			ColumnBranch:          table.NewTextCell(pipeline.DefinitionInfo.SourceRef).Build(),
			ColumnExecutor:        table.NewUserCell(commodel.User{ID: pipeline.DefinitionInfo.Creator}).Build(),
			ColumnStartTimeOrder: table.NewTextCell(func() string {
				if pipeline.TimeBegin.Year() < 2000 {
					return "-"
				}
				return pipeline.TimeBegin.Format("2006-01-02 15:04:05")
			}()).Build(),
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

func (p *provider) InitTable() table.Table {
	return table.Table{
		Columns: table.ColumnsInfo{
			Orders: []table.ColumnKey{ColumnPipelineName, ColumnPipelineStatus, ColumnCostTimeOrder, ColumnApplicationName, ColumnBranch, ColumnExecutor, ColumnStartTimeOrder},
			ColumnsMap: map[table.ColumnKey]table.Column{
				ColumnPipelineName:    {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineName)), EnableSort: false},
				ColumnPipelineStatus:  {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineStatus)), EnableSort: false},
				ColumnCostTimeOrder:   {Title: cputil.I18n(p.sdk.Ctx, string(ColumnCostTime)), EnableSort: true, FieldBindToOrder: ColumnCostTimeOrder},
				ColumnApplicationName: {Title: cputil.I18n(p.sdk.Ctx, string(ColumnApplicationName)), EnableSort: false},
				ColumnBranch:          {Title: cputil.I18n(p.sdk.Ctx, string(ColumnBranch)), EnableSort: false},
				ColumnExecutor:        {Title: cputil.I18n(p.sdk.Ctx, string(ColumnExecutor)), EnableSort: false},
				ColumnStartTimeOrder:  {Title: cputil.I18n(p.sdk.Ctx, string(ColumnStartTime)), EnableSort: true, FieldBindToOrder: ColumnStartTimeOrder},
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

type InParams struct {
	OrgID     uint64 `json:"orgID,omitempty"`
	ProjectID uint64 `json:"projectId,omitempty"`
}

func (p *provider) setInParams() error {
	b, err := json.Marshal(p.InParamsPtr())
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &p.InParams); err != nil {
		return err
	}

	p.InParams.OrgID, err = strconv.ParseUint(p.sdk.Identity.OrgID, 10, 64)
	if err != nil {
		return err
	}
	return nil
}

// InParamsPtr .
func (s *provider) InParamsPtr() interface{} { return s.StdInParamsPtr }

func init() {
	base.InitProviderWithCreator("project-pipeline-exec-list", "pipelineTable", func() servicehub.Provider {
		return &provider{}
	})
}
