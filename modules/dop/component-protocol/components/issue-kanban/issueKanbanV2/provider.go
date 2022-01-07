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

package issueKanbanV2

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/kanban"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/kanban/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-kanban/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	issuesvc "github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	stateKeyIssueViewGroupValue         = "issueViewGroupValue"         // kanban
	stateKeyIssueViewGroupChildrenValue = "issueViewGroupChildrenValue" // kanban: status
	stateKeyIssueFilterConditions       = "filterConditions"            // apistructs.IssuePagingRequest

	inParamsKeyFixedIssueType = "fixedIssueType"
	inParamsKeyIterationID    = "fixedIteration" // id
)

type Kanban struct {
	impl.DefaultKanban

	filterReq apistructs.IssuePagingRequest

	bdl           *bundle.Bundle
	issueSvc      *issuesvc.Issue
	issueStateSvc *issuestate.IssueState
}

type IssueCardExtra struct {
	Type        apistructs.IssueType     `json:"type,omitempty"`
	Priority    apistructs.IssuePriority `json:"priority,omitempty"`
	AssigneeID  string                   `json:"assigneeID,omitempty"`
	IterationID int64                    `json:"iterationID,omitempty"`
}

func (e IssueCardExtra) ToExtra() cptype.Extra {
	extraMap := cputil.MustObjJSONTransfer(
		&IssueCardExtra{
			Type:        e.Type,
			Priority:    e.Priority,
			AssigneeID:  e.AssigneeID,
			IterationID: e.IterationID,
		}, &cptype.ExtraMap{}).(*cptype.ExtraMap)
	return cptype.Extra{Extra: *extraMap}
}

func init() {
	base.InitProviderWithCreator("issue-kanban", "issueKanbanV2", func() servicehub.Provider {
		return &Kanban{}
	})
}

func (k *Kanban) Initialize(sdk *cptype.SDK) {}

func (k *Kanban) Finalize(sdk *cptype.SDK) {
	sdk.SetUserIDs(k.StdDataPtr.UserIDs)
}

func (k *Kanban) BeforeHandleOp(sdk *cptype.SDK) {
	k.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	k.issueSvc = sdk.Ctx.Value(types.IssueService).(*issuesvc.Issue)
	k.issueStateSvc = sdk.Ctx.Value(types.IssueStateService).(*issuestate.IssueState)
	gh := gshelper.NewGSHelper(sdk.GlobalState)
	filterCond, ok := gh.GetIssuePagingRequest()
	if !ok {
		panic("empty request")
	}
	k.filterReq = *filterCond
	// issue type
	issueType := apistructs.IssueType(k.StdInParamsPtr.String(inParamsKeyFixedIssueType))
	if issueType != "" {
		k.filterReq.Type = []apistructs.IssueType{issueType}
	}
	if issueType == "ALL" {
		panic("status kanban only support one issue type")
	}
	// iteration id
	iterationID := k.StdInParamsPtr.Uint64(inParamsKeyIterationID)
	if iterationID > 0 {
		k.filterReq.IterationID = int64(iterationID)
	}
	// page
	if k.filterReq.PageSize == 0 {
		k.filterReq.PageSize = 20
	}
}

func (k *Kanban) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		k.StdDataPtr = k.doFilter()
	}
}

func (k *Kanban) doFilter(specificBoardIDs ...string) *kanban.Data {
	// statuses
	stateByIssueType, err := k.issueStateSvc.GetIssueStatesMap(&apistructs.IssueStatesGetRequest{ProjectID: k.filterReq.ProjectID})
	if err != nil {
		panic(err)
	}
	stateByStateID := make(map[int64]apistructs.IssueStatus)
	for _, statuses := range stateByIssueType {
		for _, status := range statuses {
			stateByStateID[status.StateID] = status
		}
	}

	// get specific project-level issue states
	issueType := k.filterReq.Type[0]
	stateBelong, err := k.issueStateSvc.GetIssueStatesBelong(&apistructs.IssueStateRelationGetRequest{ProjectID: k.filterReq.ProjectID, IssueType: issueType})
	if err != nil {
		panic(fmt.Errorf("failed to get issue state belong, err: %v", err))
	}
	var states []apistructs.IssueStateName
	for _, belong := range stateBelong {
		states = append(states, belong.States...)
	}
	// filter by status is not available in status board
	k.filterReq.StateBelongs = nil

	var lock sync.Mutex
	var data kanban.Data
	var wg sync.WaitGroup
	boardsByStateID := make(map[int64]kanban.Board)
	for _, state := range states {
		boardID := strutil.String(state.ID)

		// filter by specific board ids
		if len(specificBoardIDs) > 0 && !strutil.Exist(specificBoardIDs, boardID) {
			continue
		}

		wg.Add(1)
		go func(state apistructs.IssueStateName) {
			defer wg.Done()

			r := k.filterReq
			r.State = []int64{state.ID}
			resp, err := k.bdl.PageIssues(r)
			if err != nil || !resp.Success {
				panic(fmt.Errorf("failed to paging issue, err: %v", err))
			}

			board := kanban.Board{
				ID:    boardID,
				Title: state.Name,
				Cards: func() (cards []kanban.Card) {
					for _, issue := range resp.Data.List {
						cards = append(cards, kanban.Card{
							ID:    strutil.String(issue.ID),
							Title: issue.Title,
							Operations: map[cptype.OperationKey]cptype.Operation{
								kanban.OpCardMoveTo{}.OpKey(): cputil.NewOpBuilder().
									WithAsync(true).
									WithServerDataPtr(&kanban.OpCardMoveToServerData{
										AllowedTargetBoardIDs: func() []string {
											var allowedStateIDs []string
											for _, button := range issue.IssueButton {
												if button.Permission {
													allowedStateIDs = append(allowedStateIDs, strutil.String(button.StateID))
												}
											}
											return strutil.DedupSlice(allowedStateIDs, true)
										}(),
									}).
									Build(),
							},
							Extra: IssueCardExtra{Type: issue.Type, Priority: issue.Priority, AssigneeID: issue.Assignee, IterationID: issue.IterationID}.ToExtra(),
						})
						data.UserIDs = append(data.UserIDs, issue.Assignee)
					}
					return
				}(),
				PageNo:   k.filterReq.PageNo,
				PageSize: k.filterReq.PageSize,
				Total:    resp.Data.Total,
				Operations: map[cptype.OperationKey]cptype.Operation{
					kanban.OpBoardLoadMore{}.OpKey(): cputil.NewOpBuilder().Build(),
				},
			}
			lock.Lock()
			boardsByStateID[state.ID] = board
			lock.Unlock()
		}(state)
	}
	wg.Wait()

	// order board by state
	for _, state := range states {
		if board, exist := boardsByStateID[state.ID]; exist {
			data.Boards = append(data.Boards, board)
		}
	}

	return &data
}

