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

package releaseFilter

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
)

type ComponentReleaseFilter struct {
	sdk *cptype.SDK
	bdl *bundle.Bundle

	Type  string `json:"type,omitempty"`
	State State  `json:"state"`
	Data  Data   `json:"data"`
}

type State struct {
	Values                Values `json:"values"`
	ReleaseFilterURLQuery string `json:"releaseFilter__urlQuery,omitempty"`
	IsProjectRelease      bool   `json:"isProjectRelease"`
	ProjectID             int64  `json:"projectID"`
}

type Values struct {
	ApplicationIDs    []string `json:"applicationIDs,omitempty"`
	BranchID          string   `json:"branchID,omitempty"`
	CommitID          string   `json:"commitID,omitempty"`
	CreatedAtStartEnd []int64  `json:"createdAtStartEnd,omitempty"`
	UserIDs           []string `json:"userIDs,omitempty"`
}

type Data struct {
	HideSave   bool        `json:"hideSave"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Key         string   `json:"key,omitempty"`
	Label       string   `json:"label,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Type        string   `json:"type,omitempty"`
	Options     []Option `json:"options,omitempty"`
}

type Option struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
}
