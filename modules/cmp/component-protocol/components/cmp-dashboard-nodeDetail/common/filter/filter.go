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

import "github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

type CommonFilter struct {
	base.DefaultProvider
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
