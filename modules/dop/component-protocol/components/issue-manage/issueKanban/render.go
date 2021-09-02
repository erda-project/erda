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

package issueKanban

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const defaultPageSize = 50

type ChangePageNoOperationData struct {
	FillMeta string `json:"fill_meta"`
	Meta     struct {
		PageData struct {
			PageNo   uint64 `json:"pageNo"`
			PageSize uint64 `json:"pageSize"`
		} `json:"pageData"`
		KanbanKey interface{} `json:"kanbanKey"`
	} `json:"meta"`
}

func GetCartOpsInfo(opsData interface{}, isDrag bool) (*OpMetaInfo, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op OperationInfo
	var dragOp DragOperationInfo
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	if !isDrag {
		err = json.Unmarshal(cont, &op)
	} else {
		err = json.Unmarshal(cont, &dragOp)
	}
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	if !isDrag {
		meta := op.Meta
		return &meta, nil
	}
	meta := dragOp.Meta
	return &meta, nil
}

func (i ComponentIssueBoard) GetFilterReq() (*IssueFilterRequest, error) {
	var inParams issueRenderInparams
	var req IssueFilterRequest
	cont, err := json.Marshal(i.sdk.InParams)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", i.sdk.InParams, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &inParams)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	if i.sdk.Identity.UserID != "" {
		req.UserID = i.sdk.Identity.UserID
	}
	req.IssuePagingRequest = i.State.FilterConditions
	_, ok := i.State.IssueViewGroupChildrenValue["kanban"]
	if ok {
		value := i.State.IssueViewGroupChildrenValue["kanban"]
		req.BoardType = BoardType(value)
	} else {
		req.BoardType = BoardTypeTime
	}
	req.PageSize = defaultPageSize
	req.PageNo = 1
	return &req, nil
}

func (i ComponentIssueBoard) GetDefaultFilterReq(req *IssueFilterRequest) error {
	var inParams issueRenderInparams
	cont, err := json.Marshal(i.sdk.InParams)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", i.sdk.InParams, err)
		return err
	}
	err = json.Unmarshal(cont, &inParams)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return err
	}
	req.IterationID, err = strconv.ParseInt(inParams.FixedIssueIteration, 10, 64)
	if inParams.FixedIssueType == "ALL" {
		req.Type = []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug, apistructs.IssueTypeEpic}
	} else {
		req.Type = append(req.Type, inParams.FixedIssueType)
	}
	if i.State.FilterConditions.IterationID != 0 {
		req.IterationID = i.State.FilterConditions.IterationID
	}
	if i.State.FilterConditions.Type != nil && len(i.State.FilterConditions.Type) != 0 {
		req.Type = i.State.FilterConditions.Type
	}
	return nil
}

func (i *ComponentIssueBoard) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state IssueBoardState
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	return nil
}

// Issue过滤，分类
func (i *ComponentIssueBoard) RenderOnFilter(ctx context.Context, req IssueFilterRequest) error {
	ib, err := i.Filter(ctx, req)
	if err != nil {
		logrus.Errorf("issue filter failed, request:%+v, err:%v", req, err)
		return err
	}
	i.Data = *ib
	return nil
}

func (i *ComponentIssueBoard) RenderOnMoveOut(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, false)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}

	// get
	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}
	// update
	is.IterationID = -1
	if err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID)); err != nil {
		return err
	}

	// refresh
	//err = i.RefreshOnMoveOut(is.ID)
	//if err != nil {
	//	logrus.Errorf("refresh on move out failed, issueID:%d", req.IssueID)
	//}
	return nil
}

func (i *ComponentIssueBoard) RefreshOnMoveOut(issueID int64) error {
	for k, v := range i.Data.Board {
		for _, is := range v.List {
			if is.ID == issueID {
				i.Data.Board[k].Delete(issueID)
				i.Data.Board[k].Total = i.Data.Board[k].Total - 1
				break
			}
		}
	}
	return nil
}

