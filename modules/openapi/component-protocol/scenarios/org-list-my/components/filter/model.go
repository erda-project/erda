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

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFilter struct {
	CtxBdl     protocol.ContextBundle
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
}

type Props struct {
	Delay int `json:"delay"`
}

type State struct {
	Conditions    []Condition       `json:"conditions"`
	Values        map[string]string `json:"values"`
	SearchEntry   string            `json:"searchEntry"`
	SearchRefresh bool              `json:"searchRefresh"`
}

type Condition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	EmptyText   string `json:"emptyText"`
	Fixed       bool   `json:"fixed"`
	ShowIndex   int    `json:"showIndex"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}
