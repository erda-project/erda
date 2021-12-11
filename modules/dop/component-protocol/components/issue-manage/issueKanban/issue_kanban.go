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
	"sync"
	"time"

	"github.com/golang-collections/collections/set"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/common"
	issue_svc "github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"
)

type IssueCart struct {
	// i18nPrinter    *message.Printer
	ctx            context.Context
	Assignee       string                        `json:"assignee"`
	Priority       apistructs.IssuePriority      `json:"priority"`
	ID             int64                         `json:"id"`
	Title          string                        `json:"title"`
	Type           apistructs.IssueType          `json:"type"`
	IssueButton    []apistructs.IssueStateButton `json:"issueButton"`
	IterationID    int64                         `json:"iterationID"`
	PlanFinishedAt *time.Time                    `json:"planFinishedAt"`
	Operations     map[string]interface{}        `json:"operations"`

	Status CardStatus  `json:"status,omitempty"`
	Labels *CardLabels `json:"labels,omitempty"`
}

type CardStatus struct {
	Text   string `json:"text,omitempty"`
	Status string `json:"status,omitempty"`
}
type CardLabels struct {
	Value []CardLabel `json:"value,omitempty"`
}
type CardLabel struct {
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
}

type CartList struct {
	// i18nPrinter *message.Printer
	ctx context.Context
	// 分类类型: 状态，处理人，时间，自定义,优先级
	Label      interface{}            `json:"label"`
	LabelKey   interface{}            `json:"labelKey"`
	Total      uint64                 `json:"total"`
	PageNo     uint64                 `json:"pageNo"`
	PageSize   uint64                 `json:"pageSize"`
	List       []IssueCart            `json:"list"`
	Operations map[string]interface{} `json:"operations"`
}

type IssueBoard struct {
	RefreshBoard bool       `json:"refreshBoard"`
	Board        []CartList `json:"board"`
	UserIDs      []string   `json:"userIDs"`
}

func (cl *CartList) Delete(issueID int64) {
	for k, v := range cl.List {
		if issueID == v.ID {
			cl.List = append(cl.List[:k], cl.List[k+1:]...)
		}
	}
}

func (cl *CartList) Add(c IssueCart) {
	cl.List = append(cl.List, c)
}

const (
	MoveOutConfirmMsg           = "confirm-to-move-out-iteration"
	CreateBoardConfirmMsg       = "confirm-to-create-board"
	DeleteBoardConfirmMsg       = "confirm-to-delete-board"
	UpdateBoardConfirmMsg       = "confirm-to-update-board"
	CreateBoardDisabledTip      = "the-number-of-boards-cannot-exceed-15"
	GuestCreateBoardDisabledTip = "no-permission-to-operate"
)

type OperationBaseInfo struct {
	// 操作展示名称
	Text string `json:"text"`
	// 是否有权限操作
	Disabled bool `json:"disabled"`
	// 确认提示
	Confirm string `json:"confirm,omitempty"`
	// 前端操作是否需要触发后端
	Reload      bool   `json:"reload"`
	DisabledTip string `json:"disabledTip"`
}

type OpMetaInfo struct {
	IssueID       int64                    `json:"ID"`
	IssueAssignee string                   `json:"issueAssignee"`
	IssuePriority apistructs.IssuePriority `json:"issuePriority"`
	StateID       int64                    `json:"stateID"`
	apistructs.IssuePanel
}

type OperationInfo struct {
	OperationBaseInfo
	Meta OpMetaInfo `json:"meta"`
}

type DragOperationInfo struct {
	Meta OpMetaInfo `json:"meta"`
	// 前端操作是否需要触发后端
	Reload bool `json:"reload"`
	// 可拖拽的范围
	TargetKeys interface{} `json:"targetKeys"`
	Async      bool        `json:"async,omitempty"`
	Disabled   bool        `json:"disabled"`
}

type MoveOutOperation OperationInfo

// 状态
type DragOperation DragOperationInfo
type MoveToOperation OperationInfo

// 处理人
//type DragToAssigneeOperation DragOperationInfo
type MoveToAssigneeOperation OperationInfo

// 优先级
type MoveToPriorityOperation OperationInfo

// 自定义看板
type CreateBoardOperation OperationInfo
type DeleteBoardOperation OperationInfo
type UpdateBoardOperation OperationInfo
type MoveToCustomOperation OperationInfo

type ChangePageNoOperation struct {
	Key      string `json:"key"`
	Reload   bool   `json:"reload"`
	FillMeta string `json:"fillMeta"`
	Meta     struct {
		KanbanKey string `json:"kanbanKey"`
	} `json:"meta"`
}

const MaxBoardNum = 15

//type DragToCustomOperation DragOperationInfo

func (c *CartList) SetCtx(ctx context.Context) {
	c.ctx = ctx
}

func (c *IssueCart) SetCtx(ctx context.Context) {
	c.ctx = ctx
}

// 移除操作渲染生成
func (c *IssueCart) RenderMoveOutOperation() {
	// 已经move out iteration
	if c.IterationID < 0 {
		return
	}
	o := MoveOutOperation{}
	o.Disabled = false
	o.Confirm = cputil.I18n(c.ctx, MoveOutConfirmMsg)
	o.Text = cputil.I18n(c.ctx, apistructs.MoveOutOperation.String())
	o.Reload = true
	o.Meta = OpMetaInfo{IssueID: c.ID}
	c.Operations[apistructs.MoveOutOperation.String()] = o
}

