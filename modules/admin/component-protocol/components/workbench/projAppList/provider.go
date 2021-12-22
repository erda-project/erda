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

package projAppList

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/workbench/common/convert"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/workbench/common/gshelper"
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/admin/services/workbench"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	CompProjAppList = "ProjAppList"

	DefaultPageNo   uint64 = 1
	DefaultPageSize uint64 = 10
)

type ProjAppList struct {
	base.DefaultProvider
	impl.DefaultList

	bdl   *bundle.Bundle
	wbSvc *workbench.Workbench

	identity  apistructs.Identity
	filterReq apistructs.WorkbenchProjAppRequest
}

func init() {
	base.InitProviderWithCreator(types.ScenarioWorkbench, CompProjAppList, func() servicehub.Provider {
		return &ProjAppList{}
	})
}

func (l *ProjAppList) Initialize(sdk *cptype.SDK) {}

func (l *ProjAppList) Finalize(sdk *cptype.SDK) {}

func (l *ProjAppList) BeforeHandleOp(sdk *cptype.SDK) {
	// get svc info
	l.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	l.wbSvc = sdk.Ctx.Value(types.WorkbenchSvc).(*workbench.Workbench)

	// get identity info
	l.identity = apistructs.Identity{
		UserID: sdk.Identity.UserID,
		OrgID:  sdk.Identity.OrgID,
	}

	// get global related stat info
	gh := gshelper.NewGSHelper(sdk.GlobalState)
	tp, _ := gh.GetWorkbenchItemType()
	query, _ := gh.GetFilterName()

	// construct filter info, check & set default value
	l.filterReq = apistructs.WorkbenchProjAppRequest{
		Type:  tp,
		Query: query,
		PageRequest: apistructs.PageRequest{
			PageNo:   DefaultPageNo,
			PageSize: DefaultPageSize,
		},
	}
	if l.filterReq.Type.IsEmpty() {
		l.filterReq.Type = apistructs.WorkbenchItemDefault
	}
}

func (l *ProjAppList) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		l.StdDataPtr = l.doFilter()
	}
}

func (l *ProjAppList) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return l.RegisterInitializeOp()
}

// RegisterListPagingOp when change page, filter needed
func (l *ProjAppList) RegisterListPagingOp(opData list.OpChangePage) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		if opData.ClientData.PageNo > 0 {
			l.filterReq.PageNo = opData.ClientData.PageNo
		}
		if opData.ClientData.PageSize > 0 {
			l.filterReq.PageSize = opData.ClientData.PageSize
		}
		l.StdDataPtr = l.doFilter()
	}
}

// RegisterItemStarOp when item stared, filter is unnecessary
func (l *ProjAppList) RegisterItemStarOp(opData list.OpItemStar) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		var (
			tp   apistructs.SubscribeType
			tpID uint64
		)

		if l.filterReq.Type == apistructs.WorkbenchItemProj {
			tp = apistructs.ProjectSubscribe
		} else {
			tp = apistructs.AppSubscribe
		}

		id, err := strconv.Atoi(opData.ClientData.DataRef.ID)
		if err != nil {
			panic(fmt.Errorf("star operation, format ClientData id failed, id: %v, error: %v", opData.ClientData.DataRef.ID, err))
		}
		tpID = uint64(id)

		req := apistructs.CreateSubscribeReq{
			Type:   tp,
			TypeID: tpID,
			Name:   opData.ClientData.DataRef.Title,
			UserID: l.identity.UserID,
		}

		_, err = l.bdl.CreateSubscribe(l.identity.UserID, l.identity.OrgID, req)
		if err != nil {
			panic(fmt.Errorf("star %v %v failed, id: %v, error: %v", req.Type, req.Name, req.TypeID, err))
		}
	}
}

func (l *ProjAppList) RegisterItemClickGotoOp(opData list.OpItemClickGoto) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {

	}
}

func (l *ProjAppList) doFilter() *list.Data {
	switch l.filterReq.Type {
	case apistructs.WorkbenchItemProj:
		return l.doFilterProj()
	case apistructs.WorkbenchItemApp:
		return l.doFilterApp()
	default:
		return l.doFilterProj()
	}
}

func (l *ProjAppList) doFilterProj() *list.Data {
	var data list.Data
	projs, err := l.wbSvc.ListQueryProjWbData(l.identity, l.filterReq.PageRequest, l.filterReq.Query)
	if err != nil {
		panic(fmt.Errorf("list query projct workbench data failed, error: %v", err))
	}

	data = list.Data{
		Total:    uint64(projs.Total),
		PageNo:   l.filterReq.PageNo,
		PageSize: l.filterReq.PageSize,
		Operations: map[cptype.OperationKey]cptype.Operation{
			list.OpChangePage{}.OpKey(): cputil.NewOpBuilder().Build(),
		},
	}

	for _, p := range projs.List {
		// TODO: construct list item
		item := list.Item{
			ID:    strconv.FormatUint(p.ProjectDTO.ID, 10),
			Title: p.ProjectDTO.Name,
			// TODO: url
			LogoURL: "",
			// TODO: operation
			Star: false,
			Labels: []list.ItemLabel{
				{
					Label: p.ProjectDTO.Type,
				},
			},
			Description: p.ProjectDTO.DisplayName,
			// TODO: url
			BackgroundImgURL: "",
			KvInfos:          convert.GenProjKvInfo(p),
			// TODO: operation
			Operations: map[cptype.OperationKey]cptype.Operation{},
			// TODO: operation
			MoreOperations: list.MoreOperations{
				Operations:      map[cptype.OperationKey]cptype.Operation{},
				OperationsOrder: []cptype.OperationKey{},
			},
		}
		data.List = append(data.List, item)
	}
	return &data
}

func (l *ProjAppList) doFilterApp() *list.Data {
	var data list.Data
	return &data
}
