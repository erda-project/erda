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

package eventTable

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentEventTable struct {
	base.DefaultProvider
	sdk *cptype.SDK
	bdl *bundle.Bundle

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	Data       Data                   `json:"data,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName        string       `json:"clusterName,omitempty"`
	FilterValues       FilterValues `json:"filterValues,omitempty"`
	PageNo             uint64       `json:"pageNo,omitempty"`
	PageSize           uint64       `json:"pageSize,omitempty"`
	Sorter             Sorter       `json:"sorterData,omitempty"`
	Total              uint64       `json:"total"`
	EventTableUQLQuery string       `json:"eventTable__urlQuery,omitempty"`
}

type FilterValues struct {
	Namespace []string `json:"namespace,omitempty"`
	Type      []string `json:"type,omitempty"`
	Search    string   `json:"search,omitempty"`
}

type Data struct {
	List []Item `json:"list"`
}

type Item struct {
	LastSeen          string `json:"lastSeen,omitempty"`
	LastSeenTimestamp int64  `json:"lastSeenTimestamp,omitempty"`
	Type              string `json:"type,omitempty"`
	Reason            string `json:"reason,omitempty"`
	Object            string `json:"object,omitempty"`
	Source            string `json:"source,omitempty"`
	Message           string `json:"message,omitempty"`
	Count             string `json:"count,omitempty"`
	CountNum          int64  `json:"countNum"`
	Name              string `json:"name,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
}

type Props struct {
	PageSizeOptions []string `json:"pageSizeOptions,omitempty"`
	Columns         []Column `json:"columns,omitempty"`
}

type Column struct {
	DataIndex string `json:"dataIndex,omitempty"`
	Title     string `json:"title,omitempty"`
	Width     int    `json:"width,omitempty"`
	Sorter    bool   `json:"sorter,omitempty"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}

type Sorter struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}