// 拖拽操作渲染生成
func (c *IssueCart) RenderDragOperation() {
	targetKeys := make(map[int64]bool)
	for _, v := range c.IssueButton {
		if v.Permission == true {
			targetKeys[v.StateID] = true
		}
	}
	o := DragOperation{}
	o.Reload = true
	o.Meta = OpMetaInfo{IssueID: c.ID}
	o.TargetKeys = targetKeys
	o.Async = true
	c.Operations[apistructs.DragOperation.String()] = o
}

func (c *IssueCart) RenderMoveToOperation() {
	for _, v := range c.IssueButton {
		o := MoveToOperation{}
		o.Disabled = !v.Permission
		o.Text = cputil.I18n(c.ctx, apistructs.MoveToOperation.String()) + v.StateName
		o.Reload = true
		o.Meta = OpMetaInfo{IssueID: c.ID, StateID: v.StateID}
		c.Operations[apistructs.MoveToOperation.String()+v.StateName] = o
	}
}

func (c *IssueCart) RenderMoveToCustomOperation(mp map[cptype.OperationKey]interface{}) {
	p := mp[cptype.OperationKey(apistructs.MoveToCustomOperation)].([]apistructs.IssuePanelIssues)
	for _, v := range p {
		o := MoveToCustomOperation{}
		o.Disabled = false
		o.Text = cputil.I18n(c.ctx, apistructs.MoveToCustomOperation.String()) + v.PanelName
		o.Reload = true
		o.Meta = OpMetaInfo{
			IssueID:    c.ID,
			IssuePanel: v.IssuePanel,
		}
		c.Operations[apistructs.MoveToCustomOperation.String()+v.PanelName] = o
	}
}

func (c *IssueCart) RenderDragToCustomOperation(mp map[cptype.OperationKey]interface{}) {
	targetKeys := make(map[int64]bool)
	p := mp[cptype.OperationKey(apistructs.DragToCustomOperation)].([]apistructs.IssuePanelIssues)
	for _, v := range p {
		targetKeys[v.PanelID] = true
	}
	o := DragOperation{}
	o.Reload = true
	o.Meta = OpMetaInfo{
		IssueID: c.ID,
	}
	o.Async = true
	o.TargetKeys = targetKeys
	c.Operations[apistructs.DragOperation.String()] = o
}

func (c *IssueCart) RenderMoveToAssigneeOperation(mp map[cptype.OperationKey]interface{}) {
	p := mp[cptype.OperationKey(apistructs.MoveToAssigneeOperation)].([]apistructs.UserInfo)
	for _, v := range p {
		if v.Name == c.Assignee {
			continue
		}
		o := MoveToAssigneeOperation{}
		o.Disabled = false
		o.Text = cputil.I18n(c.ctx, apistructs.MoveToAssigneeOperation.String()) + v.Nick
		o.Reload = true
		o.Meta = OpMetaInfo{
			IssueID:       c.ID,
			IssueAssignee: v.ID,
		}
		c.Operations[apistructs.MoveToAssigneeOperation.String()+v.Name] = o
	}
}

func (c *IssueCart) RenderDragToAssigneeOperation(mp map[cptype.OperationKey]interface{}) {
	targetKeys := make(map[string]bool)
	p := mp[cptype.OperationKey(apistructs.DragToAssigneeOperation)].([]apistructs.UserInfo)
	for _, v := range p {
		targetKeys[v.ID] = true
	}
	o := DragOperation{}
	o.Reload = true
	o.Meta = OpMetaInfo{
		IssueID: c.ID,
	}
	o.Async = true
	o.TargetKeys = targetKeys
	c.Operations[apistructs.DragOperation.String()] = o
}

func (c *IssueCart) RenderMoveToPriorityOperation(i apistructs.Issue, mp map[cptype.OperationKey]interface{}) {
	p := mp[cptype.OperationKey(apistructs.MoveToPriorityOperation)].([]apistructs.IssuePriority)
	for _, v := range p {
		o := MoveToPriorityOperation{}
		o.Disabled = false
		if v == i.Priority {
			o.Disabled = true
		}
		o.Text = cputil.I18n(c.ctx, apistructs.MoveToPriorityOperation.String()) + v.GetZhName()
		o.Reload = true
		o.Meta = OpMetaInfo{
			IssueID:       c.ID,
			IssuePriority: v,
		}
		c.Operations[apistructs.MoveToPriorityOperation.String()+string(v)] = o
	}
}

func (c *IssueCart) RenderDragToPriorityOperation(i apistructs.Issue, mp map[cptype.OperationKey]interface{}) {
	targetKeys := make(map[apistructs.IssuePriority]bool)
	p := mp[cptype.OperationKey(apistructs.DragToPriorityOperation)].([]apistructs.IssuePriority)
	for _, v := range p {
		if v == i.Priority {
			continue
		}
		targetKeys[v] = true
	}
	o := DragOperation{}
	o.Reload = true
	o.Async = true
	o.Meta = OpMetaInfo{
		IssueID: c.ID,
	}
	o.TargetKeys = targetKeys
	c.Operations[apistructs.DragOperation.String()] = o
}

