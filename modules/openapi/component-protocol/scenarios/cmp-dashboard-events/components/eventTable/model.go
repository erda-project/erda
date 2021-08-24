// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
