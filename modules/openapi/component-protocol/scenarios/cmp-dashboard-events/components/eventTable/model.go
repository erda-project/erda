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

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type ComponentEventTable struct {
	ctxBdl protocol.ContextBundle

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	Data       Data                   `json:"data,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	PageNo       uint64       `json:"pageNo,omitempty"`
	PageSize     uint64       `json:"pageSize,omitempty"`
	Total        uint64       `json:"total"`
	Sorter       Sorter       `json:"sorterData,omitempty"`
	ClusterName  string       `json:"clusterName,omitempty"`
	FilterValues FilterValues `json:"filterValues,omitempty"`
}

type FilterValues struct {
	Namespace []string `json:"namespace,omitempty"`
	Type      []string `json:"type,omitempty"`
}

type Data struct {
	List []Item `json:"list"`
}

type Item struct {
	LastSeen          string `json:"lastSeen"`
	LastSeenTimestamp int64  `json:"lastSeenTimestamp"`
	Type              string `json:"type"`
	Reason            string `json:"reason"`
	Object            string `json:"object"`
	Source            string `json:"source"`
	Message           string `json:"message"`
	Count             string `json:"count"`
	CountNum          int64  `json:"countNum"`
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
}

type Props struct {
	PageSizeOptions []string `json:"pageSizeOptions"`
	Columns         []Column `json:"columns"`
}

type Column struct {
	DataIndex string `json:"dataIndex"`
	Title     string `json:"title"`
	Width     string `json:"width"`
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