func (c *CartList) RenderChangePageNoOperation(kanbanKey string) {
	if c.Operations == nil {
		c.Operations = make(map[string]interface{})
	}
	o := ChangePageNoOperation{
		Key:      apistructs.ChangePageNoOperation.String(),
		Reload:   true,
		FillMeta: "pageData",
		Meta: struct {
			KanbanKey string `json:"kanbanKey"`
		}{kanbanKey},
	}
	c.Operations[apistructs.ChangePageNoOperation.String()] = o
}

func (c *IssueCart) RenderCartOperations(s ChartOperationSwitch, i apistructs.Issue, mp map[cptype.OperationKey]interface{}) {
	if c.Operations == nil {
		c.Operations = make(map[string]interface{})
	}

	// for task 60429. pd's roadmap, temporarily remove 'Move Out of Iteration',
	// and add a button for more options later.
	// if s.enableMoveOut {
	// 	c.RenderMoveOutOperation()
	// }
	if s.enableMoveTo {
		c.RenderMoveToOperation()
	}
	if s.enableDrag {
		c.RenderDragOperation()
	}
	if s.enableMoveToCustom {
		c.RenderMoveToCustomOperation(mp)
	}
	if s.enableDragToCustom {
		c.RenderDragToCustomOperation(mp)
	}
	if s.enableMoveToAssignee {
		c.RenderMoveToAssigneeOperation(mp)
	}
	if s.enableDragToAssignee {
		c.RenderDragToAssigneeOperation(mp)
	}
	if s.enableMoveToPriority {
		c.RenderMoveToPriorityOperation(i, mp)
	}
	if s.enableDragToPriority {
		c.RenderDragToPriorityOperation(i, mp)
	}
}

func (cl *CartList) RenderCartListOperations(s ChartOperationSwitch) {
	cl.Operations = make(map[string]interface{})
	// 删除看板
	if len(cl.List) == 0 && cl.LabelKey.(int64) != 0 {
		o := DeleteBoardOperation{}
		o.Disabled = false
		o.Confirm = cputil.I18n(cl.ctx, DeleteBoardConfirmMsg)
		o.Text = cputil.I18n(cl.ctx, apistructs.DeleteCustomOperation.String())
		o.Reload = true
		o.Meta = OpMetaInfo{IssuePanel: apistructs.IssuePanel{PanelID: cl.LabelKey.(int64)}}
		cl.Operations[apistructs.DeleteCustomOperation.String()] = o
	}
	// 更新看板
	if cl.LabelKey.(int64) != 0 {
		o := UpdateBoardOperation{}
		o.Disabled = false
		o.Confirm = cputil.I18n(cl.ctx, UpdateBoardConfirmMsg)
		o.Text = cputil.I18n(cl.ctx, apistructs.UpdateCustomOperation.String())
		o.Reload = true
		o.Meta = OpMetaInfo{IssuePanel: apistructs.IssuePanel{PanelID: cl.LabelKey.(int64)}}
		cl.Operations[apistructs.UpdateCustomOperation.String()] = o
	}
	if s.enableChangePageNo {
		cl.RenderChangePageNoOperation(strconv.FormatInt(cl.LabelKey.(int64), 10))
	}
}

// 根据完成时间(planFinishedAt)分为：未分类，已过期，1天内过期，2天内，3天内，30天，未来
type ExpireType string

func (e ExpireType) String() string {
	return string(e)
}

const (
	ExpireTypeUndefined      ExpireType = "Undefined"
	ExpireTypeExpired        ExpireType = "Expired"
	ExpireTypeExpireIn1Day   ExpireType = "ExpireIn1Day"
	ExpireTypeExpireIn2Days  ExpireType = "ExpireIn2Days"
	ExpireTypeExpireIn7Days  ExpireType = "ExpireIn7Days"
	ExpireTypeExpireIn30Days ExpireType = "ExpireIn30Days"
	ExpireTypeExpireInFuture ExpireType = "ExpireInFuture"
)

var ExpireTypes = []ExpireType{ExpireTypeUndefined, ExpireTypeExpired, ExpireTypeExpireIn1Day, ExpireTypeExpireIn2Days, ExpireTypeExpireIn7Days, ExpireTypeExpireIn30Days, ExpireTypeExpireInFuture}

// FinishTime和now都是date；hour, min, sec 都为0
func (c IssueCart) GetExpireType(now time.Time) ExpireType {
	if c.PlanFinishedAt == nil {
		return ExpireTypeUndefined
	}
	finishTime := *c.PlanFinishedAt
	if finishTime.Before(now) {
		return ExpireTypeExpired
	} else if finishTime.Before(now.Add(1 * 24 * time.Hour)) {
		return ExpireTypeExpireIn1Day
	} else if finishTime.Before(now.Add(2 * 24 * time.Hour)) {
		return ExpireTypeExpireIn2Days
	} else if finishTime.Before(now.Add(7 * 24 * time.Hour)) {
		return ExpireTypeExpireIn7Days
	} else if finishTime.Before(now.Add(30 * 24 * time.Hour)) {
		return ExpireTypeExpireIn30Days
	} else {
		return ExpireTypeExpireInFuture
	}
}

