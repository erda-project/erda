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

package page

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentPage struct {
	CtxBdl protocol.ContextBundle
	State  State `json:"state"`
	Props  Props `json:"props"`
}

type State struct {
	ActiveKey string `json:"activeKey"`
}

type Props struct {
	TabMenu []Option `json:"tabMenu"`
}

type Option struct {
	Key        string                 `json:"key"`
	Name       string                 `json:"name"`
	Operations map[string]interface{} `json:"operations"`
}

type ClickOperation struct {
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Command Command `json:"command"`
}

type Command struct {
	Key          string `json:"key"`
	ScenarioType string `json:"scenarioType"`
	ScenarioKey  string `json:"scenarioKey"`
}
