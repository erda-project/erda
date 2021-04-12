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

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentEmptyText struct {
	CtxBdl     protocol.ContextBundle
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
}

type Props struct {
	Visible     bool                     `json:"visible"`
	RenderType  string                   `json:"renderType"`
	StyleConfig Config                   `json:"styleConfig"`
	Value       map[string][]interface{} `json:"value"`
}

type State struct {
	IsEmpty bool `json:"isEmpty"`
}

type Config struct {
	FontSize   int `json:"fontSize"`
	LineHeight int `json:"lineHeight"`
}

type Redirect struct {
	Text         string         `json:"text"`
	OperationKey string         `json:"operationKey"`
	StyleConfig  RedirectConfig `json:"styleConfig"`
}

type RedirectConfig struct {
	Bold bool `json:"bold"`
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