type ChartOperationSwitch struct {
	// 所有迭代时，disable move out
	enableMoveOut bool
	// disable drag
	enableDrag bool
	// 按状态分类
	enableMoveTo bool

	// 按处理人分类
	enableMoveToAssignee bool
	enableDragToAssignee bool
	// 按自定义看板分类
	enableMoveToCustom bool
	enableDragToCustom bool
	// 按优先级分类
	enableMoveToPriority bool
	enableDragToPriority bool

	enableChangePageNo bool
}

func (c *ChartOperationSwitch) ClearAble() {
	c.enableMoveOut = false
	c.enableDrag = false
	c.enableMoveTo = false
	c.enableMoveToAssignee = false
	c.enableDragToAssignee = false
	c.enableMoveToCustom = false
	c.enableDragToCustom = false
	c.enableMoveToPriority = false
	c.enableDragToPriority = false
	c.enableChangePageNo = false
}

type ComponentIssueBoard struct {
	sdk *cptype.SDK
	bdl *bundle.Bundle

	issueSvc      *issue_svc.Issue
	issueStateSvc *issuestate.IssueState

	boardType BoardType
	swt       ChartOperationSwitch

	CompName   string                 `json:"name"`
	Data       IssueBoard             `json:"data"`
	Operations map[string]interface{} `json:"operations"`
	State      IssueBoardState        `json:"state"`

	base.DefaultProvider
}

type IssueBoardState struct {
	DropTarget interface{} `json:"dropTarget"`
	apistructs.IssuePanel
	Priority                    apistructs.IssuePriority      `json:"priority"`
	FilterConditions            apistructs.IssuePagingRequest `json:"filterConditions"`
	IssueViewGroupChildrenValue map[string]string             `json:"issueViewGroupChildrenValue"`
	IssueViewGroupValue         string                        `json:"issueViewGroupValue"`
}

type IssueFilterRequest struct {
	apistructs.IssuePagingRequest
	BoardType BoardType `json:"boardType"`
	KanbanKey string    `json:"kanbanKey"`
}

type issueRenderInparams struct {
	ProjectID           string               `json:"projectID"`
	BoardType           BoardType            `json:"boardType"`
	FixedIssueType      apistructs.IssueType `json:"fixedIssueType"`
	FixedIssueIteration string               `json:"fixedIssueIteration"`
}
type BoardType string

func (b BoardType) String() string {
	return string(b)
}

const (
	BoardTypeStatus   BoardType = "status"
	BoardTypeTime     BoardType = "deadline"
	BoardTypeAssignee BoardType = "assignee" // 已经移除处理人看板
	BoardTypeCustom   BoardType = "custom"
	BoardTypePriority BoardType = "priority"
)

var SupportBoardTypes = []BoardType{BoardTypeStatus, BoardTypeTime, BoardTypeAssignee, BoardTypeCustom}

var IssueTypes = []apistructs.IssueType{apistructs.IssueTypeTask, apistructs.IssueTypeRequirement, apistructs.IssueTypeBug, apistructs.IssueTypeEpic}

var IssueTypeStates = map[apistructs.IssueType][]apistructs.IssueState{
	apistructs.IssueTypeTask:        {apistructs.IssueStateOpen, apistructs.IssueStateWorking, apistructs.IssueStateDone},
	apistructs.IssueTypeRequirement: {apistructs.IssueStateOpen, apistructs.IssueStateWorking, apistructs.IssueStateDone},
	apistructs.IssueTypeBug:         {apistructs.IssueStateOpen, apistructs.IssueStateWontfix, apistructs.IssueStateReopen, apistructs.IssueStateResolved, apistructs.IssueStateClosed},
}

func uniq(in []string) []string {
	var out []string
	s := set.New()
	for _, v := range in {
		if !s.Has(v) {
			s.Insert(v)
			out = append(out, v)
		}
	}
	return out
}

func (i *ComponentIssueBoard) SetOperationSwitch(req *IssueFilterRequest) error {
	i.swt.ClearAble()
	switch i.boardType {
	case BoardTypeAssignee:
		i.swt.enableMoveToAssignee = true
		i.swt.enableDragToAssignee = true
	case BoardTypeTime:
	case BoardTypeStatus:
		i.swt.enableMoveTo = true
		i.swt.enableDrag = true
	case BoardTypePriority:
		i.swt.enableMoveToPriority = true
		i.swt.enableDragToPriority = true
	case BoardTypeCustom:
		i.swt.enableMoveToCustom = true
		i.swt.enableDragToCustom = true
	default:
		return nil
	}
	// 全部看板都可以移出迭代
	i.swt.enableMoveOut = true
	i.swt.enableChangePageNo = true
	return nil
}

