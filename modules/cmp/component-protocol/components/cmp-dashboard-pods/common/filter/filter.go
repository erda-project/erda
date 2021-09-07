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

type Filter struct {
	base.DefaultProvider
	CtxBdl     *bundle.Bundle
	SDK        *cptype.SDK
	Type       string                     `json:"type"`
	Operations map[string]FilterOperation `json:"operations"`
	State      State                      `json:"state"`
	Props      Props                      `json:"props"`
}

type State struct {
	Values Values `json:"values"`
}

type Props struct {
	LabelWidth int     `json:"label_width"`
	Fields     []Field `json:"fields"`
}

type FilterOperation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type Values struct {
	Keys map[string][]string `json:"keys,omitempty"`
}

type Field struct {
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Options     []Option `json:"options"`
	Key         string   `json:"key"`
	Placeholder string   `json:"placeholder"`
}

type Option struct {
	Label    string   `json:"label"`
	Value    string   `json:"value"`
	Children []Option `json:"children"`
}
