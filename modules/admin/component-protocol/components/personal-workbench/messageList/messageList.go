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

package messageList

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/personal-workbench/common"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/personal-workbench/common/gshelper"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/personal-workbench/i18n"
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/admin/services/workbench"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	CompMessageList = "messageList"

	DefaultPageNo   uint64 = 1
	DefaultPageSize uint64 = 10
)

type MessageList struct {
	base.DefaultProvider
	impl.DefaultList

	sdk   *cptype.SDK
	bdl   *bundle.Bundle
	wbSvc *workbench.Workbench

	identity  apistructs.Identity
	filterReq apistructs.WorkbenchMsgRequest
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, CompMessageList, func() servicehub.Provider {
		return &MessageList{}
	})
}

func (l *MessageList) Initialize(sdk *cptype.SDK) {}

func (l *MessageList) Finalize(sdk *cptype.SDK) {}

func (l *MessageList) BeforeHandleOp(sdk *cptype.SDK) {
	// get svc info
	l.sdk = sdk
	l.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	l.wbSvc = sdk.Ctx.Value(types.WorkbenchSvc).(*workbench.Workbench)

	// set component version
	l.sdk.Comp.Version = "2"

	// get identity info
	l.identity = apistructs.Identity{
		UserID: sdk.Identity.UserID,
		OrgID:  sdk.Identity.OrgID,
	}

	// get global related stat info
	gh := gshelper.NewGSHelper(sdk.GlobalState)
	tp, _ := gh.GetMsgTabName()

	// construct filter info, check & set default value
	l.filterReq = apistructs.WorkbenchMsgRequest{
		Type: tp,
		PageRequest: apistructs.PageRequest{
			PageNo:   DefaultPageNo,
			PageSize: DefaultPageSize,
		},
	}
}

func (l *MessageList) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		l.StdDataPtr = l.doFilter()
	}
}

func (l *MessageList) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return l.RegisterInitializeOp()
}

// RegisterChangePage when change page, filter needed
func (l *MessageList) RegisterChangePage(opData list.OpChangePage) (opFunc cptype.OperationFunc) {
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

// RegisterItemStarOp when item stared, unnecessary here
func (l *MessageList) RegisterItemStarOp(opData list.OpItemStar) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (l *MessageList) RegisterItemClickGotoOp(opData list.OpItemClickGoto) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

// RegisterItemClickOp get client data, and set message read
func (l *MessageList) RegisterItemClickOp(opData list.OpItemClick) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {

	}
}

func (l *MessageList) doFilter() (data *list.Data) {
	data = &list.Data{}

	switch l.filterReq.Type {
	case apistructs.WorkbenchItemUnreadMes:
		return l.doFilterMsg()
	default:
		logrus.Errorf("item [%v] not support", l.filterReq.Type)
	}
	return
}

func (l *MessageList) doFilterMsg() (data *list.Data) {
	data = &list.Data{}

	// list unread message
	req := apistructs.QueryMBoxRequest{PageNo: int64(l.filterReq.PageNo),
		PageSize: int64(l.filterReq.PageSize),
		Status:   apistructs.MBoxUnReadStatus}
	ms, err := l.bdl.ListMbox(l.identity, req)
	if err != nil {
		logrus.Errorf("list unread messages failed, identity: %+v,request: %+v, error:%v", l.identity, req, err)
		return
	}

	data = &list.Data{
		PageNo:   l.filterReq.PageNo,
		PageSize: l.filterReq.PageSize,
		Total:    uint64(ms.Total),
		Title:    l.sdk.I18n(i18n.I18nKeyUnreadMes),
		Operations: map[cptype.OperationKey]cptype.Operation{
			list.OpChangePage{}.OpKey(): cputil.NewOpBuilder().Build(),
		},
	}

	for _, p := range ms.List {
		item := list.Item{
			ID:           strconv.FormatInt(p.ID, 10),
			Title:        p.Title,
			TitleSummary: strconv.FormatInt(p.UnreadCount, 10),
			TitleState:   []list.StateInfo{{Status: common.UnreadMsgStatus}},
			// TODO columns info
			// ColumnsInfo:  columns,
			Operations: genClickGotoServerData(p),
		}
		data.List = append(data.List, item)
	}
	return
}

func genClickGotoServerData(i *apistructs.MBox) map[cptype.OperationKey]cptype.Operation {

	sd := list.OpItemBasicServerData{}

	// content check
	if strings.HasPrefix(i.Content, "http") {
		sd = list.OpItemBasicServerData{
			JumpOut: true,
			// url link
			Target: i.Content,
		}
	} else {
		// TODO what to do for non url content
		sd = list.OpItemBasicServerData{
			JumpOut: false,
			// non url like
			Target: i.Content,
		}
	}
	ops := map[cptype.OperationKey]cptype.Operation{
		list.OpItemClick{}.OpKey(): cputil.NewOpBuilder().Build(),
		list.OpItemClickGoto{}.OpKey(): cputil.NewOpBuilder().
			WithSkipRender(true).
			WithServerDataPtr(sd).Build(),
	}
	return ops
}