func (i *ComponentIssueBoard) SetBoardDate(c cptype.Component) error {
	var board []CartList
	cont, err := json.Marshal(c.Data["board"].([]CartList))
	if err != nil {
		logrus.Errorf("marshal component data failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &board)
	if err != nil {
		logrus.Errorf("unmarshal component dat failed, content:%v, err:%v", cont, err)
		return err
	}
	i.Data.Board = board
	return nil
}

func (i *ComponentIssueBoard) SetBoardType(bt BoardType) {
	i.boardType = bt
}

// GetUserPermission  check Guest permission
func (i *ComponentIssueBoard) CheckUserPermission(projectID uint64) error {
	isGuest := false
	scopeRole, err := i.bdl.ScopeRoleAccess(i.sdk.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   strconv.FormatUint(projectID, 10),
		},
	})
	if err != nil {
		return err
	}
	for _, role := range scopeRole.Roles {
		if role == "Guest" {
			isGuest = true
		}
	}
	if !isGuest {
		return nil
	}
	// if the permission of user is guest , could't to do operation
	for k, v := range i.Operations {
		i.Operations[k] = i.disableOperationPermission(v)
	}
	for ib, board := range i.Data.Board {
		for k, v := range board.Operations {
			i.Data.Board[ib].Operations[k] = i.disableOperationPermission(v)
		}
		for ic, ca := range board.List {
			for k, v := range ca.Operations {
				i.Data.Board[ib].List[ic].Operations[k] = i.disableOperationPermission(v)
			}
		}
	}
	return nil
}

func (i *ComponentIssueBoard) disableOperationPermission(op interface{}) interface{} {
	switch op.(type) {
	case MoveOutOperation:
		o := op.(MoveOutOperation)
		o.Disabled = true
		return o
	case DragOperation:
		o := op.(DragOperation)
		o.Disabled = true
		return o
	case MoveToOperation:
		o := op.(MoveToOperation)
		o.Disabled = true
		return o
	case MoveToPriorityOperation:
		o := op.(MoveToPriorityOperation)
		o.Disabled = true
		return o
	case CreateBoardOperation:
		o := op.(CreateBoardOperation)
		o.Disabled = true
		o.DisabledTip = i.sdk.I18n(GuestCreateBoardDisabledTip)
		return o
	case DeleteBoardOperation:
		o := op.(DeleteBoardOperation)
		o.Disabled = true
		return o
	case UpdateBoardOperation:
		o := op.(UpdateBoardOperation)
		o.Disabled = true
		return o
	case MoveToCustomOperation:
		o := op.(MoveToCustomOperation)
		o.Disabled = true
		return o
	default:
		panic("errors: invalid operationInfo type")
	}
	return nil
}

func (i *ComponentIssueBoard) Filter(ctx context.Context, req IssueFilterRequest) (ib *IssueBoard, err error) {
	// statuses
	stateByIssueType, err := i.issueStateSvc.GetIssueStatesMap(&apistructs.IssueStatesGetRequest{ProjectID: req.ProjectID})
	if err != nil {
		return nil, err
	}
	stateByStateID := make(map[int64]apistructs.IssueStatus)
	for _, statuses := range stateByIssueType {
		for _, status := range statuses {
			stateByStateID[status.StateID] = status
		}
	}

	var (
		cls   []CartList
		board IssueBoard
		uids  []string
	)
	i.Operations = make(map[string]interface{})
	// 事件分类: 所有，需求，任务，缺陷
	switch req.BoardType {
	// 所有，不含此分类
	case BoardTypeStatus:
		cls, uids, err = i.FilterByStatusConcurrent(ctx, req.IssuePagingRequest, req.KanbanKey, stateByStateID)
	case BoardTypeTime:
		cls, uids, err = i.FilterByTime(ctx, req.IssuePagingRequest, req.KanbanKey, stateByStateID)
	case BoardTypePriority:
		cls, uids, err = i.FilterByPriority(ctx, req.IssuePagingRequest, req.KanbanKey, stateByStateID)
	// 所有，不含此分类
	case BoardTypeCustom:
		cls, uids, err = i.FilterByCustom(ctx, req.IssuePagingRequest, req.KanbanKey, stateByStateID)
		// Created custom board max num 15
		o := CreateBoardOperation{}
		o.Disabled = false
		o.Confirm = i.sdk.I18n(CreateBoardConfirmMsg)
		o.Text = i.sdk.I18n(apistructs.CreateCustomOperation.String())
		o.Reload = true
		o.DisabledTip = i.sdk.I18n(CreateBoardDisabledTip)
		if len(cls) > MaxBoardNum {
			o.Disabled = true
		}
		i.Operations[apistructs.CreateCustomOperation.String()] = o
	default:
		//err = fmt.Errorf("invalid board type [%s], must in [%v]", req.BoardType, SupportBoardTypes)
		cls, uids, err = i.FilterByPriority(ctx, req.IssuePagingRequest, req.KanbanKey, stateByStateID)
	}
	if err != nil {
		return
	}
	// 时间分类i18n
	for k, v := range cls {
		cls[k].Label = i.sdk.I18n(v.Label.(string))
	}
	if req.KanbanKey != "" {
		board.RefreshBoard = false
	} else {
		board.RefreshBoard = true
	}
	board.Board = cls
	board.UserIDs = uniq(uids)
	ib = &board
	return
}

