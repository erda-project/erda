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
