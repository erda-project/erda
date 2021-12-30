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

package releaseTable

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentReleaseTable struct {
	base.DefaultProvider
	sdk *cptype.SDK
	bdl *bundle.Bundle

	Type       string                 `json:"type,omitempty"`
	Data       Data                   `json:"data"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type Data struct {
	List []Item `json:"list,omitempty"`
}

type Item struct {
	ID              string          `json:"id,omitempty"`
	Version         string          `json:"version,omitempty"`
	Application     string          `json:"application,omitempty"`
	Desc            string          `json:"desc,omitempty"`
	Creator         Creator         `json:"creator,omitempty"`
	CreatedAt       string          `json:"createdAt,omitempty"`
	Operations      TableOperations `json:"operations"`
	BatchOperations []string        `json:"batchOperations,omitempty"`
}

type Creator struct {
	RenderType string   `json:"renderType,omitempty"`
	Value      []string `json:"value,omitempty"`
}

type TableOperations struct {
	Operations map[string]interface{} `json:"operations,omitempty"`
	RenderType string                 `json:"renderType,omitempty"`
}

type Operation struct {
	Command     Command                `json:"command"`
	Confirm     string                 `json:"confirm,omitempty"`
	Key         string                 `json:"key,omitempty"`
	Reload      bool                   `json:"reload"`
	Text        string                 `json:"text,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	SuccessMsg  string                 `json:"successMsg,omitempty"`
	Disabled    bool                   `json:"disabled,omitempty"`
	DisabledTip string                 `json:"disabledTip,omitempty"`
}

type Command struct {
	JumpOut bool   `json:"jumpOut,omitempty"`
	Key     string `json:"key,omitempty"`
	Target  string `json:"target,omitempty"`
}

type Props struct {
	BatchOperations []string `json:"batchOperations,omitempty"`
	Selectable      bool     `json:"selectable"`
	Columns         []Column `json:"columns,omitempty"`
	PageSizeOptions []string `json:"PageSizeOptions,omitempty"`
	RowKey          string   `json:"rowKey,omitempty"`
}

type Column struct {
	DataIndex string `json:"dataIndex,omitempty"`
	Title     string `json:"title,omitempty"`
	Sorter    bool   `json:"sorter"`
	Align     string `json:"align,omitempty"`
}

type State struct {
	ReleaseTableURLQuery string        `json:"releaseTable__urlQuery"`
	PageNo               int64         `json:"pageNo"`
	PageSize             int64         `json:"pageSize"`
	Total                int64         `json:"total"`
	SelectedRowKeys      []string      `json:"selectedRowKeys,omitempty"`
	Sorter               Sorter        `json:"sorterData"`
	IsProjectRelease     bool          `json:"isProjectRelease"`
	ProjectID            int64         `json:"projectID"`
	IsFormal             bool          `json:"isFormal"`
	VersionValues        VersionValues `json:"versionValues"`
	FilterValues         FilterValues  `json:"filterValues"`
	ApplicationID        int64         `json:"applicationID"`
}

type VersionValues struct {
	Version string `json:"version,omitempty"`
}

type FilterValues struct {
	BranchID          string   `json:"branchID,omitempty"`
	CommitID          string   `json:"commitID,omitempty"`
	UserIDs           []string `json:"userIDs,omitempty"`
	CreatedAtStartEnd []int64  `json:"createdAtStartEnd,omitempty"`
}

type Sorter struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}