func (i ComponentIssueBoard) GetByStatus(req apistructs.IssuePagingRequest, ignoreDone bool) (rsp *apistructs.IssuePagingResponse, err error) {
	if len(req.Type) == 0 || len(req.Type) != 1 {
		err = fmt.Errorf("issue type number is not 1, type: %v", req.Type)
		return
	}
	if req.PageSize == 0 {
		req.PageSize = 200
	}
	// 获取当前项目，特定IssueType的IssueStates
	//it := req.Type[0]
	//isReq := apistructs.IssueStateRelationGetRequest{ProjectID: int64(req.ProjectID), IssueType: it}
	//is, err := i.bdl.GetIssueStateBelong(isReq)
	//if err != nil {
	//	logrus.Errorf("get issue state belong failed, request:%+v, err:%v", isReq, err)
	//	return
	//}

	//var states []apistructs.IssueStateName
	//var issueStates []int64
	//for _, v := range is {
	//	if ignoreDone && (v.StateBelong == apistructs.IssueStateBelongClosed || v.StateBelong == apistructs.IssueStateBelongDone) {
	//		continue
	//	}
	//	states = append(states, v.States...)
	//}

	//for _, v := range states {
	//	issueStates = append(issueStates, v.ID)
	//}
	//if ignoreDone {
	//	req.State = issueStates
	//}

	rsp, err = i.bdl.PageIssues(req)
	if err != nil {
		logrus.Errorf("page issues failed, request:%+v, err:%v", req, err)
		return
	}
	return
}

func (i ComponentIssueBoard) GetIssues(req apistructs.IssuePagingRequest, ignoreDone bool) (result []apistructs.Issue, uids []string, err error) {
	if req.Type == nil {
		req.Type = IssueTypes
	}
	var wg sync.WaitGroup

	wg.Add(len(req.Type))
	for _, v := range req.Type {
		v := v
		go func() {
			defer func() {
				wg.Done()
			}()
			if err != nil {
				return
			}
			r := req
			r.Type = []apistructs.IssueType{v}
			rsp, err := i.GetByStatus(r, ignoreDone)
			if err != nil {
				return
			}
			result = append(result, rsp.Data.List...)
			uids = append(uids, rsp.UserIDs...)
		}()
	}
	wg.Wait()
	return
}

func GenCart(bt BoardType, i apistructs.Issue, ctx context.Context, s ChartOperationSwitch, mp map[cptype.OperationKey]interface{}, stateByStateID map[int64]apistructs.IssueStatus) IssueCart {
	var c IssueCart
	c.ID = i.ID
	c.Title = i.Title
	c.Type = i.Type
	c.IssueButton = i.IssueButton
	c.IterationID = i.IterationID
	c.Assignee = i.Assignee
	c.Priority = i.Priority
	c.PlanFinishedAt = i.PlanFinishedAt
	c.Status = func() (cardStatus CardStatus) {
		if len(stateByStateID) == 0 {
			return
		}
		status, ok := stateByStateID[i.State]
		if !ok {
			return
		}
		return CardStatus{
			Text:   status.StateName,
			Status: common.GetUIIssueState(status.StateBelong),
		}
	}()
	c.Labels = func() *CardLabels {
		var labels []CardLabel
		for _, label := range i.LabelDetails {
			labels = append(labels, CardLabel{
				Label: label.Name,
				Color: label.Color,
			})
		}
		if len(labels) == 0 {
			return nil
		}
		return &CardLabels{Value: labels}
	}()
	c.SetCtx(ctx)
	c.RenderCartOperations(s, i, mp)
	return c
}

func (c *CartList) GenCartList(ctx context.Context, s ChartOperationSwitch) {
	c.SetCtx(ctx)
	c.RenderCartListOperations(s)
}

// 按状态过滤 并发
func (i ComponentIssueBoard) FilterByStatusConcurrent(ctx context.Context, req apistructs.IssuePagingRequest, kanbanKey string, stateByStateID map[int64]apistructs.IssueStatus) (cls []CartList, uids []string, err error) {
	if len(req.Type) == 0 || len(req.Type) != 1 {
		err = fmt.Errorf("issue type number is not 1, type: %v", req.Type)
		return
	}
	// 获取当前项目，特定IssueType的IssueStates
	it := req.Type[0]
	isReq := apistructs.IssueStateRelationGetRequest{ProjectID: req.ProjectID, IssueType: it}
	is, err := i.bdl.GetIssueStateBelong(isReq)
	if err != nil {
		logrus.Errorf("get issue state belong failed, request:%+v, err:%v", isReq, err)
		return
	}

	var states []apistructs.IssueStateName
	if kanbanKey == "" {
		for _, v := range is {
			states = append(states, v.States...)
		}
	} else {
		for _, v := range is {
			for _, v2 := range v.States {
				if strconv.FormatInt(v2.ID, 10) == kanbanKey {
					states = append(states, v2)
					goto loop
				}
			}
		}
	}
loop:
	// filter by status is not avialble in status board
	if len(req.StateBelongs) > 0 {
		req.StateBelongs = nil
	}

	date := struct {
		Map  map[int64]CartList
		Lock sync.Mutex
	}{
		Map: make(map[int64]CartList),
	}

	var wg sync.WaitGroup
	wg.Add(len(states))
	for _, v := range states {
		go func(state apistructs.IssueStateName) {
			defer func() {
				wg.Done()
			}()
			if err != nil {
				return
			}

			r := req
			// 按特定种类IssueType的一个IssueState并发查询
			r.State = []int64{state.ID}
			rsp, e := i.bdl.PageIssues(r)
			if e != nil {
				err = e
				logrus.Errorf("page issues failed, request:%+v, err:%v", r, e)
				return
			}
			// 生成CartList
			cl := CartList{}
			cl.Label = state.Name
			cl.LabelKey = state.ID
			cl.Total = rsp.Data.Total
			cl.PageNo = req.PageNo
			cl.PageSize = req.PageSize
			if i.swt.enableChangePageNo {
				cl.RenderChangePageNoOperation(strconv.FormatInt(cl.LabelKey.(int64), 10))
			}
			for _, v := range rsp.Data.List {
				c := GenCart(i.boardType, v, ctx, i.swt, nil, stateByStateID)
				cl.Add(c)
			}

			date.Lock.Lock()
			date.Map[state.ID] = cl
			date.Lock.Unlock()
			// uid
			uids = append(uids, rsp.UserIDs...)
		}(v)
	}
	wg.Wait()

	for _, v := range states {
		cls = append(cls, date.Map[v.ID])
	}
	return
}

