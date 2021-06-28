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

	CommonFilter
}

type CommonFilter struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type Operations struct {
	Reload bool   `json:"reload"`
	Key    string `json:"key"`
}

type Options struct {
	Key        string `json:"key"`
	Text       string `json:"text"`
	Operations map[string]interface{}
}

type Props struct {
	Delay   uint64 `json:"delay"`
	Visible bool   `json:"visible"`
}

type StateConditions struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	EmptyText   string `json:"emptyText"`
	Fixed       bool   `json:"fixed"`
	ShowIndex   uint64 `json:"showIndex"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type State struct {
	Values        map[string]interface{} `json:"values"`
	Conditions    []StateConditions      `json:"conditions"`
	IsFirstFilter bool                   `json:"isFirstFilter"`
	IsEmpty       bool                   `json:"isEmpty"`
}
