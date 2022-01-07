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

package inputFilter

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

type ComponentInputFilter struct {
	sdk *cptype.SDK

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	Conditions          []Condition `json:"conditions,omitempty"`
	Values              Values      `json:"values"`
	InputFilterURLQuery string      `json:"inputFilter__urlQuery,omitempty"`
}

type Values struct {
	Version string `json:"version,omitempty"`
}

type Condition struct {
	EmptyText   string `json:"emptyText,omitempty"`
	Fixed       bool   `json:"fixed"`
	Key         string `json:"key,omitempty"`
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Type        string `json:"type,omitempty"`
}

type Show struct {
	Show bool `json:"show"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}
