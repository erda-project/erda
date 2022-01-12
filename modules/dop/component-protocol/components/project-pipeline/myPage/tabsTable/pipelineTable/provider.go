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
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
	"github.com/erda-project/erda/modules/msp/apm/service/common/transaction"
)

type PipelineTable struct {
	impl.DefaultTable

	bdl      *bundle.Bundle
	gsHelper *gshelper.GSHelper
	sdk      *cptype.SDK
	InParams InParams       `json:"-"`
	PageNo   uint64         `json:"-"`
	PageSize uint64         `json:"-"`
	Total    uint64         `json:"-"`
	Sorts    []*common.Sort `json:"-"`

	ProjectPipelineSvc *projectpipeline.ProjectPipelineService
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

func (p *PipelineTable) BeforeHandleOp(sdk *cptype.SDK) {
	p.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.gsHelper = gshelper.NewGSHelper(sdk.GlobalState)
	p.sdk = sdk
	if err := p.setInParams(); err != nil {
		panic(err)
	}
	p.ProjectPipelineSvc = sdk.Ctx.Value(types.ProjectPipelineService).(*projectpipeline.ProjectPipelineService)
	//cputil.MustObjJSONTransfer(&p.StdStatePtr, &p.State)
}

func (p *PipelineTable) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.SetPagingFromGlobalState()
		p.SetSortsFromGlobalState()
		p.StdDataPtr = &table.Data{
			Table: table.Table{
				Columns:  p.SetTableColumns(),
				Rows:     p.SetTableRows(),
				PageNo:   p.PageNo,
				PageSize: p.PageSize,
				Total:    p.Total,
			},
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpTableChangePage{}.OpKey(): cputil.NewOpBuilder().WithServerDataPtr(&table.OpTableChangePageServerData{}).Build(),
				table.OpTableChangeSort{}.OpKey(): cputil.NewOpBuilder().Build(),
				table.OpBatchRowsHandle{}.OpKey(): cputil.NewOpBuilder().WithText("批量操作").WithServerDataPtr(&table.OpBatchRowsHandleServerData{
					Options: []table.OpBatchRowsHandleOption{
						{
							ID:   "run",
							Text: "执行",
							//AllowedRowIDs: []string{"row1", "row2"},
						},
					},
				}).Build(),
			},
		}
	}
}

func (p *PipelineTable) SetTableColumns() table.ColumnsInfo {
	return table.ColumnsInfo{
		Orders: []table.ColumnKey{ColumnPipelineName, ColumnPipelineStatus, ColumnCostTime, ColumnApplicationName, ColumnBranch, ColumnExecutor, ColumnStartTime},
		ColumnsMap: map[table.ColumnKey]table.Column{
			ColumnPipelineName:    {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineName))},
			ColumnPipelineStatus:  {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineStatus))},
			ColumnCostTime:        {Title: cputil.I18n(p.sdk.Ctx, string(ColumnCostTime)), EnableSort: true},
			ColumnApplicationName: {Title: cputil.I18n(p.sdk.Ctx, string(ColumnApplicationName))},
			ColumnBranch:          {Title: cputil.I18n(p.sdk.Ctx, string(ColumnBranch))},
			ColumnExecutor:        {Title: cputil.I18n(p.sdk.Ctx, string(ColumnExecutor))},
			ColumnStartTime:       {Title: cputil.I18n(p.sdk.Ctx, string(ColumnStartTime)), EnableSort: true},
		},
	}
}

