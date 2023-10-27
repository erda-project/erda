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

package issueTable

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"

	"github.com/erda-project/erda/apistructs"
)

var (
	priorityIcon = map[apistructs.IssuePriority]string{
		apistructs.IssuePriorityUrgent: "ISSUE_ICON.priority.URGENT", // 紧急
		apistructs.IssuePriorityHigh:   "ISSUE_ICON.priority.HIGH",   // 高
		apistructs.IssuePriorityNormal: "ISSUE_ICON.priority.NORMAL", // 中
		apistructs.IssuePriorityLow:    "ISSUE_ICON.priority.LOW",    // 低
	}

	stateIcon = map[string]string{
		string(apistructs.IssueStateOpen):     "ISSUE_ICON.state.OPEN",
		string(apistructs.IssueStateWorking):  "ISSUE_ICON.state.WORKING",
		string(apistructs.IssueStateTesting):  "ISSUE_ICON.state.TESTING",
		string(apistructs.IssueStateDone):     "ISSUE_ICON.state.DONE",
		string(apistructs.IssueStateResolved): "ISSUE_ICON.state.RESOLVED",
		string(apistructs.IssueStateReopen):   "ISSUE_ICON.state.REOPEN",
		string(apistructs.IssueStateWontfix):  "ISSUE_ICON.state.WONTFIX",
		string(apistructs.IssueStateDup):      "ISSUE_ICON.state.DUP",
		string(apistructs.IssueStateClosed):   "ISSUE_ICON.state.CLOSED",
	}
)

type ComponentAction struct {
	labels   []apistructs.ProjectLabel
	isGuest  bool
	userMap  map[string]string
	state    State
	issueSvc pb.IssueCoreServiceServer
	ctx      context.Context
}

type ProgressBlock struct {
	Value      string `json:"value"`
	RenderType string `json:"renderType"`
	HiddenText bool   `json:"hiddenText"`
}

type Severity struct {
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Operations  map[string]interface{} `json:"operations"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type TableColumnMultiple struct {
	RenderType string        `json:"renderType,omitempty"`
	Direction  string        `json:"direction,omitempty"`
	Renders    []interface{} `json:"renders,omitempty"`
}

type TableColumnTextWithIcon struct {
	RenderType string `json:"renderType,omitempty"`
	Value      string `json:"value,omitempty"`
	PrefixIcon string `json:"prefixIcon,omitempty"`
}

type TableColumnTagsRow struct {
	RenderType string                  `json:"renderType,omitempty"`
	Value      []TableColumnTagsRowTag `json:"value,omitempty"`
	ShowCount  int                     `json:"showCount,omitempty"`
}

type TableColumnTagsRowTag struct {
	Color string `json:"color,omitempty"`
	Label string `json:"label,omitempty"`
}

type State struct {
	Menus []map[string]interface{} `json:"menus"`
	// Operations  map[string]interface{} `json:"operations"`
	// PrefixIcon  string                 `json:"prefixIcon"`
	Value           string   `json:"value"`
	RenderType      string   `json:"renderType"`
	Disabled        bool     `json:"disabled"`
	DisabledTip     string   `json:"disabledTip"`
	SelectedRowKeys []string `json:"selectedRowKeys"`
	ProjectID       uint64   `json:"projectId"`
}

type Priority struct {
	Operations  map[string]interface{} `json:"operations"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type Deadline struct {
	RenderType     string                 `json:"renderType"`
	Value          string                 `json:"value"`
	NoBorder       bool                   `json:"noBorder"`
	DisabledBefore string                 `json:"disabledBefore"`
	DisabledAfter  string                 `json:"disabledAfter"`
	Operations     map[string]interface{} `json:"operations"`
}

type Assignee struct {
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	Scope       string                 `json:"scope"`
	Operations  map[string]interface{} `json:"operations"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type TableItem struct {
	//Assignee    map[string]string `json:"assignee"`
	Id          string     `json:"id"`
	IterationID int64      `json:"iterationID"`
	Priority    Priority   `json:"priority"`
	Progress    Progress   `json:"progress,omitempty"`
	Severity    Severity   `json:"severity,omitempty"`
	Complexity  Complexity `json:"complexity,omitempty"`
	State       State      `json:"state"`
	// Title       Title      `json:"title"`
	Type            string    `json:"type"`
	Deadline        Deadline  `json:"deadline"`
	Assignee        Assignee  `json:"assignee"`
	ClosedAt        Time      `json:"closedAt"`
	Name            Name      `json:"name"`
	ReopenCount     TextBlock `json:"reopenCount,omitempty"`
	CreatedAt       Time      `json:"createdAt"`
	Owner           Assignee  `json:"owner"`
	Creator         Assignee  `json:"creator"`
	PlanStartedAt   Time      `json:"planStartedAt"`
	Iteration       TextBlock `json:"iteration"`
	BatchOperations []string  `json:"batchOperations"`

	Properties []*pb.IssuePropertyExtraProperty `json:"properties"`
}

type TextBlock struct {
	Value      string `json:"value"`
	RenderType string `json:"renderType"`
}

type Name struct {
	RenderType   string       `json:"renderType"`
	PrefixIcon   string       `json:"prefixIcon"`
	Value        string       `json:"value"`
	ExtraContent ExtraContent `json:"extraContent"`
}

type ExtraContent struct {
	RenderType string  `json:"renderType"`
	Value      []Label `json:"value"`
	ShowCount  int     `json:"showCount,omitempty"`
}

type Label struct {
	Color string `json:"color"`
	Label string `json:"label"`
}

type Complexity struct {
	RenderType string `json:"renderType"`
	PrefixIcon string `json:"prefixIcon"`
	Value      string `json:"value"`
}

type Time struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	NoBorder   bool   `json:"noBorder"`
}

type PriorityOperationData struct {
	Meta struct {
		Priority string `json:"priority"`
		ID       string `json:"id"`
	} `json:"meta"`
}
type DeadlineOperationData struct {
	Meta struct {
		DeadlineValue string `json:"deadlineValue"`
		ID            string `json:"id"`
	} `json:"meta"`
}
type StateOperationData struct {
	Meta struct {
		State string `json:"state"`
		ID    string `json:"id"`
	} `json:"meta"`
}
type SeverityOperationData struct {
	Meta struct {
		Severity string `json:"severity"`
		ID       string `json:"id"`
	} `json:"meta"`
}
type AssigneeOperationData struct {
	Meta struct {
		Assignee string `json:"assignee"`
		ID       string `json:"id"`
	} `json:"meta"`
}

type BatchState struct {
	Visible         bool     `json:"visible"`
	SelectedRowKeys []string `json:"selectedRowKeys"`
}

type Operation struct {
	Key        string `json:"key"`
	Reload     bool   `json:"reload"`
	Meta       Meta   `json:"meta"`
	SuccessMsg string `json:"successMsg"`
}

type Props struct {
	Status  string `json:"status"`
	Content string `json:"content"`
	Title   string `json:"title"`
}

type Meta struct {
	Type cptype.OperationKey `json:"type"`
}