func (i *ComponentIssueBoard) RenderOnDrag(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, true)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}

	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}

	switch i.boardType {
	case BoardTypeStatus:
		currentState := is.State
		is.State = int64(i.State.DropTarget.(float64))
		if currentState == is.State {
			return nil
		}
		err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		//err:=i.RefreshOnMoveTo(is.ID,currentState,is.State)
		//if err!=nil{
		//	return err
		//}
	case BoardTypeAssignee:
		currentAssignee := is.Assignee
		is.Assignee = i.State.DropTarget.(string)
		if currentAssignee == is.Assignee {
			return nil
		}
		err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
	case BoardTypePriority:
		currentPriority := is.Priority
		is.Priority = apistructs.IssuePriority(i.State.DropTarget.(string))
		if is.Priority == currentPriority {
			return nil
		}
		err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
	case BoardTypeTime:
		logrus.Infof("drag ignore board type: time")
	case BoardTypeCustom:
		err = i.bdl.UpdateIssuePanelIssue(i.sdk.Identity.UserID, int64(i.State.DropTarget.(float64)), is.ID, int64(i.State.FilterConditions.ProjectID))
	default:
		err := fmt.Errorf("invalid board type, only support: [%v]", SupportBoardTypes)
		logrus.Errorf(err.Error())
		return err
	}
	if err != nil {
		logrus.Errorf("update issue failed, req:%v, err:%v", req, err)
		return err
	}
	return nil
}

func (i *ComponentIssueBoard) RenderOnMoveTo(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, false)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}

	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}
	//from := is.State
	//to := req.StateID
	is.State = req.StateID
	err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
	if err != nil {
		logrus.Errorf("update issue failed, req:%v, err:%v", req, err)
		return err
	}
	//err = i.RefreshOnMoveTo(req.IssueID, from, to)
	//if err != nil {
	//	logrus.Errorf("refresh on move to failed, issueID:%d, from:%d, to:%d", req.IssueID, from, to)
	//}
	return nil
}
func (i *ComponentIssueBoard) RenderOnMoveToAssignee(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, false)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}
	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}
	//from := is.State
	//to := req.StateID
	is.Assignee = req.IssueAssignee
	err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
	if err != nil {
		logrus.Errorf("update issue failed, req:%v, err:%v", req, err)
		return err
	}
	//err = i.RefreshOnMoveTo(req.IssueID, from, to)
	//if err != nil {
	//	logrus.Errorf("refresh on move to failed, issueID:%d, from:%d, to:%d", req.IssueID, from, to)
	//}
	return nil
}

func (i *ComponentIssueBoard) RenderOnMoveToPriority(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, false)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}
	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}
	//from := is.State
	//to := req.StateID
	is.Priority = req.IssuePriority
	err = i.bdl.UpdateIssueTicketUser(i.sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
	if err != nil {
		logrus.Errorf("update issue failed, req:%v, err:%v", req, err)
		return err
	}
	//err = i.RefreshOnMoveTo(req.IssueID, from, to)
	//if err != nil {
	//	logrus.Errorf("refresh on move to failed, issueID:%d, from:%d, to:%d", req.IssueID, from, to)
	//}
	return nil
}

// 全量更新时间长，在改变状态时，可以直接move到另一个状态list，同时删掉当前list中的item
// func (i *ComponentIssueBoard) RefreshOnMoveTo(issueID, from, to int64) error {
// 	is, err := i.bdl.GetIssue(uint64(issueID))
// 	if err != nil {
// 		logrus.Errorf("get issue failed, req:%v, err:%v", issueID, err)
// 		return err
// 	}
// 	c := GenCart(i.boardType, *is, i.ctxBdl.I18nPrinter, i.swt, nil)
// 	for k, v := range i.Data.Board {
// 		if v.LabelKey.(int64) == from {
// 			i.Data.Board[k].Delete(issueID)
// 			i.Data.Board[k].Total = i.Data.Board[k].Total - 1
// 		}
// 		if v.LabelKey.(int64) == to {
// 			i.Data.Board[k].Add(c)
// 			i.Data.Board[k].Total = i.Data.Board[k].Total + 1
// 		}
// 	}
// 	return nil
// }

