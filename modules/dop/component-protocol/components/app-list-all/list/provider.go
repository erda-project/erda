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

package list

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/personal-workbench/common"
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/app-list-all/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	DefaultPageNo   = 1
	DefaultPageSize = 20
)

type List struct {
	base.DefaultProvider
	impl.DefaultList

	sdk       *cptype.SDK
	bdl       *bundle.Bundle
	identity  apistructs.Identity
	filterReq *apistructs.ApplicationListRequest
}

func init() {
	base.InitProviderWithCreator("app-list-all", "list", func() servicehub.Provider {
		return &List{}
	})
}

func (l *List) Initialize(sdk *cptype.SDK) {}

func (l *List) Finalize(sdk *cptype.SDK) {}

func (l *List) BeforeHandleOp(sdk *cptype.SDK) {
	l.sdk = sdk
	l.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	l.identity = apistructs.Identity{
		UserID: sdk.Identity.UserID,
		OrgID:  sdk.Identity.OrgID,
	}
	gh := gshelper.NewGSHelper(sdk.GlobalState)
	req, _ := gh.GetAppPagingRequest()
	req.PageNo = DefaultPageNo
	req.PageSize = DefaultPageSize
	req.ProjectID = l.StdInParamsPtr.Uint64("projectId")
	l.filterReq = req
}

func (l *List) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		l.StdDataPtr = l.doFilterApp()
	}
}

func (l *List) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return l.RegisterInitializeOp()
}

func (l *List) RegisterChangePage(opData list.OpChangePage) (opFunc cptype.OperationFunc) {
	if opData.ClientData.PageNo > 0 {
		l.filterReq.PageNo = int(opData.ClientData.PageNo)
	}
	if opData.ClientData.PageSize > 0 {
		l.filterReq.PageSize = int(opData.ClientData.PageSize)
	}
	l.StdDataPtr = l.doFilterApp()
	return nil
}

func (l *List) RegisterItemClickGotoOp(opData list.OpItemClickGoto) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (l *List) RegisterItemStarOp(opData list.OpItemStar) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (l *List) RegisterItemClickOp(opData list.OpItemClick) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (l *List) doFilterApp() (data *list.Data) {
	data = &list.Data{}
	apps, err := l.bdl.GetAppList(l.identity.OrgID, l.identity.UserID, *l.filterReq)
	if err != nil {
		logrus.Errorf("list query app workbench data failed, error: %v", err)
		panic(err)
	}

	data = &list.Data{
		Total:    uint64(apps.Total),
		PageNo:   uint64(l.filterReq.PageNo),
		PageSize: uint64(l.filterReq.PageSize),
		Operations: map[cptype.OperationKey]cptype.Operation{
			list.OpChangePage{}.OpKey(): cputil.NewOpBuilder().Build(),
		},
	}

	for _, p := range apps.List {
		item := list.Item{
			ID:      strconv.FormatUint(p.ID, 10),
			LogoURL: p.Logo,
			Title:   p.Name,
			KvInfos: l.GenAppKvInfo(p),
			Operations: map[cptype.OperationKey]cptype.Operation{
				list.OpItemClickGoto{}.OpKey(): cputil.NewOpBuilder().
					WithSkipRender(true).
					WithServerDataPtr(list.OpItemClickGotoServerData{
						OpItemBasicServerData: list.OpItemBasicServerData{
							Params: map[string]interface{}{
								common.OpKeyProjectID: p.ProjectID,
								common.OpKeyAppID:     p.ID,
							},
							Target: common.OpValTargetRepo,
						},
					}).
					Build(),
			},
		}
		data.List = append(data.List, item)
	}
	return
}
