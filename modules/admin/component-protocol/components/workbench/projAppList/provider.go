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
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/admin/services/workbench"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	CompProjAppList = "ProjAppList"
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
	l.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	l.wbSvc = sdk.Ctx.Value(types.WorkbenchSvc).(*workbench.Workbench)
	l.identity = apistructs.Identity{
		UserID: sdk.Identity.UserID,
		OrgID:  sdk.Identity.OrgID,
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

func (l *ProjAppList) RegisterListPagingOp(opData list.OpChangePage) (opFunc cptype.OperationFunc) {
	// TODO:
	return func(sdk *cptype.SDK) {}
}

func (l *ProjAppList) RegisterItemStarOp(opData list.OpItemStar) (opFunc cptype.OperationFunc) {
	// TODO:
	return func(sdk *cptype.SDK) {}
}

func (l *ProjAppList) RegisterItemClickGotoOp(opData list.OpItemClickGoto) (opFunc cptype.OperationFunc) {
	// TODO:
	return func(sdk *cptype.SDK) {}
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
		PageNo:   uint64(l.filterReq.PageNo),
		PageSize: uint64(l.filterReq.PageSize),
		Operations: map[cptype.OperationKey]cptype.Operation{
			list.OpChangePage{}.OpKey(): cputil.NewOpBuilder().Build(),
		},
	}

	for _, p := range projs.List {
		// TODO: construct list item
		item := list.Item{
			ID:               strconv.FormatUint(p.ProjectDTO.ID, 10),
			Title:            "",
			LogoURL:          "",
			Star:             false,
			Labels:           []list.ItemLabel{},
			Description:      "",
			BackgroundImgURL: "",
			KvInfos:          []list.KvInfo{},
			Operations:       map[cptype.OperationKey]cptype.Operation{},
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