// FilterByPriority 按优先级过滤 去掉终态状态 [CLOSED, DONE]
func (i ComponentIssueBoard) FilterByPriority(ctx context.Context, req apistructs.IssuePagingRequest, kanbanKey string, stateByStateID map[int64]apistructs.IssueStatus) (cls []CartList, uids []string, err error) {
	// 用于生成权限
	var priorityList []apistructs.IssuePriority
	if kanbanKey != "" {
		priorityList = append(priorityList, apistructs.IssuePriority(kanbanKey))
	} else {
		priorityList = apistructs.IssuePriorityList
	}
	mp := make(map[cptype.OperationKey]interface{})
	mp[cptype.OperationKey(apistructs.MoveToPriorityOperation)] = priorityList
	mp[cptype.OperationKey(apistructs.DragToPriorityOperation)] = priorityList

	date := struct {
		Map  map[apistructs.IssuePriority]CartList
		Lock sync.Mutex
	}{
		Map: make(map[apistructs.IssuePriority]CartList),
	}

	var wg sync.WaitGroup
	wg.Add(len(priorityList))
	for _, v := range priorityList {
		go func(state apistructs.IssuePriority) {
			defer func() {
				wg.Done()
			}()
			if err != nil {
				return
			}

			r := req
			// 按特定种类IssueType的一个IssueState并发查询
			r.Priority = []apistructs.IssuePriority{state}
			rsp, e := i.bdl.PageIssues(r)
			if e != nil {
				err = e
				logrus.Errorf("page issues failed, request:%+v, err:%v", r, e)
				return
			}
			// 生成CartList
			cl := CartList{}
			cl.Label = string(state)
			cl.LabelKey = cl.Label
			cl.Total = rsp.Data.Total
			cl.PageNo = req.PageNo
			cl.PageSize = req.PageSize
			if i.swt.enableChangePageNo {
				cl.RenderChangePageNoOperation(cl.LabelKey.(string))
			}
			for _, v := range rsp.Data.List {
				c := GenCart(i.boardType, v, ctx, i.swt, mp, stateByStateID)
				cl.Add(c)
			}
			date.Lock.Lock()
			date.Map[state] = cl
			date.Lock.Unlock()
			// uid
			uids = append(uids, rsp.UserIDs...)
		}(v)
	}
	wg.Wait()
	for _, v := range priorityList {
		cls = append(cls, date.Map[v])
	}

	return
}

// FilterByTime 根据完成时间(planFinishedAt)分为：未分类，已过期，1天内过期，2天内，7天内，30天，未来
func (i ComponentIssueBoard) FilterByTime(ctx context.Context, req apistructs.IssuePagingRequest, kanbanKey string, stateByStateID map[int64]apistructs.IssueStatus) (cls []CartList, uids []string, err error) {
	// 为减少事件数量，不需要展示终态的事项[CLOSED, DONE]
	timeMap := getTimeMap()
	// map并发写需要加锁
	date := struct {
		Map  map[ExpireType]CartList
		Lock sync.Mutex
	}{
		Map: make(map[ExpireType]CartList),
	}

	var expireTypes []ExpireType
	if kanbanKey != "" {
		expireTypes = append(expireTypes, ExpireType(kanbanKey))
	} else {
		expireTypes = ExpireTypes
	}

	var wg sync.WaitGroup
	wg.Add(len(expireTypes))

	// 按特定PlanFinishedAt的一个并发查询、
	for _, et := range expireTypes {
		go func(tm ExpireType) {
			defer func() {
				wg.Done()
			}()
			if err != nil {
				return
			}

			r := req
			if tm == ExpireTypeUndefined {
				r.IsEmptyPlanFinishedAt = true
			} else {
				r.StartFinishedAt = timeMap[tm][0] * 1000
				if req.StartFinishedAt != 0 && r.StartFinishedAt < req.StartFinishedAt {
					r.StartFinishedAt = req.StartFinishedAt
				}
				r.EndFinishedAt = timeMap[tm][1] * 1000
				if req.EndFinishedAt != 0 && (r.EndFinishedAt > req.EndFinishedAt || tm == ExpireTypeExpireInFuture) {
					r.EndFinishedAt = req.EndFinishedAt
				}
			}
			rsp, e := i.bdl.PageIssues(r)
			if e != nil {
				err = e
				logrus.Errorf("page issues failed, request:%+v, err:%v", r, e)
				return
			}
			// 生成CartList
			cl := CartList{}
			cl.Label = string(tm)
			cl.LabelKey = cl.Label
			cl.Total = rsp.Data.Total
			cl.PageNo = req.PageNo
			cl.PageSize = req.PageSize
			if i.swt.enableChangePageNo {
				cl.RenderChangePageNoOperation(cl.LabelKey.(string))
			}
			for _, v := range rsp.Data.List {
				c := GenCart(i.boardType, v, ctx, i.swt, nil, stateByStateID)
				cl.Add(c)
			}
			date.Lock.Lock()
			date.Map[tm] = cl
			date.Lock.Unlock()
			// uid
			uids = append(uids, rsp.UserIDs...)
		}(et)
	}
	wg.Wait()
	for _, v := range expireTypes {
		cls = append(cls, date.Map[v])
	}
	return
}