func (p *PipelineTable) SetTableRows() []table.Row {
	var descCols, ascCols []string
	for _, v := range p.Sorts {
		field := func() string {
			if v.FieldKey == string(ColumnCostTime) {
				return "cost_time"
			}
			if v.FieldKey == string(ColumnStartTime) {
				return "started_at"
			}
			return ""
		}()
		if field == "" {
			continue
		}
		if v.Ascending {
			ascCols = append(ascCols, field)
		} else {
			descCols = append(descCols, field)
		}
	}
	if len(ascCols) == 0 && len(descCols) == 0 {
		descCols = append(descCols, "started_at")
	}

	filter := p.gsHelper.GetGlobalTableFilter()
	list, total, err := p.ProjectPipelineSvc.List(p.sdk.Ctx, deftype.ProjectPipelineList{
		ProjectID: p.InParams.ProjectID,
		AppName:   filter.App,
		Creator: func() []string {
			if p.gsHelper.GetGlobalPipelineTab() == "mine" {
				return []string{p.sdk.Identity.UserID}
			}
			return filter.Creator
		}(),
		Executor: filter.Executor,
		PageNo:   p.PageNo,
		PageSize: p.PageSize,
		Category: func() []string {
			if p.gsHelper.GetGlobalPipelineTab() == "primary" {
				return []string{"primary"}
			}
			return nil
		}(),
		Name: p.gsHelper.GetGlobalNameInputFilter(),
		TimeCreated: func() []string {
			timeCreated := make([]string, 0)
			if len(filter.CreatedAtStartEnd) == 2 {
				timeCreated = append(timeCreated, time.Unix(filter.CreatedAtStartEnd[0]/1000, 0).String())
				timeCreated = append(timeCreated, time.Unix(filter.CreatedAtStartEnd[1]/1000, 0).String())
			}
			return timeCreated
		}(),
		TimeStarted: func() []string {
			timeStarted := make([]string, 0)
			if len(filter.StartedAtStartEnd) == 2 {
				timeStarted = append(timeStarted, time.Unix(filter.StartedAtStartEnd[0]/1000, 0).String())
				timeStarted = append(timeStarted, time.Unix(filter.StartedAtStartEnd[1]/1000, 0).String())
			}
			return timeStarted
		}(),
		Status:       filter.Status,
		DescCols:     descCols,
		AscCols:      ascCols,
		IdentityInfo: apistructs.IdentityInfo{UserID: p.sdk.Identity.UserID},
	})
	if err != nil {
		logrus.Errorf("failed to list project pipeline, err: %s", err.Error())
		//return nil
	}

	p.Total = uint64(total)
	rows := make([]table.Row, 0, len(list))
	for _, v := range list {
		rows = append(rows, table.Row{
			ID:         table.RowID(v.ID),
			Selectable: true,
			Selected:   false,
			CellsMap: map[table.ColumnKey]table.Cell{
				ColumnPipelineName:    table.NewTextCell(v.Name).Build(),
				ColumnPipelineStatus:  table.NewTextCell(cputil.I18n(p.sdk.Ctx, string(ColumnPipelineStatus)) + "success").Build(),
				ColumnCostTime:        table.NewTextCell(fmt.Sprintf("%v s", v.CostTime)).Build(),
				ColumnApplicationName: table.NewTextCell(getApplicationNameFromDefinitionRemote(v.Remote)).Build(),
				ColumnBranch:          table.NewTextCell(v.Ref).Build(),
				ColumnExecutor:        table.NewUserCell(commodel.User{ID: v.Creator}).Build(),
				ColumnStartTime:       table.NewTextCell(v.StartedAt.AsTime().Format("2006-01-02 15:04:05")).Build(),
			},
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpRowSelect{}.OpKey(): cputil.NewOpBuilder().Build(),
				table.OpRowAdd{}.OpKey():    cputil.NewOpBuilder().Build(),
				table.OpRowEdit{}.OpKey():   cputil.NewOpBuilder().Build(),
				table.OpRowDelete{}.OpKey(): cputil.NewOpBuilder().Build(),
			},
		})
	}
	return rows
}

func getApplicationNameFromDefinitionRemote(remote string) string {
	values := strings.Split(remote, string(filepath.Separator))
	if len(values) != 3 {
		return remote
	}
	return values[2]
}

func (p *PipelineTable) SetSortsFromGlobalState() {
	globalState := *p.sdk.GlobalState
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
	p.Sorts = sorts
}

func (p *PipelineTable) SetPagingFromGlobalState() {
	globalState := *p.sdk.GlobalState
	pageNo, pageSize := 1, common.DefaultPageSize
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
	p.PageNo = uint64(pageNo)
	p.PageSize = uint64(pageSize)
}

func (p *PipelineTable) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *PipelineTable) RegisterTablePagingOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		(*sdk.GlobalState)[transaction.StateKeyTransactionPaging] = opData.ClientData
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *PipelineTable) RegisterTableChangePageOp(opData table.OpTableChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		(*sdk.GlobalState)[transaction.StateKeyTransactionPaging] = opData.ClientData
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *PipelineTable) RegisterTableSortOp(opData table.OpTableChangeSort) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		(*sdk.GlobalState)[transaction.StateKeyTransactionSort] = opData.ClientData
		p.RegisterInitializeOp()(sdk)
	}
}

func (p *PipelineTable) RegisterBatchRowsHandleOp(opData table.OpBatchRowsHandle) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *PipelineTable) RegisterRowSelectOp(opData table.OpRowSelect) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *PipelineTable) RegisterRowAddOp(opData table.OpRowAdd) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *PipelineTable) RegisterRowEditOp(opData table.OpRowEdit) (opFunc cptype.OperationFunc) {
	return nil
}

func (p *PipelineTable) RegisterRowDeleteOp(opData table.OpRowDelete) (opFunc cptype.OperationFunc) {
	return nil
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "pipelineTable", func() servicehub.Provider {
		return &PipelineTable{}
	})
}
