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

package list

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentList struct {
	CtxBdl     protocol.ContextBundle
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
	Data       []Org                  `json:"data"`
}

type State struct {
	PageNo        int    `json:"pageNo"`
	PageSize      int    `json:"pageSize"`
	Total         int    `json:"total"`
	SearchEntry   string `json:"searchEntry"`
	SearchRefresh bool   `json:"searchRefresh"`
	IsEmpty       bool   `json:"isEmpty"`
}

type Org struct {
	Id          uint64                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	PrefixImg   string                 `json:"prefixImg"`
	ExtraInfos  []ExtraInfo            `json:"extraInfos"`
	Operations  map[string]interface{} `json:"operations"`
}

type ExtraInfo struct {
	Icon    string `json:"icon"`
	Text    string `json:"text"`
	ToolTip string `json:"tooltip"`
}

type Option struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type OperationBase struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type ClickOperation struct {
	Key     string  `json:"key"`
	Show    bool    `json:"show"`
	Reload  bool    `json:"reload"`
	Command Command `json:"command"`
}

type ManageOperation struct {
	Key     string  `json:"key"`
	Text    string  `json:"text"`
	Reload  bool    `json:"reload"`
	Command Command `json:"command"`
}

type ExitOperation struct {
	Key     string `json:"key"`
	Text    string `json:"text"`
	Reload  bool   `json:"reload"`
	Confirm string `json:"confirm"`
	Meta    Meta   `json:"meta"`
}

type Meta struct {
	Id uint64 `json:"id"`
}

type Command struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

type Props struct {
	PageSizeOptions []string `json:"pageSizeOptions"`
	Visible         bool     `json:"visible"`
}
