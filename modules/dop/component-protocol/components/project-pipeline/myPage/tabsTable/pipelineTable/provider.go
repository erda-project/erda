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
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
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
	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
	"github.com/erda-project/erda/modules/msp/apm/service/common/transaction"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	setPrimary cptype.OperationKey = "setPrimary"
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
	UserIDs  []string       `json:"-"`

	ProjectPipelineSvc *projectpipeline.ProjectPipelineService
}

const (
	ColumnPipelineName         table.ColumnKey = "pipelineName"
	ColumnPipelineLatestStatus table.ColumnKey = "pipelineLatestStatus"
	ColumnLatestCostTime       table.ColumnKey = "latestCostTime"
	ColumnApplicationName      table.ColumnKey = "applicationName"
	ColumnBranch               table.ColumnKey = "branch"
	ColumnLatestExecutor       table.ColumnKey = "latestExecutor"
	ColumnLatestStartTime      table.ColumnKey = "latestStartTime"
	ColumnMoreOperations       table.ColumnKey = "moreOperations"

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
							ID:              "batchRun",
							Text:            "执行",
							ForbiddenRowIDs: []string{},
						},
					},
				}).Build(),
			},
		}
		(*sdk.GlobalState)[protocol.GlobalInnerKeyUserIDs.String()] = strutil.DedupSlice(p.UserIDs, true)
	}
}

func (p *PipelineTable) SetTableColumns() table.ColumnsInfo {
	return table.ColumnsInfo{
		Orders: []table.ColumnKey{ColumnPipelineName, ColumnPipelineLatestStatus, ColumnLatestCostTime, ColumnApplicationName, ColumnBranch, ColumnLatestExecutor, ColumnLatestStartTime, ColumnMoreOperations},
		ColumnsMap: map[table.ColumnKey]table.Column{
			ColumnPipelineName:         {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineName))},
			ColumnPipelineLatestStatus: {Title: cputil.I18n(p.sdk.Ctx, string(ColumnPipelineLatestStatus))},
			ColumnLatestCostTime:       {Title: cputil.I18n(p.sdk.Ctx, string(ColumnLatestCostTime)), EnableSort: true},
			ColumnApplicationName:      {Title: cputil.I18n(p.sdk.Ctx, string(ColumnApplicationName))},
			ColumnBranch:               {Title: cputil.I18n(p.sdk.Ctx, string(ColumnBranch))},
			ColumnLatestExecutor:       {Title: cputil.I18n(p.sdk.Ctx, string(ColumnLatestExecutor))},
			ColumnLatestStartTime:      {Title: cputil.I18n(p.sdk.Ctx, string(ColumnLatestStartTime)), EnableSort: true},
			ColumnMoreOperations:       {Title: cputil.I18n(p.sdk.Ctx, string(ColumnMoreOperations))},
		},
	}
}