// TODO 增加自定义看板
func (i *ComponentIssueBoard) RenderOnAddCustom() error {
	var ipr apistructs.IssuePanelRequest
	ipr.PanelName = i.State.PanelName
	ipr.ProjectID = i.State.FilterConditions.ProjectID
	ipr.UserID = i.sdk.Identity.UserID
	_, err := i.bdl.CreateIssuePanel(ipr)
	if err != nil {
		logrus.Errorf("add panel failed, project:%v, err:%v", i.State.FilterConditions.ProjectID, err)
		return err
	}
	return nil
}

// RenderOnUpdateCustom 更新自定义看板
func (i *ComponentIssueBoard) RenderOnUpdateCustom() error {
	var req apistructs.IssuePanelRequest
	req.IssuePanel = i.State.IssuePanel
	req.ProjectID = i.State.FilterConditions.ProjectID
	req.UserID = i.sdk.Identity.UserID
	_, err := i.bdl.UpdateIssuePanel(req)
	if err != nil {
		logrus.Errorf("update panel failed, project:%v, err:%v", i.State.FilterConditions.ProjectID, err)
		return err
	}
	return nil
}

// RenderOnDeleteCustom 删除自定义看板
func (i *ComponentIssueBoard) RenderOnDeleteCustom() error {
	var req apistructs.IssuePanelRequest
	req.IssuePanel = i.State.IssuePanel
	req.ProjectID = i.State.FilterConditions.ProjectID
	req.UserID = i.sdk.Identity.UserID
	_, err := i.bdl.DeleteIssuePanel(req)
	if err != nil {
		logrus.Errorf("delete panel failed, project:%v, panelID:%v, err:%v", i.State.FilterConditions.ProjectID, i.State.PanelID, err)
		return err
	}
	return nil
}

func (i *ComponentIssueBoard) RenderOnMoveToCustom(opsData interface{}) error {
	req, err := GetCartOpsInfo(opsData, false)
	if err != nil {
		logrus.Errorf("get ops data failed, state:%v, err:%v", opsData, err)
		return err
	}
	is, err := i.bdl.GetIssue(uint64(req.IssueID))
	if err != nil {
		logrus.Errorf("get issue failed, req:%v, err:%v", req, err)
		return err
	}
	//from := i.State.PanelID
	to := req.PanelID
	err = i.bdl.UpdateIssuePanelIssue(i.sdk.Identity.UserID, to, is.ID, int64(i.State.FilterConditions.ProjectID))
	if err != nil {
		logrus.Errorf("update panel issue failed, req:%v, err:%v", req, err)
		return err
	}
	//err = i.RefreshOnMoveTo(req.IssueID, from, to)
	//if err != nil {
	//	logrus.Errorf("refresh on move to failed, issueID:%d, from:%d, to:%d", req.IssueID, from, to)
	//}
	return nil
}

func (i *ComponentIssueBoard) RenderDefault(c *cptype.Component, g *cptype.GlobalStateData) {

}

func (i *ComponentIssueBoard) RenderProtocol(c *cptype.Component, g *cptype.GlobalStateData) {
	if c.Data == nil {
		d := make(cptype.ComponentData)
		c.Data = d
	}
	(*c).Data["board"] = i.Data.Board
	(*c).Data["refreshBoard"] = i.Data.RefreshBoard
	(*g)[protocol.GlobalInnerKeyUserIDs.String()] = i.Data.UserIDs
}

