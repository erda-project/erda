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

type CommonFilter struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type"`
	Props      Props                  `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type Operation struct {
	Reload bool   `json:"reload"`
	Key    string `json:"key"`
}

type Options struct {
	Label    string    `json:"label"`
	Value    string    `json:"value"`
	Children []Options `json:"children"`
}

type Props struct {
	Delay uint64 `json:"delay"`
}

type StateCondition struct {
	Key         string    `json:"key"`
	Label       string    `json:"label"`
	EmptyText   string    `json:"emptyText"`
	Fixed       bool      `json:"fixed"`
	ShowIndex   uint64    `json:"showIndex"`
	Placeholder string    `json:"placeholder"`
	Type        string    `json:"type"`
	Options     []Options `json:"options"`
}

type State struct {
	Values map[string]interface{} `json:"values"`
	// 0: input 1: select
	Conditions    []StateCondition `json:"conditions"`
	IsFirstFilter bool             `json:"isFirstFilter"`
}
