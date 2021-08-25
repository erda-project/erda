// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package browsePublicProjects

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type ComponentBrowsePublic struct {
	ctxBdl     protocol.ContextBundle
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
	Show    bool    `json:"show"`
}

type Command struct {
	Key     string `json:"key"`
	Target  string `json:"target"`
	JumpOut bool   `json:"jumpOut"`
	Visible bool   `json:"visible"`
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