// getTimeMap return timeMap,key: ExpireType,value: struct of startTime and endTime
func getTimeMap() map[ExpireType][]int64 {
	nowTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	tomorrow := nowTime.Add(time.Hour * time.Duration(24))
	twoDay := nowTime.Add(time.Hour * time.Duration(24*2))
	sevenDay := nowTime.Add(time.Hour * time.Duration(24*7))
	thirtyDay := nowTime.Add(time.Hour * time.Duration(24*30))
	timeMap := map[ExpireType][]int64{
		ExpireTypeUndefined:      {0, 0},
		ExpireTypeExpired:        {1, nowTime.Add(time.Second * time.Duration(-1)).Unix()},
		ExpireTypeExpireIn1Day:   {nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},
		ExpireTypeExpireIn2Days:  {tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // 明天
		ExpireTypeExpireIn7Days:  {twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7天
		ExpireTypeExpireIn30Days: {sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30天
		ExpireTypeExpireInFuture: {thirtyDay.Unix(), 0},                                                    // 未来
	}
	return timeMap
}

// 按自定义看板 过滤
func (i ComponentIssueBoard) FilterByCustom(ctx context.Context, req apistructs.IssuePagingRequest, kanbanKey string, stateByStateID map[int64]apistructs.IssueStatus) (cls []CartList, uids []string, err error) {
	rsp, err := i.bdl.GetIssuePanel(apistructs.IssuePanelRequest{IssuePagingRequest: req})
	if err != nil {
		return
	}
	var rspList []apistructs.IssuePanelIssues
	if kanbanKey == "" {
		rspList = rsp
	} else {
		for _, v := range rsp {
			if strconv.FormatInt(v.PanelID, 10) == kanbanKey {
				rspList = append(rspList, v)
				break
			}
		}
	}

	mp := make(map[cptype.OperationKey]interface{})
	mp[cptype.OperationKey(apistructs.MoveToCustomOperation)] = rspList
	mp[cptype.OperationKey(apistructs.DragToCustomOperation)] = rspList

	date := struct {
		Map  map[apistructs.IssuePanelIssues]CartList
		Lock sync.Mutex
	}{
		Map: make(map[apistructs.IssuePanelIssues]CartList),
	}

	var wg sync.WaitGroup
	wg.Add(len(rspList))
	for _, pl := range rspList {
		go func(panel apistructs.IssuePanelIssues) {
			defer func() {
				wg.Done()
			}()
			rsp, e := i.bdl.GetIssuePanelIssue(apistructs.IssuePanelRequest{
				IssuePanel:         apistructs.IssuePanel{PanelID: panel.PanelID},
				IssuePagingRequest: req,
			})
			if e != nil {
				err = e
				return
			}
			cl := CartList{}
			cl.Label = panel.PanelName
			cl.LabelKey = panel.PanelID
			cl.Total = rsp.Total
			cl.PageNo = req.PageNo
			cl.PageSize = req.PageSize
			if rsp.Total > 0 {
				for _, v := range rsp.Issues {
					c := GenCart(i.boardType, v, ctx, i.swt, mp, stateByStateID)
					// 不能转移到自己
					cMove := c.Operations[apistructs.MoveToCustomOperation.String()+panel.PanelName].(MoveToCustomOperation)
					cMove.Disabled = true
					c.Operations[apistructs.MoveToCustomOperation.String()+panel.PanelName] = cMove
					cDrag := c.Operations[apistructs.DragOperation.String()].(DragOperation)
					tagget := cDrag.TargetKeys.(map[int64]bool)
					tagget[panel.PanelID] = false
					cDrag.TargetKeys = tagget
					c.Operations[apistructs.DragOperation.String()] = cDrag
					cl.Add(c)
					uids = append(uids, v.Assignee)
				}
			}

			cl.GenCartList(ctx, i.swt)
			date.Lock.Lock()
			date.Map[panel] = cl
			date.Lock.Unlock()
		}(pl)
	}
	wg.Wait()
	for _, v := range rspList {
		cls = append(cls, date.Map[v])
	}
	uids = strutil.DedupSlice(uids)
	return
}

func init() {
	base.InitProviderWithCreator("issue-manage", "issueKanban", func() servicehub.Provider {
		return &ComponentIssueBoard{}
	})
}
