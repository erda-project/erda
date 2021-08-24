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

package filter

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type ComponentFilter struct {
	ctxBdl protocol.ContextBundle

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName string      `json:"clusterName,omitempty"`
	Conditions  []Condition `json:"conditions,omitempty"`
	Values      Values      `json:"values,omitempty"`
}

type Values struct {
	Type      []string `json:"type,omitempty"`
	Namespace []string `json:"namespace,omitempty"`
}

type Condition struct {
	Key     string   `json:"key,omitempty"`
	Label   string   `json:"label,omitempty"`
	Type    string   `json:"type,omitempty"`
	Fixed   bool     `json:"fixed"`
	Options []Option `json:"options,omitempty"`
}

type Option struct {
	Label    string   `json:"label,omitempty"`
	Value    string   `json:"value,omitempty"`
	Children []Option `json:"children,omitempty"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}
