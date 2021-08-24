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

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName string      `json:"clusterName,omitempty"`
	Conditions  []Condition `json:"conditions,omitempty"`
	Values      Values      `json:"values,omitempty"`
}

type Values struct {
	Type      []string `json:"type,omitempty"`
	Namespace []string `json:"namespace,omitempty"`
}

type Condition struct {
	Key     string   `json:"key,omitempty"`
	Label   string   `json:"label,omitempty"`
	Type    string   `json:"type,omitempty"`
	Fixed   bool     `json:"fixed"`
	Options []Option `json:"options,omitempty"`
}

type Option struct {
	Label    string   `json:"label,omitempty"`
	Value    string   `json:"value,omitempty"`
	Children []Option `json:"children,omitempty"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}
