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

package emptyText

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type ComponentText struct {
	ctxBdl protocol.ContextBundle

	CommonText
}

type CommonText struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type Operation struct {
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Command Command `json:"command"`
}

type Command struct {
	Key          string `json:"key"`
	ScenarioType string `json:"scenarioType"`
	ScenarioKey  string `json:"scenarioKey"`
}

type StyleConfig struct {
	FontSize   uint64 `json:"fontSize"`
	LineHeight uint64 `json:"lineHeight"`
}

type Props struct {
	Visible     bool                   `json:"visible"`
	RenderType  string                 `json:"renderType"`
	StyleConfig StyleConfig            `json:"styleConfig"`
	Value       map[string]interface{} `json:"value"`
}

type State struct {
	IsEmpty bool `json:"isEmpty"`
}