func (p *PipelineTable) SetTableRows() []table.Row {
	var descCols, ascCols []string
	for _, v := range p.Sorts {
		field := func() string {
			if v.FieldKey == string(ColumnLatestCostTime) {
				return "cost_time"
			}
			if v.FieldKey == string(ColumnLatestStartTime) {
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
	var inParamsAppName string
	if p.InParams.AppID != 0 {
		app, err := p.bdl.GetApp(p.InParams.AppID)
		if err != nil {
			logrus.Errorf("failed to GetApp, err: %s", err.Error())
		}
		inParamsAppName = app.Name
	}

	filter := p.gsHelper.GetGlobalTableFilter()
	list, total, err := p.ProjectPipelineSvc.List(p.sdk.Ctx, deftype.ProjectPipelineList{
		ProjectID: p.InParams.ProjectID,
		AppName: func() []string {
			if inParamsAppName != "" {
				return []string{inParamsAppName}
			}
			return filter.App
		}(),
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

	var (
		pipelineYmlNames []string
		pipelineSources  []apistructs.PipelineSource
	)

	definitionYmlSourceMap := make(map[string]string)
	for _, v := range list {
		var extraValue = apistructs.PipelineDefinitionExtraValue{}
		err = json.Unmarshal([]byte(v.Extra.Extra), &extraValue)
		if err != nil {
			logrus.Errorf("failed to list Unmarshal Extra, err: %s", err.Error())
		}
		pipelineYmlNames = append(pipelineYmlNames, extraValue.CreateRequest.PipelineYmlName)
		pipelineSources = append(pipelineSources, extraValue.CreateRequest.PipelineSource)
		definitionYmlSourceMap[v.ID] = fmt.Sprintf("%s%s", extraValue.CreateRequest.PipelineYmlName, extraValue.CreateRequest.PipelineSource)
	}
	crons, err := p.bdl.PageListPipelineCrons(apistructs.PipelineCronPagingRequest{
		Sources:  pipelineSources,
		YmlNames: pipelineYmlNames,
		PageSize: 1,
		PageNo:   999,
	})
	if err != nil {
		logrus.Errorf("failed to list PageListPipelineCrons, err: %s", err.Error())
	}
	ymlSourceMapCronMap := make(map[string]*apistructs.PipelineCronDTO)
	if crons != nil {
		for _, v := range crons.Data {
			ymlSourceMapCronMap[fmt.Sprintf("%s%s", v.PipelineYmlName, v.PipelineSource)] = v
		}
	}

	p.Total = uint64(total)
	rows := make([]table.Row, 0, len(list))
	for _, v := range list {
		p.UserIDs = append(p.UserIDs, v.Creator, v.Executor)
		rows = append(rows, table.Row{
			ID:         table.RowID(strconv.FormatInt(v.PipelineId, 10)),
			Selectable: true,
			Selected:   false,
			CellsMap: map[table.ColumnKey]table.Cell{
				ColumnPipelineName: table.NewTextCell(v.Name).Build(),
				ColumnPipelineLatestStatus: table.NewCompleteTextCell(commodel.Text{
					Text: func() string {
						if v.Status == "" {
							return "-"
						}
						return cputil.I18n(p.sdk.Ctx, "pipelineStatus"+v.Status)
					}(),
					Status: func() commodel.UnifiedStatus {
						if apistructs.PipelineStatus(v.Status).IsRunningStatus() {
							return commodel.ProcessingStatus
						}
						if apistructs.PipelineStatus(v.Status).IsFailedStatus() {
							return commodel.ErrorStatus
						}
						return commodel.DefaultStatus
					}(),
				}).Build(),
				ColumnLatestCostTime: table.NewTextCell(func() string {
					if v.CostTime <= 0 {
						return "-"
					}
					return fmt.Sprintf("%v s", v.CostTime)
				}()).Build(),
				ColumnApplicationName: table.NewTextCell(getApplicationNameFromDefinitionRemote(v.Remote)).Build(),
				ColumnBranch:          table.NewTextCell(v.Ref).Build(),
				ColumnLatestExecutor:  table.NewUserCell(commodel.User{ID: v.Creator}).Build(),
				ColumnLatestStartTime: table.NewTextCell(func() string {
					v.StartedAt.AsTime().Format("2006-01-02 15:04:05")
					if v.StartedAt.AsTime().Unix() <= 0 {
						return "-"
					}
					return v.StartedAt.AsTime().Format("2006-01-02 15:04:05")
				}()).Build(),
				ColumnMoreOperations: table.NewMoreOperationsCell(commodel.MoreOperations{
					Ops: p.SetTableMoreOpItem(v, definitionYmlSourceMap, ymlSourceMapCronMap),
				}).Build(),
			},
			Operations: map[cptype.OperationKey]cptype.Operation{
				table.OpRowSelect{}.OpKey(): func() cptype.Operation {
					build := cputil.NewOpBuilder().Build()
					serviceCnt := make(cptype.OpServerData)
					serviceCnt["id"] = v.ID
					build.ServerData = &serviceCnt
					return build
				}(),
			},
		})
	}
	return rows
}

func (p *PipelineTable) SetTableMoreOpItem(definition *pb.PipelineDefinition, definitionYmlSourceMap map[string]string, ymlSourceMapCronMap map[string]*apistructs.PipelineCronDTO) []commodel.MoreOpItem {
	items := make([]commodel.MoreOpItem, 0)
	build := cputil.NewOpBuilder().Build()
	serviceCnt := make(cptype.OpServerData)
	serviceCnt["id"] = definition.ID
	build.ServerData = &serviceCnt
	items = append(items, commodel.MoreOpItem{
		ID: func() string {
			if definition.Category == "primary" {
				return "unsetPrimary"
			}
			return "setPrimary"
		}(),
		Text: cputil.I18n(p.sdk.Ctx, func() string {
			if definition.Category == "primary" {
				return "unsetPrimary"
			}
			return "setPrimary"
		}()),
		Operations: map[cptype.OperationKey]cptype.Operation{
			commodel.OpMoreOperationsItemClick{}.OpKey(): build,
		},
	})
	items = append(items, commodel.MoreOpItem{
		ID: func() string {
			if apistructs.PipelineStatus(definition.Status).IsRunningStatus() {
				return "cancelRun"
			}
			return "run"
		}(),
		Text: cputil.I18n(p.sdk.Ctx, func() string {
			if apistructs.PipelineStatus(definition.Status).IsRunningStatus() {
				return "cancelRun"
			}
			return "run"
		}()),
		Operations: map[cptype.OperationKey]cptype.Operation{
			commodel.OpMoreOperationsItemClick{}.OpKey(): build,
		},
	})
	if apistructs.PipelineStatus(definition.Status).IsFailedStatus() {
		items = append(items, commodel.MoreOpItem{
			ID:   "rerunFromFail",
			Text: cputil.I18n(p.sdk.Ctx, "rerunFromFail"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				commodel.OpMoreOperationsItemClick{}.OpKey(): build,
			},
		})
	}
	if apistructs.PipelineStatus(definition.Status).IsEndStatus() {
		items = append(items, commodel.MoreOpItem{
			ID:   "rerun",
			Text: cputil.I18n(p.sdk.Ctx, "rerun"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				commodel.OpMoreOperationsItemClick{}.OpKey(): build,
			},
		})
	}
	if v, ok := ymlSourceMapCronMap[definitionYmlSourceMap[definition.ID]]; ok {
		items = append(items, commodel.MoreOpItem{
			ID: func() string {
				if *v.Enable {
					return "cancelCron"
				}
				return "cron"
			}(),
			Text: cputil.I18n(p.sdk.Ctx, func() string {
				if *v.Enable {
					return "cancelCron"
				}
				return "cron"
			}()),
			Operations: map[cptype.OperationKey]cptype.Operation{
				commodel.OpMoreOperationsItemClick{}.OpKey(): build,
			},
		})
	}

	items = append(items, commodel.MoreOpItem{
		ID:   "delete",
		Text: cputil.I18n(p.sdk.Ctx, "delete"),
		Operations: map[cptype.OperationKey]cptype.Operation{
			commodel.OpMoreOperationsItemClick{}.OpKey(): build,
		},
	})
	return items
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
	return func(sdk *cptype.SDK) {
		switch opData.ClientData.DataRef.ID {
		case "batchRun":
			_, err := p.ProjectPipelineSvc.BatchRun(p.sdk.Ctx, deftype.ProjectPipelineBatchRun{
				PipelineDefinitionIDs: opData.ClientData.SelectedRowIDs,
				ProjectID:             p.InParams.ProjectID,
				IdentityInfo:          apistructs.IdentityInfo{UserID: p.sdk.Identity.UserID},
			})
			if err != nil {
				panic(err)
			}
		}
		p.RegisterInitializeOp()(sdk)
	}
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

func (p *PipelineTable) RegisterMoreOperationOp(opData OpMoreOperationsItemClick) {
	id := opData.ServerData.ID
	switch opData.ClientData.DataRef.ID {
	case "setPrimary":
		_, err := p.ProjectPipelineSvc.SetPrimary(p.sdk.Ctx, deftype.ProjectPipelineCategory{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "unsetPrimary":
		_, err := p.ProjectPipelineSvc.UnSetPrimary(p.sdk.Ctx, deftype.ProjectPipelineCategory{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "run":
		_, err := p.ProjectPipelineSvc.Run(p.sdk.Ctx, deftype.ProjectPipelineRun{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "cancelRun":
		_, err := p.ProjectPipelineSvc.Cancel(p.sdk.Ctx, deftype.ProjectPipelineCancel{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "rerun":
		_, err := p.ProjectPipelineSvc.Rerun(p.sdk.Ctx, deftype.ProjectPipelineRerun{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "rerunFromFail":
		_, err := p.ProjectPipelineSvc.FailRerun(p.sdk.Ctx, deftype.ProjectPipelineFailRerun{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "corn":
		_, err := p.ProjectPipelineSvc.StartCron(p.sdk.Ctx, deftype.ProjectPipelineStartCron{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "cancelCron":
		_, err := p.ProjectPipelineSvc.EndCron(p.sdk.Ctx, deftype.ProjectPipelineEndCron{
			PipelineDefinitionID: id,
			ProjectID:            p.InParams.ProjectID,
			IdentityInfo:         apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	case "delete":
		_, err := p.ProjectPipelineSvc.Delete(p.sdk.Ctx, deftype.ProjectPipelineDelete{
			ID:           id,
			ProjectID:    p.InParams.ProjectID,
			IdentityInfo: apistructs.IdentityInfo{UserID: cputil.GetUserID(p.sdk.Ctx)},
		})
		if err != nil {
			panic(err)
		}
	}
	p.RegisterInitializeOp()(p.sdk)
}

func (p *PipelineTable) RegisterCompNonStdOps() (opFuncs map[cptype.OperationKey]cptype.OperationFunc) {
	return map[cptype.OperationKey]cptype.OperationFunc{
		commodel.OpMoreOperationsItemClick{}.OpKey(): func(sdk *cptype.SDK) {
			p.RegisterMoreOperationOp(*cputil.MustObjJSONTransfer(&sdk.Event.OperationData, &OpMoreOperationsItemClick{}).(*OpMoreOperationsItemClick))
		},
	}
}

type OpMoreOperationsItemClick struct {
	commodel.OpMoreOperationsItemClick
	ClientData OpMoreOperationsItemClickClientData `json:"clientData"`
	ServerData OpMoreOperationsItemClickServerData `json:"serverData,omitempty"`
}

type OpMoreOperationsItemClickClientData struct {
	DataRef       *commodel.MoreOpItem `json:"dataRef,omitempty"`
	ParentDataRef table.Row            `json:"parentDataRef,omitempty"`
}

type OpMoreOperationsItemClickServerData struct {
	ID string `json:"id"`
}