func (i *ComponentIssueBoard) Render(ctx context.Context, c *cptype.Component, _ cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) (err error) {
	err = i.GenComponentState(c)
	if err != nil {
		return
	}

	i.sdk = cputil.SDK(ctx)
	i.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)

	visable := make(map[string]bool)
	visable["visible"] = false
	if i.State.IssueViewGroupValue != "kanban" {
		visable["visible"] = false
		c.Props = visable
		return
	}
	visable["visible"] = true
	visable["isLoadMore"] = true
	c.Props = visable

	fReq, err := i.GetFilterReq()
	if err != nil {
		logrus.Errorf("get filter request failed, content:%+v, err:%v", *gs, err)
		return
	}
	//err = i.SetBoardDate(*c)
	//if err!=nil{
	//	return
	//}

	i.SetBoardType(fReq.BoardType)
	err = i.SetOperationSwitch(fReq)
	if err != nil {
		logrus.Errorf("set operation switch failed, request:%+v, err:%v", fReq, err)
		return
	}

	if strings.HasPrefix(event.Operation.String(), apistructs.MoveToCustomOperation.String()) {
		event.Operation = cptype.OperationKey(apistructs.MoveToCustomOperation)
	} else if strings.HasPrefix(event.Operation.String(), apistructs.MoveToAssigneeOperation.String()) {
		event.Operation = cptype.OperationKey(apistructs.MoveToAssigneeOperation)
	} else if strings.HasPrefix(event.Operation.String(), apistructs.MoveToPriorityOperation.String()) {
		event.Operation = cptype.OperationKey(apistructs.MoveToPriorityOperation)
	} else if strings.HasPrefix(event.Operation.String(), apistructs.MoveToOperation.String()) {
		event.Operation = cptype.OperationKey(apistructs.MoveToOperation)
	}

	switch apistructs.OperationKey(event.Operation) {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		err = i.GetDefaultFilterReq(fReq)
		if err != nil {
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.FilterOperation:
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.MoveOutOperation:
		err = i.RenderOnMoveOut(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action state failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.MoveToOperation:
		err = i.RenderOnMoveTo(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action state failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.MoveToAssigneeOperation:
		err = i.RenderOnMoveToAssignee(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action state failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.MoveToCustomOperation:
		err = i.RenderOnMoveToCustom(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.MoveToPriorityOperation:
		err = i.RenderOnMoveToPriority(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.DragOperation:
		err = i.RenderOnDrag(event.OperationData)
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.CreateCustomOperation:
		err = i.RenderOnAddCustom()
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.UpdateCustomOperation:
		// TODO
		err = i.RenderOnUpdateCustom()
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.DeleteCustomOperation:
		err = i.RenderOnDeleteCustom()
		if err != nil {
			logrus.Errorf("generate action custom failed,  err:%v", err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	case apistructs.ChangePageNoOperation:
		if err := i.setChangeNoOperationReq(fReq, event); err != nil {
			logrus.Errorf("render on setChangeNoOperationReq failed, request:%+v, err:%v", *fReq, err)
			return err
		}
		err = i.RenderOnFilter(ctx, *fReq)
		if err != nil {
			logrus.Errorf("render on filter failed, request:%+v, err:%v", *fReq, err)
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, use default operation instead", event.Operation)
	}
	if err = i.CheckUserPermission(fReq.ProjectID); err != nil {
		return err
	}
	c.Operations = i.Operations
	i.RenderProtocol(c, gs)
	return
}

// setChangeNoOperationReq set the pageNo,pageSize and kanbanKey of req and return kanbanKey
func (i *ComponentIssueBoard) setChangeNoOperationReq(req *IssueFilterRequest, event cptype.ComponentEvent) error {
	dataStr, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	data := ChangePageNoOperationData{}
	if err = json.Unmarshal(dataStr, &data); err != nil {
		return err
	}
	if data.Meta.PageData.PageNo != 0 && data.Meta.PageData.PageSize != 0 {
		req.PageNo = data.Meta.PageData.PageNo
		req.PageSize = data.Meta.PageData.PageSize
	}
	if data.Meta.KanbanKey != "" {
		req.KanbanKey = data.Meta.KanbanKey.(string)
	}
	return nil
}