func (k *Kanban) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return k.RegisterInitializeOp()
}

// RegisterBoardCreateOp no need here.
func (k *Kanban) RegisterBoardCreateOp(opData kanban.OpBoardCreate) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		logrus.Infof("mock create board: titie: %s", opData.ClientData.Title)
		k.StdDataPtr = k.doFilter()
		k.StdDataPtr.Boards = append(k.StdDataPtr.Boards, kanban.Board{
			ID:       opData.ClientData.Title,
			Title:    opData.ClientData.Title,
			PageNo:   1,
			PageSize: 20,
			Total:    0,
			Operations: map[cptype.OperationKey]cptype.Operation{
				kanban.OpBoardLoadMore{}.OpKey(): cputil.NewOpBuilder().Build(),
				kanban.OpBoardUpdate{}.OpKey():   cputil.NewOpBuilder().Build(),
				kanban.OpBoardDelete{}.OpKey():   cputil.NewOpBuilder().Build(),
			},
		})
	}
}

// RegisterBoardLoadMoreOp only return specific board data.
func (k *Kanban) RegisterBoardLoadMoreOp(opData kanban.OpBoardLoadMore) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		if opData.ClientData.PageNo > 0 {
			k.filterReq.PageNo = opData.ClientData.PageNo
		}
		if opData.ClientData.PageSize > 0 {
			k.filterReq.PageSize = opData.ClientData.PageSize
		}
		k.StdDataPtr = k.doFilter(opData.ClientData.DataRef.ID)
	}
}

// RegisterBoardUpdateOp no need here.
func (k *Kanban) RegisterBoardUpdateOp(opData kanban.OpBoardUpdate) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		// mock
		logrus.Infof("mock update board: %s, fromTitle: %s, newTitle: %s", opData.ClientData.DataRef.ID, opData.ClientData.DataRef.Title, opData.ClientData.Title)
		k.StdDataPtr = k.doFilter()
		for i, board := range k.StdDataPtr.Boards {
			if board.ID == opData.ClientData.DataRef.ID {
				k.StdDataPtr.Boards[i].Title = opData.ClientData.Title
			}
		}
	}
}

// RegisterBoardDeleteOp no need here.
func (k *Kanban) RegisterBoardDeleteOp(opData kanban.OpBoardDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		logrus.Infof("mock delete board: %s", opData.ClientData.DataRef.ID)
		k.StdDataPtr = k.doFilter()
		var newBoards []kanban.Board
		for _, board := range k.StdDataPtr.Boards {
			if board.ID != opData.ClientData.DataRef.ID {
				newBoards = append(newBoards, board)
			}
		}
		k.StdDataPtr.Boards = newBoards
	}
}

func (k *Kanban) RegisterCardMoveToOp(opData kanban.OpCardMoveTo) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		issueID, err := strconv.ParseUint(opData.ClientData.DataRef.ID, 10, 64)
		if err != nil {
			panic(fmt.Errorf("invalid card issueID: %s, err: %v", opData.ClientData.DataRef.ID, err))
		}
		targetStateID, err := strconv.ParseInt(opData.ClientData.TargetBoardID, 10, 64)
		if err != nil {
			panic(fmt.Errorf("invalid state id: %s, err: %v", opData.ClientData.TargetBoardID, err))
		}
		if err := k.issueSvc.UpdateIssue(apistructs.IssueUpdateRequest{
			ID:    issueID,
			State: &targetStateID,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		}); err != nil {
			panic(fmt.Errorf("failed to update issue: %v", err))
		}
		k.StdDataPtr = k.doFilter()
	}
}
