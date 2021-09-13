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

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentFilter struct {
	base.DefaultProvider
	bdl *bundle.Bundle
	sdk *cptype.SDK

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName    string      `json:"clusterName,omitempty"`
	Conditions     []Condition `json:"conditions,omitempty"`
	Values         Values      `json:"values,omitempty"`
	FilterURLQuery string      `json:"filter__urlQuery,omitempty"`
}

type Condition struct {
	HaveFilter  bool     `json:"haveFilter,omitempty"`
	Key         string   `json:"key,omitempty"`
	Label       string   `json:"label,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Type        string   `json:"type,omitempty"`
	Fixed       bool     `json:"fixed,omitempty"`
	Options     []Option `json:"options,omitempty"`
}

type Option struct {
	Label    string   `json:"label,omitempty"`
	Value    string   `json:"value,omitempty"`
	Children []Option `json:"children,omitempty"`
}

type Values struct {
	Kind      []string `json:"kind,omitempty"`
	Namespace []string `json:"namespace,omitempty"`
	Status    []string `json:"status,omitempty"`
	Node      []string `json:"node,omitempty"`
	Search    string   `json:"search,omitempty"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}
