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

package fileFormModal

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFileFormModal struct {
	CtxBdl     protocol.ContextBundle
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type State struct {
	FormVisible bool     `json:"formVisible"`
	SceneSetKey int      `json:"sceneSetKey"`
	Visible     bool     `json:"visible"`
	FormData    FormData `json:"formData"`
	ActionType  string   `json:"actionType"`
	SceneId     uint64   `json:"sceneId"`
}

type FormData struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	// SetID       uint64 `json:"setId"`
	ScenesSet *uint64 `json:"scenesSet"`
}

type Props struct {
	Title  string  `json:"title"`
	Fields []Entry `json:"fields"`
}

type Entry struct {
	Key            string         `json:"key"`
	Label          string         `json:"label"`
	Required       bool           `json:"required"`
	Component      string         `json:"component"`
	Rules          []Rule         `json:"rules"`
	ComponentProps ComponentProps `json:"componentProps"`
}

type Rule struct {
	Pattern string `json:"pattern"`
	Msg     string `json:"msg"`
}
type ComponentProps struct {
	MaxLength   int           `json:"maxLength"`
	Placeholder string        `json:"placeholder"`
	Options     []interface{} `json:"options"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}
