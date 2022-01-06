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

package filter

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type ComponentFilter struct {
	sdk      *cptype.SDK
	bdl      *bundle.Bundle
	issueSvc *issue.Issue
	filter.CommonFilter
	State State `json:"state,omitempty"`

	FrontendUrlQuery string              `json:"-"`
	projectID        uint64              `json:"-"`
	Members          []apistructs.Member `json:"-"`

	fixedIterationID uint64 `json:"-"`
}

type State struct {
	Base64UrlQueryParams string                 `json:"filter__urlQuery,omitempty"`
	Conditions           []filter.PropCondition `json:"conditions,omitempty"`
	Values               FrontendConditions     `json:"values,omitempty"`
}

type FrontendConditions struct {
	IterationIDs []int64  `json:"iteration,omitempty"`
	AssigneeIDs  []string `json:"member,omitempty"`
	LabelIDs     []uint64 `json:"label,omitempty"`
}

const OperationKeyFilter filter.OperationKey = "filter"
